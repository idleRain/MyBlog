// 演示数据种子：生成可浏览的分类、标签、文章与评论样例。
package database

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"MyBlog/internal/model"
	"MyBlog/internal/service"

	"gorm.io/gorm"
)

// productionServerMode 生产运行模式标识，取值与 config.yaml 的 server.mode 一致。
const productionServerMode = "release"

// demoDataset 演示数据集定义，与数据库写入解耦以便单测校验其自洽性。
type demoDataset struct {
	Categories []demoCategory
	Tags       []demoTag
	Articles   []demoArticle
	Comments   []demoComment
}

// demoCategory 演示分类，Slug 为幂等键。
type demoCategory struct {
	Slug        string
	Name        string
	Description string
	SortOrder   int
	IsFeatured  bool
}

// demoTag 演示标签，Name 为幂等键。
// Slug 必须显式给出：仓储层在 slug 为空时才生成，种子数据绕过仓储直接写库，
// 空 slug 会撞上 tags.slug 的唯一索引，第二个标签即写入失败。
type demoTag struct {
	Name  string
	Slug  string
	Color string
	IsHot bool
}

// demoArticle 演示文章，Slug 为幂等键。
// PublishedMonthsAgo 把发布时间铺开到多个月份，使归档页在演示数据下呈现跨月分组。
type demoArticle struct {
	Slug               string
	Title              string
	Summary            string
	Content            string
	CategorySlug       string
	TagNames           []string
	Status             model.ArticleStatus
	PublishedMonthsAgo int
	IsFeatured         bool
	IsTop              bool
	CommentEnabled     bool
	ViewCount          uint
	LikeCount          uint
	BookmarkCount      uint
}

// demoComment 演示评论。ReplyTo 为空表示根评论，否则取值为同文章内某个根评论的作者名。
type demoComment struct {
	ArticleSlug string
	AuthorName  string
	Content     string
	Status      model.CommentStatus
	IsPinned    bool
	ReplyTo     string
}

// DemoSeedOptions 演示数据写入参数。
type DemoSeedOptions struct {
	AuthorID   uint   // 演示文章作者，取已存在的管理员账号
	ServerMode string // 当前运行模式，release 下默认拒绝写入
	// AllowProduction 显式放行生产模式下的演示数据写入，仅在操作者确认后开启。
	AllowProduction bool
}

// SeedDemoData 幂等写入演示数据。
// 幂等键为分类与文章 slug、标签名称：已存在的记录按幂等键复用，重复执行不产生重复内容。
// release 模式下默认拒绝写入，避免演示内容进入生产站点。
func SeedDemoData(db *gorm.DB, opts DemoSeedOptions) error {
	if opts.ServerMode == productionServerMode && !opts.AllowProduction {
		return fmt.Errorf("运行模式为 %s，拒绝写入演示数据；确认无误后显式加 --force 放行", productionServerMode)
	}
	if opts.AuthorID == 0 {
		return errors.New("演示数据缺少作者，请先初始化管理员账号")
	}

	dataset := buildDemoDataset()
	if err := validateDemoDataset(dataset); err != nil {
		return fmt.Errorf("演示数据集不自洽: %w", err)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		categoryIDs, err := ensureDemoCategories(tx, dataset.Categories)
		if err != nil {
			return err
		}
		tagIDs, err := ensureDemoTags(tx, dataset.Tags)
		if err != nil {
			return err
		}
		articleIDs, err := ensureDemoArticles(tx, dataset.Articles, opts.AuthorID, categoryIDs, tagIDs)
		if err != nil {
			return err
		}
		if err := ensureDemoComments(tx, dataset.Comments, articleIDs); err != nil {
			return err
		}
		return syncDemoCounters(tx, dataset)
	})
	if err != nil {
		return err
	}

	log.Printf("演示数据写入完成：%s", describeDemoDataset(dataset))
	return nil
}

// ensureDemoCategories 按 slug 幂等创建演示分类，返回 slug 到分类 ID 的映射。
func ensureDemoCategories(tx *gorm.DB, categories []demoCategory) (map[string]uint, error) {
	ids := make(map[string]uint, len(categories))

	for _, item := range categories {
		category := model.Category{Slug: item.Slug}
		// 已存在的分类沿用原记录，不覆盖站点上既有内容。
		err := tx.Where(model.Category{Slug: item.Slug}).
			Attrs(model.Category{
				Name:        item.Name,
				Description: item.Description,
				SortOrder:   item.SortOrder,
				Status:      model.CategoryStatusShown,
				IsFeatured:  item.IsFeatured,
				Level:       1,
			}).
			FirstOrCreate(&category).Error
		if err != nil {
			return nil, fmt.Errorf("写入演示分类 %s 失败: %w", item.Slug, err)
		}
		if err := fillDemoCategoryPath(tx, category); err != nil {
			return nil, err
		}
		ids[item.Slug] = category.ID
	}

	return ids, nil
}

