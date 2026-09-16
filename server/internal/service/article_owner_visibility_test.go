package service

import (
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// captureListParams 构造捕获列表查询参数并返回混合状态文章的测试替身。
// 三篇文章模拟"本人草稿 + 他人已发布 + 他人草稿"的数据集，替身内模拟
// SQL 的可见边界与状态筛选语义，供断言 service 层的完整可见性行为。
func captureListParams(captured *repository.ArticleListParams) *fakeArticleRepo {
	allArticles := []*model.Article{
		{ID: 10, Title: "本人草稿", AuthorID: 2, Status: model.ArticleStatusDraft},
		{ID: 11, Title: "他人已发布", AuthorID: 3, Status: model.ArticleStatusPublished},
		{ID: 12, Title: "他人草稿", AuthorID: 4, Status: model.ArticleStatusDraft},
	}

	return &fakeArticleRepo{
		list: func(params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			*captured = *params
			visible := make([]*model.Article, 0, len(allArticles))
			for _, article := range allArticles {
				// 模拟 SQL 的作者可见边界：边界外的非公开文章不可见。
				if params.VisibleAuthorID != 0 &&
					article.Status != model.ArticleStatusPublished &&
					article.AuthorID != params.VisibleAuthorID {
					continue
				}
				// 模拟 SQL 的状态筛选：与请求状态不符的记录不返回。
				if params.Status != "" && article.Status != params.Status {
					continue
				}
				visible = append(visible, article)
			}
			return visible, int64(len(visible)), nil
		},
	}
}

// TestGetArticleListShowsOwnDraftsToEditor editor 无 article:manage 权限，
// 列表可见范围应放宽为"已发布 + 本人全状态"，草稿不再从列表中消失。
func TestGetArticleListShowsOwnDraftsToEditor(t *testing.T) {
	captured := repository.ArticleListParams{}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 2, Role: "editor", Status: 1}}
	svc := NewArticleService(captureListParams(&captured), userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	viewer := uint(2)
	result, err := svc.GetArticleList(&GetArticleListRequest{Page: 1, PageSize: 10}, &viewer)
	if err != nil {
		t.Fatalf("editor 列表查询失败: %v", err)
	}

	if captured.Status != "" {
		t.Errorf("editor 未传状态时不应被强制改写状态参数，实际为 %s", captured.Status)
	}
	if captured.VisibleAuthorID != 2 {
		t.Errorf("可见性边界应锚定 editor 本人，实际为 %d", captured.VisibleAuthorID)
	}

	// 本人草稿与他人的已发布文章可见，他人的草稿必须被过滤。
	if len(result.Articles) != 2 {
		t.Fatalf("可见文章数 = %d, 期望 2", len(result.Articles))
	}
	if result.Articles[0].ID != 10 || result.Articles[1].ID != 11 {
		t.Errorf("可见文章 = %d,%d, 期望本人草稿 10 与他人已发布 11", result.Articles[0].ID, result.Articles[1].ID)
	}
}

// TestGetArticleListEditorDraftFilterAnchorsToSelf editor 显式筛选草稿时，
// 结果必须锚定本人，不得借状态参数拉取他人草稿。
func TestGetArticleListEditorDraftFilterAnchorsToSelf(t *testing.T) {
	captured := repository.ArticleListParams{}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 2, Role: "editor", Status: 1}}
	svc := NewArticleService(captureListParams(&captured), userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	viewer := uint(2)
	result, err := svc.GetArticleList(&GetArticleListRequest{Page: 1, PageSize: 10, Status: "draft"}, &viewer)
	if err != nil {
		t.Fatalf("editor 草稿筛选失败: %v", err)
	}

	if captured.Status != model.ArticleStatusDraft {
		t.Errorf("请求的草稿筛选应原样下发，实际为 %s", captured.Status)
	}
	if captured.VisibleAuthorID != 2 {
		t.Errorf("草稿筛选必须叠加本人可见边界，实际为 %d", captured.VisibleAuthorID)
	}
	if len(result.Articles) != 1 || result.Articles[0].ID != 10 {
		t.Errorf("草稿筛选应仅命中本人草稿，实际返回 %v", result.Articles)
	}
}

