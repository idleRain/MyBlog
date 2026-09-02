package repository

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"MyBlog/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrArticleNotFound 文章不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrArticleNotFound = errors.New("文章不存在")

// ArticleRepositoryInterface 文章仓储接口
type ArticleRepositoryInterface interface {
	// 基础CRUD操作
	Create(article *model.Article) error
	GetByID(id uint) (*model.Article, error)
	GetBySlug(slug string) (*model.Article, error)
	Update(article *model.Article) error
	Delete(id uint) error

	// 查询操作
	List(params *ArticleListParams) ([]*model.Article, int64, error)
	GetByAuthor(authorID uint, params *ArticleListParams) ([]*model.Article, int64, error)
	GetByCategory(categoryID uint, params *ArticleListParams) ([]*model.Article, int64, error)
	GetByTag(tagID uint, params *ArticleListParams) ([]*model.Article, int64, error)
	Search(keyword string, params *ArticleListParams) ([]*model.Article, int64, error)

	// 统计操作
	GetPopular(limit int) ([]*model.Article, error)
	GetRecent(limit int) ([]*model.Article, error)
	IncrementViewCount(id uint) error
	UpdateCommentCount(id uint) error

	// 互动操作
	AddLike(articleID, userID uint) (bool, error)
	RemoveLike(articleID, userID uint) (bool, error)
	AddBookmark(articleID, userID uint) (bool, error)
	RemoveBookmark(articleID, userID uint) (bool, error)

	// 分类和标签关联
	AddCategory(articleID, categoryID uint) error
	RemoveCategory(articleID, categoryID uint) error
	AddTag(articleID, tagID uint) error
	RemoveTag(articleID, tagID uint) error
	SyncTags(articleID uint, tagIDs []uint) error
	SyncCategories(articleID uint, categoryIDs []uint) error

	// 状态管理
	Publish(id uint) error
	Unpublish(id uint) error
	Archive(id uint) error
	SetPrivate(id uint) error
}

// ArticleListParams 文章列表查询参数
type ArticleListParams struct {
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Status   model.ArticleStatus `json:"status"`
	AuthorID uint                `json:"authorId"`
	SortBy   string              `json:"sortBy"` // 排序字段：created_at、updated_at、published_at、view_count、like_count
	Order    string              `json:"order"`  // asc, desc
	Search   string              `json:"search"`
}

// ArticleRepository 文章仓储实现
type ArticleRepository struct {
	db *gorm.DB
}

// NewArticleRepository 创建文章仓储实例
func NewArticleRepository(db *gorm.DB) ArticleRepositoryInterface {
	return &ArticleRepository{db: db}
}

// Create 创建文章
func (r *ArticleRepository) Create(article *model.Article) error {
	if article.Slug == "" {
		article.Slug = generateSlug(article.Title)
	}

	// 确保slug唯一
	if err := r.ensureUniqueSlug(article); err != nil {
		return err
	}

	// 设置发布时间
	if article.Status == model.ArticleStatusPublished && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	return r.db.Create(article).Error
}

// GetByID 根据ID获取文章
func (r *ArticleRepository) GetByID(id uint) (*model.Article, error) {
	var article model.Article
	err := r.db.Preload("Author").
		Preload("Category").
		Preload("Categories").
		Preload("Tags").
		First(&article, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}

	return &article, nil
}

// GetBySlug 根据Slug获取文章
func (r *ArticleRepository) GetBySlug(slug string) (*model.Article, error) {
	var article model.Article
	err := r.db.Preload("Author").
		Preload("Category").
		Preload("Categories").
		Preload("Tags").
		Where("slug = ?", slug).
		First(&article).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}

	return &article, nil
}

// Update 更新文章
func (r *ArticleRepository) Update(article *model.Article) error {
	// 更新时若 slug 为空，回退为按标题生成，避免唯一索引写入空值。
	if article.Slug == "" {
		article.Slug = generateSlug(article.Title)
	}

	// 检查slug唯一性
	if err := r.ensureUniqueSlug(article); err != nil {
		return err
	}

	// 如果状态变更为发布，设置发布时间
	if article.Status == model.ArticleStatusPublished && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	return r.db.Save(article).Error
}

// Delete 删除文章，采用软删除，并在同一事务内清理关联关系与回滚分类标签计数。
func (r *ArticleRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 查询文章关联的分类与标签，用于回滚计数。
		var categories []model.ArticleCategory
		if err := tx.Where("article_id = ?", id).Find(&categories).Error; err != nil {
			return err
		}
		var tags []model.ArticleTag
		if err := tx.Where("article_id = ?", id).Find(&tags).Error; err != nil {
			return err
		}

		// 删除文章关联的分类与标签记录。
		if err := tx.Where("article_id = ?", id).Delete(&model.ArticleCategory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Delete(&model.ArticleTag{}).Error; err != nil {
			return err
		}

		// 回滚分类文章计数与标签使用计数。
		for _, category := range categories {
			if err := r.decrementCategoryCount(tx, category.CategoryID); err != nil {
				return err
			}
		}
		for _, tag := range tags {
			if err := r.decrementTagCount(tx, tag.TagID); err != nil {
				return err
			}
		}

		// 软删除文章本体。
		return tx.Delete(&model.Article{}, id).Error
	})
}