// fillDemoCategoryPath 回填顶级分类的物化路径，与服务层顶级分类的 /ID 口径一致。
func fillDemoCategoryPath(tx *gorm.DB, category model.Category) error {
	if category.Path != "" {
		return nil
	}

	path := fmt.Sprintf("/%d", category.ID)
	err := tx.Model(&model.Category{}).Where("id = ?", category.ID).UpdateColumn("path", path).Error
	if err != nil {
		return fmt.Errorf("回填演示分类 %s 的物化路径失败: %w", category.Slug, err)
	}
	return nil
}

// ensureDemoTags 按名称幂等创建演示标签，返回名称到标签 ID 的映射。
func ensureDemoTags(tx *gorm.DB, tags []demoTag) (map[string]uint, error) {
	ids := make(map[string]uint, len(tags))

	for _, item := range tags {
		tag := model.Tag{Name: item.Name}
		err := tx.Where(model.Tag{Name: item.Name}).
			Attrs(model.Tag{
				Slug:   item.Slug,
				Color:  item.Color,
				Status: model.TagStatusEnabled,
				IsHot:  item.IsHot,
			}).
			FirstOrCreate(&tag).Error
		if err != nil {
			return nil, fmt.Errorf("写入演示标签 %s 失败: %w", item.Name, err)
		}
		ids[item.Name] = tag.ID
	}

	return ids, nil
}

// ensureDemoArticles 按 slug 幂等创建演示文章并维护分类与标签关联，返回 slug 到文章 ID 的映射。
func ensureDemoArticles(
	tx *gorm.DB,
	articles []demoArticle,
	authorID uint,
	categoryIDs map[string]uint,
	tagIDs map[string]uint,
) (map[string]uint, error) {
	ids := make(map[string]uint, len(articles))

	for _, item := range articles {
		existing, err := findDemoArticle(tx, item.Slug)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			ids[item.Slug] = existing.ID
			continue
		}

		article, err := buildDemoArticle(item, authorID, categoryIDs, time.Now())
		if err != nil {
			return nil, err
		}
		if err := tx.Create(article).Error; err != nil {
			return nil, fmt.Errorf("写入演示文章 %s 失败: %w", item.Slug, err)
		}
		if err := mountDemoArticle(tx, article, item, categoryIDs, tagIDs); err != nil {
			return nil, err
		}
		ids[item.Slug] = article.ID
	}

	return ids, nil
}

// buildDemoArticle 由数据集条目构造文章实体，派生字段经服务层的唯一实现计算。
func buildDemoArticle(
	item demoArticle,
	authorID uint,
	categoryIDs map[string]uint,
	now time.Time,
) (*model.Article, error) {
	categoryID := categoryIDs[item.CategorySlug]

	article := &model.Article{
		Title:          item.Title,
		Slug:           item.Slug,
		Summary:        item.Summary,
		Content:        item.Content,
		AuthorID:       authorID,
		CategoryID:     &categoryID,
		Status:         item.Status,
		OriginType:     model.ArticleOriginOriginal,
		IsFeatured:     item.IsFeatured,
		IsTop:          item.IsTop,
		CommentEnabled: item.CommentEnabled,
		ViewCount:      item.ViewCount,
		LikeCount:      item.LikeCount,
		BookmarkCount:  item.BookmarkCount,
		SEODescription: item.Summary,
	}

	// 草稿不带发布时间；已发布、私有与归档状态都写入发布时间，供列表与归档页按时间排序。
	if item.Status != model.ArticleStatusDraft {
		publishedAt := now.AddDate(0, -item.PublishedMonthsAgo, 0)
		article.PublishedAt = &publishedAt
	}

	if err := service.ApplyContentDerivedFields(article); err != nil {
		return nil, fmt.Errorf("渲染演示文章 %s 正文失败: %w", item.Slug, err)
	}
	return article, nil
}

// findDemoArticle 按 slug 查找已存在的演示文章，不存在时返回 nil。
func findDemoArticle(tx *gorm.DB, slug string) (*model.Article, error) {
	var article model.Article
	err := tx.Where(model.Article{Slug: slug}).First(&article).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询演示文章 %s 失败: %w", slug, err)
	}
	return &article, nil
}

