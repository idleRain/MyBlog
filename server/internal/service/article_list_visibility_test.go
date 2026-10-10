package service

import (
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// 可见性模型统一后的回归用例。
// 这些用例锁定两类入口的语义差异：查看者维度入口按身份放宽，
// 公开聚合面入口固定仅已发布，防止后续改动把差异重新变成实现分叉。

// newVisibilityTestService 构造注入指定查看者身份的文章服务测试实例。
// viewer 为 nil 时模拟匿名访问，其余场景按传入用户的角色解析文章管理权限。
func newVisibilityTestService(viewer *domain.User) *ArticleService {
	return NewArticleService(&fakeArticleRepo{}, &fakeUserRepo{user: viewer}, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)
}

// TestApplyListVisibilityByScopeAndViewer 覆盖两种可见性范围在四类查看者下的边界。
// 请求显式筛选草稿，用于区分「保留请求状态」与「强制已发布」两种结果。
func TestApplyListVisibilityByScopeAndViewer(t *testing.T) {
	adminID := uint(1)
	ownerID := uint(2)
	otherID := uint(3)

	cases := []struct {
		name              string
		scope             articleListScope
		viewer            *uint
		viewerRole        string
		ownerID           uint
		wantStatus        model.ArticleStatus
		wantVisibleAuthor uint
		wantCanManage     bool
	}{
		{
			name: "查看者维度-匿名强制已发布", scope: listScopeViewerAware, viewer: nil,
			wantStatus: model.ArticleStatusPublished,
		},
		{
			name: "查看者维度-作者本人保留状态并锚定本人", scope: listScopeViewerAware,
			viewer: &ownerID, viewerRole: "editor", ownerID: ownerID,
			wantStatus: model.ArticleStatusDraft, wantVisibleAuthor: ownerID,
		},
		{
			name: "查看者维度-非本人登录用户强制已发布", scope: listScopeViewerAware,
			viewer: &otherID, viewerRole: "user", ownerID: ownerID,
			wantStatus: model.ArticleStatusPublished,
		},
		{
			name: "查看者维度-管理员保留状态且无作者边界", scope: listScopeViewerAware,
			viewer: &adminID, viewerRole: "admin", ownerID: ownerID,
			wantStatus: model.ArticleStatusDraft, wantCanManage: true,
		},
		{
			name: "仅已发布-匿名", scope: listScopePublishedOnly, viewer: nil,
			wantStatus: model.ArticleStatusPublished,
		},
		{
			name: "仅已发布-登录普通用户", scope: listScopePublishedOnly,
			viewer: &otherID, viewerRole: "user", ownerID: otherID,
			wantStatus: model.ArticleStatusPublished,
		},
		{
			name: "仅已发布-资源所有者本人不放宽", scope: listScopePublishedOnly,
			viewer: &ownerID, viewerRole: "editor", ownerID: ownerID,
			wantStatus: model.ArticleStatusPublished,
		},
		{
			name: "仅已发布-管理员不放宽", scope: listScopePublishedOnly,
			viewer: &adminID, viewerRole: "admin", ownerID: ownerID,
			wantStatus: model.ArticleStatusPublished,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var viewer *domain.User
			if testCase.viewer != nil {
				viewer = &domain.User{ID: *testCase.viewer, Role: testCase.viewerRole, Status: 1}
			}
			svc := newVisibilityTestService(viewer)

			params := &repository.ArticleListParams{Status: model.ArticleStatusDraft}
			canManage := svc.applyListVisibility(params, testCase.scope, testCase.viewer, testCase.ownerID)

			if params.Status != testCase.wantStatus {
				t.Errorf("状态边界 = %q, 期望 %q", params.Status, testCase.wantStatus)
			}
			if params.VisibleAuthorID != testCase.wantVisibleAuthor {
				t.Errorf("作者可见边界 = %d, 期望 %d", params.VisibleAuthorID, testCase.wantVisibleAuthor)
			}
			if canManage != testCase.wantCanManage {
				t.Errorf("管理权限标记 = %v, 期望 %v", canManage, testCase.wantCanManage)
			}
		})
	}
}

// TestGetArticlesByCategoryForcesPublished 分类列表属公开聚合面，
// 请求携带的状态筛选不得放宽可见范围。
func TestGetArticlesByCategoryForcesPublished(t *testing.T) {
	captured := repository.ArticleListParams{}
	repo := &fakeArticleRepo{
		getByCategory: func(categoryID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			captured = *params
			return []*model.Article{publishedArticle(1)}, 1, nil
		},
	}
	svc := newArticleTestService(repo)

	if _, err := svc.GetArticlesByCategory(1, &GetArticleListRequest{Page: 1, PageSize: 10, Status: "draft"}); err != nil {
		t.Fatalf("分类列表查询失败: %v", err)
	}

	if captured.Status != model.ArticleStatusPublished {
		t.Errorf("分类列表应强制已发布，实际为 %s", captured.Status)
	}
	if captured.VisibleAuthorID != 0 {
		t.Errorf("分类列表不应有作者可见边界，实际为 %d", captured.VisibleAuthorID)
	}
}

// TestGetArticlesByTagForcesPublished 标签列表与分类列表同属公开聚合面。
func TestGetArticlesByTagForcesPublished(t *testing.T) {
	captured := repository.ArticleListParams{}
	repo := &fakeArticleRepo{
		getByTag: func(tagID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			captured = *params
			return []*model.Article{publishedArticle(1)}, 1, nil
		},
	}
	svc := newArticleTestService(repo)

	if _, err := svc.GetArticlesByTag(1, &GetArticleListRequest{Page: 1, PageSize: 10, Status: "private"}); err != nil {
		t.Fatalf("标签列表查询失败: %v", err)
	}

	if captured.Status != model.ArticleStatusPublished {
		t.Errorf("标签列表应强制已发布，实际为 %s", captured.Status)
	}
	if captured.VisibleAuthorID != 0 {
		t.Errorf("标签列表不应有作者可见边界，实际为 %d", captured.VisibleAuthorID)
	}
}

// TestSearchArticlesForcesPublished 搜索无查看者维度，只输出已发布文章。
func TestSearchArticlesForcesPublished(t *testing.T) {
	captured := repository.ArticleListParams{}
	repo := &fakeArticleRepo{
		search: func(keyword string, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			captured = *params
			return []*model.Article{publishedArticle(1)}, 1, nil
		},
	}
	svc := newArticleTestService(repo)

	if _, err := svc.SearchArticles("测试", &GetArticleListRequest{Page: 1, PageSize: 10, Status: "draft"}); err != nil {
		t.Fatalf("搜索查询失败: %v", err)
	}

	if captured.Status != model.ArticleStatusPublished {
		t.Errorf("搜索应强制已发布，实际为 %s", captured.Status)
	}
}

// TestGetRelatedArticlesForcesPublished 相关文章同属公开聚合面。
func TestGetRelatedArticlesForcesPublished(t *testing.T) {
	captured := repository.ArticleListParams{}
	categoryID := uint(9)
	repo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return &model.Article{ID: id, Status: model.ArticleStatusPublished, CategoryID: &categoryID}, nil
		},
		getByCategory: func(gotCategoryID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			captured = *params
			return nil, 0, nil
		},
	}
	svc := newArticleTestService(repo)

	if _, err := svc.GetRelatedArticles(1, 5); err != nil {
		t.Fatalf("相关文章查询失败: %v", err)
	}

	if captured.Status != model.ArticleStatusPublished {
		t.Errorf("相关文章应强制已发布，实际为 %s", captured.Status)
	}
}