// List 获取文章列表
func (r *ArticleRepository) List(params *ArticleListParams) ([]*model.Article, int64, error) {
	query := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags")

	// 应用筛选条件
	query = r.applyFilters(query, params)

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序
	query = r.applyPagination(query, params)
	query = r.applySorting(query, params)

	var articles []*model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetByAuthor 获取指定作者的文章
func (r *ArticleRepository) GetByAuthor(authorID uint, params *ArticleListParams) ([]*model.Article, int64, error) {
	params.AuthorID = authorID
	return r.List(params)
}

// GetByCategory 获取指定分类的文章
func (r *ArticleRepository) GetByCategory(categoryID uint, params *ArticleListParams) ([]*model.Article, int64, error) {
	query := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Where("category_id = ? OR EXISTS (SELECT 1 FROM article_categories WHERE article_categories.article_id = articles.id AND article_categories.category_id = ?)", categoryID, categoryID)

	// 应用其他筛选条件
	query = r.applyFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = r.applyPagination(query, params)
	query = r.applySorting(query, params)

	var articles []*model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetByTag 获取指定标签的文章
func (r *ArticleRepository) GetByTag(tagID uint, params *ArticleListParams) ([]*model.Article, int64, error) {
	query := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Joins("JOIN article_tags ON article_tags.article_id = articles.id").
		Where("article_tags.tag_id = ?", tagID)

	// 应用其他筛选条件
	query = r.applyFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = r.applyPagination(query, params)
	query = r.applySorting(query, params)

	var articles []*model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// Search 全文搜索文章
func (r *ArticleRepository) Search(keyword string, params *ArticleListParams) ([]*model.Article, int64, error) {
	if keyword == "" {
		return r.List(params)
	}

	searchTerm := "%" + keyword + "%"
	query := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Where("title LIKE ? OR content LIKE ? OR summary LIKE ?", searchTerm, searchTerm, searchTerm)

	// 应用其他筛选条件
	query = r.applyFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = r.applyPagination(query, params)
	query = r.applySorting(query, params)

	var articles []*model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetPopular 获取热门文章
func (r *ArticleRepository) GetPopular(limit int) ([]*model.Article, error) {
	var articles []*model.Article
	err := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Where("status = ?", model.ArticleStatusPublished).
		Order("view_count DESC, like_count DESC").
		Limit(limit).
		Find(&articles).Error

	return articles, err
}

// GetRecent 获取最新文章
func (r *ArticleRepository) GetRecent(limit int) ([]*model.Article, error) {
	var articles []*model.Article
	err := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Where("status = ?", model.ArticleStatusPublished).
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error

	return articles, err
}

// IncrementViewCount 增加浏览量
func (r *ArticleRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// UpdateCommentCount 更新评论数
func (r *ArticleRepository) UpdateCommentCount(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("(SELECT COUNT(*) FROM comments WHERE article_id = ? AND deleted_at IS NULL)", id)).Error
}

// AddLike 添加点赞记录，依赖唯一索引防重复，返回是否新增并递增点赞计数。
func (r *ArticleRepository) AddLike(articleID, userID uint) (bool, error) {
	var added bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 依赖 (article_id, user_id) 唯一索引，重复点赞时静默忽略。
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.ArticleLike{
			ArticleID: articleID,
			UserID:    userID,
		})
		if result.Error != nil {
			return result.Error
		}
		added = result.RowsAffected > 0
		// 仅在真正新增时递增点赞计数，避免重复点赞导致计数虚高。
		if added {
			return r.incrementLikeCount(tx, articleID)
		}
		return nil
	})
	return added, err
}

// RemoveLike 移除点赞记录，返回是否实际删除并递减点赞计数。
func (r *ArticleRepository) RemoveLike(articleID, userID uint) (bool, error) {
	var removed bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("article_id = ? AND user_id = ?", articleID, userID).
			Delete(&model.ArticleLike{})
		if result.Error != nil {
			return result.Error
		}
		removed = result.RowsAffected > 0
		// 仅在真正删除时递减计数，未点赞时无需回滚。
		if removed {
			return r.decrementLikeCount(tx, articleID)
		}
		return nil
	})
	return removed, err
}