// mountDemoArticle 写入文章的分类与标签关联，关联表是分类页与标签页的查询依据。
func mountDemoArticle(
	tx *gorm.DB,
	article *model.Article,
	item demoArticle,
	categoryIDs map[string]uint,
	tagIDs map[string]uint,
) error {
	categoryMount := model.ArticleCategory{ArticleID: article.ID, CategoryID: categoryIDs[item.CategorySlug]}
	if err := tx.Create(&categoryMount).Error; err != nil {
		return fmt.Errorf("挂载演示文章 %s 的分类失败: %w", item.Slug, err)
	}

	for _, tagName := range item.TagNames {
		tagMount := model.ArticleTag{ArticleID: article.ID, TagID: tagIDs[tagName]}
		if err := tx.Create(&tagMount).Error; err != nil {
			return fmt.Errorf("挂载演示文章 %s 的标签 %s 失败: %w", item.Slug, tagName, err)
		}
	}
	return nil
}

// ensureDemoComments 幂等写入演示评论：先写根评论，再按作者名解析回复的父评论。
// 幂等键为「文章 + 作者名」：评论没有业务唯一列，该组合由 validateDemoComments 保证唯一，
// 因此可以安全地作为重复执行的判定依据。
func ensureDemoComments(tx *gorm.DB, comments []demoComment, articleIDs map[string]uint) error {
	rootIDs := make(map[string]uint)
	replies := make([]demoComment, 0, len(comments))

	for _, item := range comments {
		if item.ReplyTo != "" {
			replies = append(replies, item)
			continue
		}

		commentID, err := ensureDemoRootComment(tx, item, articleIDs[item.ArticleSlug])
		if err != nil {
			return err
		}
		rootIDs[demoRootAuthorKey(item.ArticleSlug, item.AuthorName)] = commentID
	}

	for _, item := range replies {
		rootID := rootIDs[demoRootAuthorKey(item.ArticleSlug, item.ReplyTo)]
		if err := ensureDemoReply(tx, item, articleIDs[item.ArticleSlug], rootID); err != nil {
			return err
		}
	}

	return nil
}

// ensureDemoRootComment 按文章与作者名幂等写入根评论，返回评论 ID。
func ensureDemoRootComment(tx *gorm.DB, item demoComment, articleID uint) (uint, error) {
	comment := model.Comment{ArticleID: articleID, AuthorName: item.AuthorName}
	err := tx.Where(model.Comment{ArticleID: articleID, AuthorName: item.AuthorName}).
		Attrs(model.Comment{
			Level:    1,
			Content:  item.Content,
			Status:   item.Status,
			IsPinned: item.IsPinned,
		}).
		FirstOrCreate(&comment).Error
	if err != nil {
		return 0, fmt.Errorf("写入演示根评论失败: %w", err)
	}
	return comment.ID, nil
}

// ensureDemoReply 按文章与作者名幂等写入回复，并重算父评论的回复计数。
func ensureDemoReply(tx *gorm.DB, item demoComment, articleID, rootID uint) error {
	parentID := rootID
	comment := model.Comment{ArticleID: articleID, AuthorName: item.AuthorName}
	err := tx.Where(model.Comment{ArticleID: articleID, AuthorName: item.AuthorName}).
		Attrs(model.Comment{
			ParentID: &parentID,
			RootID:   &rootID,
			Level:    2,
			Content:  item.Content,
			Status:   item.Status,
		}).
		FirstOrCreate(&comment).Error
	if err != nil {
		return fmt.Errorf("写入演示回复失败: %w", err)
	}

	// 回复计数按关联回复实数重算，重复执行不会累加。
	// 计数与更新拆成两条语句：MySQL 不允许在 UPDATE 的子查询中引用同一张目标表。
	var replyCount int64
	err = tx.Model(&model.Comment{}).Where("parent_id = ?", rootID).Count(&replyCount).Error
	if err != nil {
		return fmt.Errorf("统计演示回复数失败: %w", err)
	}

	err = tx.Model(&model.Comment{}).Where("id = ?", rootID).
		UpdateColumn("reply_count", replyCount).Error
	if err != nil {
		return fmt.Errorf("重算演示回复计数失败: %w", err)
	}
	return nil
}

