package service

import (
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// captureAuthorList 构造捕获查询参数的作者文章测试替身。
func captureAuthorList(captured **repository.ArticleListParams) *fakeArticleRepo {
	return &fakeArticleRepo{
		getByAuthor: func(authorID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			*captured = params
			return []*model.Article{publishedArticle(1)}, 1, nil
		},
	}
}

// TestGetArticlesByAuthorForcesPublishedForAnonymous 验证匿名请求无法借状态参数拉取非公开文章。
func TestGetArticlesByAuthorForcesPublishedForAnonymous(t *testing.T) {
	var captured *repository.ArticleListParams
	svc := newArticleTestService(captureAuthorList(&captured))

	req := &GetArticleListRequest{Status: "draft", Page: 1, PageSize: 10}
	result, err := svc.GetArticlesByAuthor(1, req, nil)
	if err != nil {
		t.Fatalf("作者文章查询失败: %v", err)
	}
	if captured.Status != model.ArticleStatusPublished {
		t.Errorf("匿名请求应强制已发布状态，实际为 %s", captured.Status)
	}
	if len(result.Articles) != 1 {
		t.Errorf("返回文章数 = %d, 期望 1", len(result.Articles))
	}
}

// TestGetArticlesByAuthorForcesPublishedForRegularUser 验证普通登录用户同样被限制为已发布状态。
func TestGetArticlesByAuthorForcesPublishedForRegularUser(t *testing.T) {
	var captured *repository.ArticleListParams
	userRepo := &fakeUserRepo{user: &domain.User{ID: 2, Role: "user", Status: 1}}
	svc := NewArticleService(captureAuthorList(&captured), userRepo, NewRBACService()).(*ArticleService)

	viewer := uint(2)
	req := &GetArticleListRequest{Status: "private", Page: 1, PageSize: 10}
	if _, err := svc.GetArticlesByAuthor(1, req, &viewer); err != nil {
		t.Fatalf("作者文章查询失败: %v", err)
	}
	if captured.Status != model.ArticleStatusPublished {
		t.Errorf("普通用户应强制已发布状态，实际为 %s", captured.Status)
	}
}

// TestGetArticlesByAuthorKeepsStatusForAdmin 验证管理员可按请求状态筛选作者文章。
func TestGetArticlesByAuthorKeepsStatusForAdmin(t *testing.T) {
	var captured *repository.ArticleListParams
	svc := newArticleTestService(captureAuthorList(&captured))

	viewer := uint(1)
	req := &GetArticleListRequest{Status: "draft", Page: 1, PageSize: 10}
	if _, err := svc.GetArticlesByAuthor(1, req, &viewer); err != nil {
		t.Fatalf("作者文章查询失败: %v", err)
	}
	if captured.Status != model.ArticleStatusDraft {
		t.Errorf("管理员应保留请求的筛选状态，实际为 %s", captured.Status)
	}
}