// AddBookmark 添加收藏记录，依赖唯一索引防重复，返回是否新增并递增收藏计数。
func (r *ArticleRepository) AddBookmark(articleID, userID uint) (bool, error) {
	var added bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.ArticleBookmark{
			ArticleID: articleID,
			UserID:    userID,
		})
		if result.Error != nil {
			return result.Error
		}
		added = result.RowsAffected > 0
		if added {
			return r.incrementBookmarkCount(tx, articleID)
		}
		return nil
	})
	return added, err
}

// RemoveBookmark 移除收藏记录，返回是否实际删除并递减收藏计数。
func (r *ArticleRepository) RemoveBookmark(articleID, userID uint) (bool, error) {
	var removed bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("article_id = ? AND user_id = ?", articleID, userID).
			Delete(&model.ArticleBookmark{})
		if result.Error != nil {
			return result.Error
		}
		removed = result.RowsAffected > 0
		if removed {
			return r.decrementBookmarkCount(tx, articleID)
		}
		return nil
	})
	return removed, err
}

// AddCategory 添加分类关联
func (r *ArticleRepository) AddCategory(articleID, categoryID uint) error {
	articleCategory := &model.ArticleCategory{
		ArticleID:  articleID,
		CategoryID: categoryID,
	}
	return r.db.Create(articleCategory).Error
}

// RemoveCategory 移除分类关联
func (r *ArticleRepository) RemoveCategory(articleID, categoryID uint) error {
	return r.db.Where("article_id = ? AND category_id = ?", articleID, categoryID).
		Delete(&model.ArticleCategory{}).Error
}

// AddTag 添加标签关联
func (r *ArticleRepository) AddTag(articleID, tagID uint) error {
	articleTag := &model.ArticleTag{
		ArticleID: articleID,
		TagID:     tagID,
	}
	return r.db.Create(articleTag).Error
}

// RemoveTag 移除标签关联
func (r *ArticleRepository) RemoveTag(articleID, tagID uint) error {
	return r.db.Where("article_id = ? AND tag_id = ?", articleID, tagID).
		Delete(&model.ArticleTag{}).Error
}

// SyncTags 同步标签关联，替换文章的全部标签，并维护各标签使用计数。
func (r *ArticleRepository) SyncTags(articleID uint, tagIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 查询现有标签关联，用于回滚被移除标签的使用计数。
		var existingTags []model.ArticleTag
		if err := tx.Where("article_id = ?", articleID).Find(&existingTags).Error; err != nil {
			return err
		}

		// 删除现有关联
		if err := tx.Where("article_id = ?", articleID).Delete(&model.ArticleTag{}).Error; err != nil {
			return err
		}

		// 回滚被移除标签的使用计数
		for _, existing := range existingTags {
			if err := r.decrementTagCount(tx, existing.TagID); err != nil {
				return err
			}
		}

		// 添加新关联并递增对应标签的使用计数
		for _, tagID := range tagIDs {
			articleTag := &model.ArticleTag{
				ArticleID: articleID,
				TagID:     tagID,
			}
			if err := tx.Create(articleTag).Error; err != nil {
				return err
			}
			if err := r.incrementTagCount(tx, tagID); err != nil {
				return err
			}
		}

		return nil
	})
}

// SyncCategories 同步分类关联，替换文章的全部分类，并维护各分类文章计数。
func (r *ArticleRepository) SyncCategories(articleID uint, categoryIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 查询现有分类关联，用于回滚被移除分类的文章计数。
		var existingCategories []model.ArticleCategory
		if err := tx.Where("article_id = ?", articleID).Find(&existingCategories).Error; err != nil {
			return err
		}

		// 删除现有关联
		if err := tx.Where("article_id = ?", articleID).Delete(&model.ArticleCategory{}).Error; err != nil {
			return err
		}

		// 回滚被移除分类的文章计数
		for _, existing := range existingCategories {
			if err := r.decrementCategoryCount(tx, existing.CategoryID); err != nil {
				return err
			}
		}

		// 添加新关联并递增对应分类的文章计数
		for _, categoryID := range categoryIDs {
			articleCategory := &model.ArticleCategory{
				ArticleID:  articleID,
				CategoryID: categoryID,
			}
			if err := tx.Create(articleCategory).Error; err != nil {
				return err
			}
			if err := r.incrementCategoryCount(tx, categoryID); err != nil {
				return err
			}
		}

		return nil
	})
}

// Publish 发布文章
func (r *ArticleRepository) Publish(id uint) error {
	now := time.Now()
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.ArticleStatusPublished,
			"published_at": &now,
		}).Error
}

// Unpublish 取消发布文章
func (r *ArticleRepository) Unpublish(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Update("status", model.ArticleStatusDraft).Error
}

// Archive 归档文章，记录归档时间，保留发布时间供历史展示。
func (r *ArticleRepository) Archive(id uint) error {
	now := time.Now()
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.ArticleStatusArchived,
			"archived_at": &now,
		}).Error
}

