package service

import (
	"testing"
	"time"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// archiveArticle 构造指定发布时间的已发布文章，仅填充归档页消费的最小列集。
func archiveArticle(id uint, publishedAt time.Time) *model.Article {
	return &model.Article{
		ID:          id,
		Title:       "归档文章",
		Status:      model.ArticleStatusPublished,
		PublishedAt: &publishedAt,
	}
}

// TestGetArticleArchivesEmptyRepo 空库时归档返回空列表。
func TestGetArticleArchivesEmptyRepo(t *testing.T) {
	repo := &fakeArticleRepo{}
	svc := articleServiceForArchives(repo)

	groups, err := svc.GetArticleArchives()
	if err != nil {
		t.Fatalf("空库归档查询失败: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("空库应返回空归档，实际 %v", groups)
	}
}

// TestGetArticleArchivesSingleMonth 单月情形下聚合计数与行装配一致。
func TestGetArticleArchivesSingleMonth(t *testing.T) {
	publishedAt := time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC)
	repo := &fakeArticleRepo{
		archiveGroups: []repository.ArticleArchiveGroup{{Year: 2026, Month: 3, Total: 1}},
		archiveRows: []*model.Article{
			{ID: 1, Title: "三月文章", Slug: "march-post", PublishedAt: &publishedAt},
		},
	}
	svc := articleServiceForArchives(repo)

	groups, err := svc.GetArticleArchives()
	if err != nil {
		t.Fatalf("单月归档查询失败: %v", err)
	}

	if len(groups) != 1 || groups[0].Year != 2026 {
		t.Fatalf("应产生一个年度分组，实际 %v", groups)
	}
	if groups[0].Total != 1 {
		t.Errorf("年度总数 = %d，期望 1", groups[0].Total)
	}
	if len(groups[0].Months) != 1 || groups[0].Months[0].Month != 3 {
		t.Fatalf("应产生单一月份分组，实际 %v", groups[0].Months)
	}
	articles := groups[0].Months[0].Articles
	if len(articles) != 1 || articles[0].Slug != "march-post" {
		t.Errorf("月内文章装配错误，实际 %v", articles)
	}
}

// TestGetArticleArchivesAcrossYears 跨年多篇时按年月分组，计数取自聚合行。
func TestGetArticleArchivesAcrossYears(t *testing.T) {
	december := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	june := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	july := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	repo := &fakeArticleRepo{
		// 聚合行带 SQL 侧排序结果：年份倒序后月份倒序。
		archiveGroups: []repository.ArticleArchiveGroup{
			{Year: 2026, Month: 7, Total: 1},
			{Year: 2026, Month: 6, Total: 2},
			{Year: 2025, Month: 12, Total: 1},
		},
		// 文章行按发布时间倒序。
		archiveRows: []*model.Article{
			{ID: 10, Title: "七月", Slug: "july", PublishedAt: &july},
			{ID: 11, Title: "六月甲", Slug: "june-a", PublishedAt: &june},
			{ID: 12, Title: "六月乙", Slug: "june-b", PublishedAt: &june},
			{ID: 13, Title: "十二月", Slug: "december", PublishedAt: &december},
		},
	}
	svc := articleServiceForArchives(repo)

	groups, err := svc.GetArticleArchives()
	if err != nil {
		t.Fatalf("跨年归档查询失败: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("应产生两个年度分组，实际 %d", len(groups))
	}
	first, second := groups[0], groups[1]
	if first.Year != 2026 || second.Year != 2025 {
		t.Fatalf("年度分组应为 2026 与 2025，实际 %d / %d", first.Year, second.Year)
	}
	// 年度总数取自聚合行的 COUNT，文章行按月归位。
	if first.Total != 3 || second.Total != 1 {
		t.Errorf("年度总数应为 3 与 1，实际 %d / %d", first.Total, second.Total)
	}
	if len(first.Months) != 2 || first.Months[0].Month != 7 || first.Months[1].Month != 6 {
		t.Errorf("月份分组应为 7 月与 6 月，实际 %v", first.Months)
	}
	if len(first.Months[1].Articles) != 2 {
		t.Errorf("6 月文章数应为 2，实际 %d", len(first.Months[1].Articles))
	}
	if len(second.Months[0].Articles) != 1 || second.Months[0].Articles[0].Slug != "december" {
		t.Errorf("2025 年 12 月装配错误，实际 %v", second.Months[0].Articles)
	}
}

// TestBuildArticleArchivesSkipsMissingPublishedAt 装配层兜底跳过缺失发布时间的行，
// 与仓储侧的 published_at 过滤构成双保险。
func TestBuildArticleArchivesSkipsMissingPublishedAt(t *testing.T) {
	repo := &fakeArticleRepo{
		archiveGroups: []repository.ArticleArchiveGroup{{Year: 2026, Month: 3, Total: 1}},
		archiveRows: []*model.Article{
			{ID: 9, Title: "无发布时间"},
			archiveArticle(2, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)),
		},
	}

	groups := buildArticleArchives(repo.archiveGroups, repo.archiveRows)

	if len(groups) != 1 || groups[0].Total != 1 {
		t.Fatalf("缺失发布时间的行不应计入，实际 %v", groups)
	}
	if len(groups[0].Months[0].Articles) != 1 || groups[0].Months[0].Articles[0].ID != 2 {
		t.Errorf("月内装配应只含合规行，实际 %v", groups[0].Months[0].Articles)
	}
}

// TestGetRelatedArticlesQueriesTagsOnce 断言相关文章的标签候选只触发一次仓储查询。
func TestGetRelatedArticlesQueriesTagsOnce(t *testing.T) {
	categoryID := uint(5)
	current := &model.Article{
		ID:         1,
		CategoryID: &categoryID,
		Status:     model.ArticleStatusPublished,
		Tags:       []model.Tag{{ID: 11}, {ID: 12}, {ID: 13}},
	}

	repo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return current, nil
		},
		getByCategory: func(categoryID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			// 同分类候选不足 limit，触发标签候选查询。
			return nil, 0, nil
		},
		getByTags: func(tagIDs []uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			if len(tagIDs) != 3 {
				t.Errorf("标签候选应一次性携带全部标签，实际 %v", tagIDs)
			}
			return []*model.Article{
				{ID: 2, Status: model.ArticleStatusPublished},
				{ID: 3, Status: model.ArticleStatusPublished},
			}, 2, nil
		},
	}
	svc := articleServiceForArchives(repo)

	related, err := svc.GetRelatedArticles(1, 5)
	if err != nil {
		t.Fatalf("相关文章查询失败: %v", err)
	}

	// 无论挂几个标签，标签候选只允许一次仓储调用。
	if repo.getByTagsCalls != 1 {
		t.Errorf("标签候选应仅查询一次，实际 %d 次", repo.getByTagsCalls)
	}
	if len(related) != 2 {
		t.Errorf("相关文章数 = %d，期望 2", len(related))
	}
}