// syncDemoCounters 以实际关联数重算演示数据的派生计数。
// 重算而非累加，使重复执行不会放大计数；范围限定在演示记录内，不触碰站点既有数据。
func syncDemoCounters(tx *gorm.DB, dataset demoDataset) error {
	// 文章评论数与评论域口径一致：统计该文章下未软删除的全部评论，不限审核状态。
	err := tx.Model(&model.Article{}).Where("slug IN ?", demoArticleSlugs(dataset)).
		UpdateColumn("comment_count", gorm.Expr(
			"(SELECT COUNT(*) FROM comments WHERE comments.article_id = articles.id AND comments.deleted_at IS NULL)")).
		Error
	if err != nil {
		return fmt.Errorf("重算演示文章评论数失败: %w", err)
	}

	// 分类文章数与标签使用数统计挂载关系，与文章域维护口径一致，不限文章状态。
	err = tx.Model(&model.Category{}).Where("slug IN ?", demoCategorySlugs(dataset)).
		UpdateColumn("article_count", gorm.Expr(
			"(SELECT COUNT(*) FROM article_categories WHERE article_categories.category_id = categories.id)")).
		Error
	if err != nil {
		return fmt.Errorf("重算演示分类文章数失败: %w", err)
	}

	err = tx.Model(&model.Tag{}).Where("name IN ?", demoTagNames(dataset)).
		UpdateColumn("usage_count", gorm.Expr(
			"(SELECT COUNT(*) FROM article_tags WHERE article_tags.tag_id = tags.id)")).
		Error
	if err != nil {
		return fmt.Errorf("重算演示标签使用数失败: %w", err)
	}

	return nil
}

// demoArticleSlugs 收集数据集内的文章 slug。
func demoArticleSlugs(dataset demoDataset) []string {
	slugs := make([]string, 0, len(dataset.Articles))
	for _, item := range dataset.Articles {
		slugs = append(slugs, item.Slug)
	}
	return slugs
}

// demoCategorySlugs 收集数据集内的分类 slug。
func demoCategorySlugs(dataset demoDataset) []string {
	slugs := make([]string, 0, len(dataset.Categories))
	for _, item := range dataset.Categories {
		slugs = append(slugs, item.Slug)
	}
	return slugs
}

// demoTagNames 收集数据集内的标签名称。
func demoTagNames(dataset demoDataset) []string {
	names := make([]string, 0, len(dataset.Tags))
	for _, item := range dataset.Tags {
		names = append(names, item.Name)
	}
	return names
}

// demoSlugPattern 演示数据的 slug 形态，与 pkg/slug 的产出保持一致。
// 种子数据绕过仓储直接写库，仓储层的 slug 生成不会执行，故该形态必须由数据集自身保证。
var demoSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// validateDemoSlugs 校验一类记录的 slug 非空、形态合法且互不重复。
// 空 slug 会撞上唯一索引，形态非法会产出无法访问的 URL，重复会直接写入失败。
func validateDemoSlugs(label string, slugs []string) error {
	seen := make(map[string]struct{}, len(slugs))

	for _, value := range slugs {
		if value == "" {
			return fmt.Errorf("演示%s的 slug 不能为空", label)
		}
		if !demoSlugPattern.MatchString(value) {
			return fmt.Errorf("演示%s的 slug 形态非法: %s", label, value)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("演示%s的 slug 重复: %s", label, value)
		}
		seen[value] = struct{}{}
	}

	return nil
}

// validateDemoDataset 校验演示数据集自洽：幂等键唯一、关联存在、覆盖必要状态。
// 数据集由代码常量定义，不自洽即为程序缺陷，须在写入前失败而不是留下半套数据。
func validateDemoDataset(dataset demoDataset) error {
	if len(dataset.Categories) == 0 || len(dataset.Tags) == 0 || len(dataset.Articles) == 0 {
		return errors.New("演示数据集必须同时包含分类、标签与文章")
	}

	categorySlugs := make(map[string]struct{}, len(dataset.Categories))
	categorySlugList := make([]string, 0, len(dataset.Categories))
	for _, item := range dataset.Categories {
		if _, exists := categorySlugs[item.Slug]; exists {
			return fmt.Errorf("演示分类 slug 重复: %s", item.Slug)
		}
		categorySlugs[item.Slug] = struct{}{}
		categorySlugList = append(categorySlugList, item.Slug)
	}
	if err := validateDemoSlugs("分类", categorySlugList); err != nil {
		return err
	}

	tagNames := make(map[string]struct{}, len(dataset.Tags))
	tagSlugs := make([]string, 0, len(dataset.Tags))
	for _, item := range dataset.Tags {
		if _, exists := tagNames[item.Name]; exists {
			return fmt.Errorf("演示标签名称重复: %s", item.Name)
		}
		tagNames[item.Name] = struct{}{}
		tagSlugs = append(tagSlugs, item.Slug)
	}
	if err := validateDemoSlugs("标签", tagSlugs); err != nil {
		return err
	}

	if err := validateDemoArticles(dataset.Articles, categorySlugs, tagNames); err != nil {
		return err
	}
	return validateDemoComments(dataset.Comments, dataset.Articles)
}