// TestGetArticleListKeepsFullScopeForAdmin 管理员语义不变：保留请求状态且无作者边界。
func TestGetArticleListKeepsFullScopeForAdmin(t *testing.T) {
	captured := repository.ArticleListParams{}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(captureListParams(&captured), userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	viewer := uint(1)
	if _, err := svc.GetArticleList(&GetArticleListRequest{Page: 1, PageSize: 10, Status: "draft"}, &viewer); err != nil {
		t.Fatalf("管理员列表查询失败: %v", err)
	}

	if captured.Status != model.ArticleStatusDraft {
		t.Errorf("管理员应保留请求筛选状态，实际为 %s", captured.Status)
	}
	if captured.VisibleAuthorID != 0 {
		t.Errorf("管理员不应被施加作者可见边界，实际为 %d", captured.VisibleAuthorID)
	}
}

// TestGetArticleListForcesPublishedForAnonymous 匿名请求维持仅已发布语义。
func TestGetArticleListForcesPublishedForAnonymous(t *testing.T) {
	captured := repository.ArticleListParams{}
	svc := NewArticleService(captureListParams(&captured), &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	if _, err := svc.GetArticleList(&GetArticleListRequest{Page: 1, PageSize: 10}, nil); err != nil {
		t.Fatalf("匿名列表查询失败: %v", err)
	}

	if captured.Status != model.ArticleStatusPublished {
		t.Errorf("匿名请求应强制已发布状态，实际为 %s", captured.Status)
	}
	if captured.VisibleAuthorID != 0 {
		t.Errorf("匿名请求不应有作者可见边界，实际为 %d", captured.VisibleAuthorID)
	}
}

// TestGetArticlesByAuthorShowsAllToSelf 作者本人查看自己的作者页列表时可见全部状态。
func TestGetArticlesByAuthorShowsAllToSelf(t *testing.T) {
	captured := repository.ArticleListParams{}
	var capturedAuthorID uint
	userRepo := &fakeUserRepo{user: &domain.User{ID: 2, Role: "editor", Status: 1}}
	repo := &fakeArticleRepo{
		getByAuthor: func(authorID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			capturedAuthorID = authorID
			captured = *params
			return []*model.Article{publishedArticle(1)}, 1, nil
		},
	}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	viewer := uint(2)
	if _, err := svc.GetArticlesByAuthor(2, &GetArticleListRequest{Page: 1, PageSize: 10}, &viewer); err != nil {
		t.Fatalf("作者本人列表查询失败: %v", err)
	}

	if capturedAuthorID != 2 || captured.VisibleAuthorID != 2 {
		t.Errorf("作者本人视角应放开全状态可见，参数为 author=%d visible=%d", capturedAuthorID, captured.VisibleAuthorID)
	}
	if captured.Status != "" {
		t.Errorf("作者本人未传状态时不应强制改写，实际为 %s", captured.Status)
	}
}

// TestGetArticlesByAuthorAnchorsPublishedForOthers 非本人查看他人作者页维持仅已发布语义。
func TestGetArticlesByAuthorAnchorsPublishedForOthers(t *testing.T) {
	captured := repository.ArticleListParams{}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 3, Role: "user", Status: 1}}
	repo := &fakeArticleRepo{
		getByAuthor: func(authorID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			captured = *params
			return []*model.Article{publishedArticle(1)}, 1, nil
		},
	}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	viewer := uint(3)
	if _, err := svc.GetArticlesByAuthor(2, &GetArticleListRequest{Page: 1, PageSize: 10}, &viewer); err != nil {
		t.Fatalf("他人作者页查询失败: %v", err)
	}

	if captured.Status != model.ArticleStatusPublished {
		t.Errorf("非本人视角应强制已发布，实际为 %s", captured.Status)
	}
	if captured.VisibleAuthorID != 0 {
		t.Errorf("非本人视角不应有作者可见边界，实际为 %d", captured.VisibleAuthorID)
	}
}