// articleServiceForArchives 创建供归档与相关文章测试使用的最小服务装配。
func articleServiceForArchives(repo *fakeArticleRepo) *ArticleService {
	return NewArticleService(repo, &fakeUserRepo{}, NewRBACService(), nil, nil).(*ArticleService)
}

// TestGetArticleListUserLookupsIndependentOfSize 验证列表可见性判定不随文章长度查用户表。
func TestGetArticleListUserLookupsIndependentOfSize(t *testing.T) {
	userRepo := &fakeUserRepo{user: &domain.User{ID: 7, Role: "user"}}
	articles := make([]*model.Article, 0, 5)
	for _, id := range []uint{1, 2, 3, 4, 5} {
		draft := draftArticleByAuthor(id, 7)
		articles = append(articles, draft)
	}
	repo := &fakeArticleRepo{
		list: func(params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			return articles, int64(len(articles)), nil
		},
	}
	svc := &ArticleService{articleRepo: repo, userRepo: userRepo, rbacService: NewRBACService()}

	response, err := svc.GetArticleList(&GetArticleListRequest{Page: 1, PageSize: 5}, &[]uint{7}[0])
	if err != nil {
		t.Fatalf("文章列表查询失败: %v", err)
	}

	// 权限解析一次完成，查库次数与列表长度无关（作者本人两字段分支不查库）。
	if userRepo.getByIDCalls != 1 {
		t.Errorf("用户表查询应仅 1 次，实际 %d 次", userRepo.getByIDCalls)
	}
	if len(response.Articles) != 5 {
		t.Errorf("可见文章数 = %d，期望 5", len(response.Articles))
	}
}

// draftArticleByAuthor 构造指定作者的草稿文章，供可见性判定使用。
func draftArticleByAuthor(id uint, authorID uint) *model.Article {
	return &model.Article{ID: id, Title: "草稿", AuthorID: authorID, Status: model.ArticleStatusDraft}
}