// validateDemoArticles 校验文章条目的幂等键与关联引用，并要求至少存在一篇已发布文章。
func validateDemoArticles(
	articles []demoArticle,
	categorySlugs map[string]struct{},
	tagNames map[string]struct{},
) error {
	articleSlugs := make(map[string]struct{}, len(articles))
	slugList := make([]string, 0, len(articles))
	published := 0

	for _, item := range articles {
		if _, exists := articleSlugs[item.Slug]; exists {
			return fmt.Errorf("演示文章 slug 重复: %s", item.Slug)
		}
		articleSlugs[item.Slug] = struct{}{}
		slugList = append(slugList, item.Slug)

		if _, exists := categorySlugs[item.CategorySlug]; !exists {
			return fmt.Errorf("演示文章 %s 引用了不存在的分类 %s", item.Slug, item.CategorySlug)
		}
		for _, tagName := range item.TagNames {
			if _, exists := tagNames[tagName]; !exists {
				return fmt.Errorf("演示文章 %s 引用了不存在的标签 %s", item.Slug, tagName)
			}
		}
		if item.Status == model.ArticleStatusPublished {
			published++
		}
	}

	if err := validateDemoSlugs("文章", slugList); err != nil {
		return err
	}
	if published == 0 {
		return errors.New("演示数据集至少需要一篇已发布文章")
	}
	return nil
}

// validateDemoComments 校验评论引用的文章存在、作者名在同一文章内唯一，且回复指向更早声明的根评论。
// 作者名唯一是评论写入幂等键的前提：评论没有业务唯一列，重复执行只能按「文章 + 作者名」判定。
func validateDemoComments(comments []demoComment, articles []demoArticle) error {
	articleSlugs := make(map[string]struct{}, len(articles))
	for _, item := range articles {
		articleSlugs[item.Slug] = struct{}{}
	}

	rootAuthors := make(map[string]struct{})
	allAuthors := make(map[string]struct{})
	for _, item := range comments {
		if _, exists := articleSlugs[item.ArticleSlug]; !exists {
			return fmt.Errorf("演示评论引用了不存在的文章 %s", item.ArticleSlug)
		}
		if item.AuthorName == "" {
			return fmt.Errorf("演示评论缺少作者名，无法作为幂等键与回复目标: 文章 %s", item.ArticleSlug)
		}

		key := demoRootAuthorKey(item.ArticleSlug, item.AuthorName)
		if _, exists := allAuthors[key]; exists {
			return fmt.Errorf("同一文章下的评论作者名重复，幂等键将有歧义: %s / %s",
				item.ArticleSlug, item.AuthorName)
		}
		allAuthors[key] = struct{}{}

		if item.ReplyTo == "" {
			rootAuthors[key] = struct{}{}
			continue
		}

		// 回复只允许指向已声明且更早出现的根评论，避免出现悬空父评论。
		if _, exists := rootAuthors[demoRootAuthorKey(item.ArticleSlug, item.ReplyTo)]; !exists {
			return fmt.Errorf("演示回复 %s 指向了不存在的根评论 %s（文章 %s）",
				item.AuthorName, item.ReplyTo, item.ArticleSlug)
		}
	}

	return nil
}

// demoRootAuthorKey 构造根评论的定位键，文章与作者名共同决定唯一目标。
func demoRootAuthorKey(articleSlug, authorName string) string {
	return articleSlug + "\x00" + authorName
}

// describeDemoDataset 返回数据集规模的简述，供种子命令输出。
func describeDemoDataset(dataset demoDataset) string {
	return strings.Join([]string{
		fmt.Sprintf("分类 %d", len(dataset.Categories)),
		fmt.Sprintf("标签 %d", len(dataset.Tags)),
		fmt.Sprintf("文章 %d", len(dataset.Articles)),
		fmt.Sprintf("评论 %d", len(dataset.Comments)),
	}, " / ")
}