// SetPrivate 设置为私有文章
func (r *ArticleRepository) SetPrivate(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Update("status", model.ArticleStatusPrivate).Error
}

// 私有辅助方法

// ensureUniqueSlug 确保slug唯一
func (r *ArticleRepository) ensureUniqueSlug(article *model.Article) error {
	originalSlug := article.Slug
	counter := 1

	for {
		var count int64
		query := r.db.Model(&model.Article{}).Where("slug = ?", article.Slug)

		// 如果是更新操作，排除当前文章
		if article.ID != 0 {
			query = query.Where("id != ?", article.ID)
		}

		if err := query.Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			break
		}

		article.Slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	return nil
}

// incrementLikeCount 递增文章点赞计数，供点赞事务内调用。
func (r *ArticleRepository) incrementLikeCount(tx *gorm.DB, articleID uint) error {
	return tx.Model(&model.Article{}).
		Where("id = ?", articleID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// decrementLikeCount 递减文章点赞计数，点赞数不小于零。
func (r *ArticleRepository) decrementLikeCount(tx *gorm.DB, articleID uint) error {
	return tx.Model(&model.Article{}).
		Where("id = ?", articleID).
		UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).Error
}

// incrementBookmarkCount 递增文章收藏计数，供收藏事务内调用。
func (r *ArticleRepository) incrementBookmarkCount(tx *gorm.DB, articleID uint) error {
	return tx.Model(&model.Article{}).
		Where("id = ?", articleID).
		UpdateColumn("bookmark_count", gorm.Expr("bookmark_count + 1")).Error
}

// decrementBookmarkCount 递减文章收藏计数，收藏数不小于零。
func (r *ArticleRepository) decrementBookmarkCount(tx *gorm.DB, articleID uint) error {
	return tx.Model(&model.Article{}).
		Where("id = ?", articleID).
		UpdateColumn("bookmark_count", gorm.Expr("CASE WHEN bookmark_count > 0 THEN bookmark_count - 1 ELSE 0 END")).Error
}

// incrementTagCount 递增标签使用计数，供标签同步事务内调用。
func (r *ArticleRepository) incrementTagCount(tx *gorm.DB, tagID uint) error {
	return tx.Model(&model.Tag{}).
		Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

// decrementTagCount 递减标签使用计数，使用数不小于零。
func (r *ArticleRepository) decrementTagCount(tx *gorm.DB, tagID uint) error {
	return tx.Model(&model.Tag{}).
		Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("CASE WHEN usage_count > 0 THEN usage_count - 1 ELSE 0 END")).Error
}

// incrementCategoryCount 递增分类文章计数，供分类同步事务内调用。
func (r *ArticleRepository) incrementCategoryCount(tx *gorm.DB, categoryID uint) error {
	return tx.Model(&model.Category{}).
		Where("id = ?", categoryID).
		UpdateColumn("article_count", gorm.Expr("article_count + 1")).Error
}

// decrementCategoryCount 递减分类文章计数，文章数不小于零。
func (r *ArticleRepository) decrementCategoryCount(tx *gorm.DB, categoryID uint) error {
	return tx.Model(&model.Category{}).
		Where("id = ?", categoryID).
		UpdateColumn("article_count", gorm.Expr("CASE WHEN article_count > 0 THEN article_count - 1 ELSE 0 END")).Error
}

// generateSlug 从标题生成slug，中文标题过滤后为空时回退为基于时间戳的标识。
func generateSlug(title string) string {
	// 简单的slug生成逻辑，实际项目中可能需要更复杂的处理
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// 移除特殊字符，仅保留字母、数字与连字符。
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// 中文标题过滤后结果可能为空，回退为 article-<纳秒时间戳>-<随机数>，确保 slug 非空且唯一。
	if result.Len() == 0 {
		return fmt.Sprintf("article-%d-%d", time.Now().UnixNano(), rand.IntN(1_000_000))
	}

	return result.String()
}

// applyFilters 应用筛选条件
func (r *ArticleRepository) applyFilters(query *gorm.DB, params *ArticleListParams) *gorm.DB {
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if params.AuthorID != 0 {
		query = query.Where("author_id = ?", params.AuthorID)
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("title LIKE ? OR content LIKE ? OR summary LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	return query
}

// applyPagination 应用分页
func (r *ArticleRepository) applyPagination(query *gorm.DB, params *ArticleListParams) *gorm.DB {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	offset := (params.Page - 1) * params.PageSize
	return query.Offset(offset).Limit(params.PageSize)
}

// applySorting 应用排序
func (r *ArticleRepository) applySorting(query *gorm.DB, params *ArticleListParams) *gorm.DB {
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.Order == "" {
		params.Order = "desc"
	}

	orderStr := params.SortBy + " " + strings.ToUpper(params.Order)
	return query.Order(orderStr)
}
