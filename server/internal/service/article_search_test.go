package service

import (
	"errors"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// TestSearchArticlesRecordsLog 验证搜索写入搜索日志并返回结果。
func TestSearchArticlesRecordsLog(t *testing.T) {
	statsRepo := &fakeStatsRepo{}
	repo := &fakeArticleRepo{
		search: func(keyword string, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			return []*model.Article{publishedArticle(1)}, 1, nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), statsRepo, &fakeNotificationRepo{}).(*ArticleService)

	result, err := svc.SearchArticles("svelte", &GetArticleListRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}
	if len(result.Articles) != 1 {
		t.Errorf("返回文章数 = %d, 期望 1", len(result.Articles))
	}
	if len(statsRepo.searchLogs) != 1 {
		t.Fatalf("搜索日志数 = %d, 期望 1", len(statsRepo.searchLogs))
	}
	log := statsRepo.searchLogs[0]
	if log.Keyword != "svelte" {
		t.Errorf("日志关键词 = %q, 期望 svelte", log.Keyword)
	}
	if log.Status != model.OperationStatusSuccess {
		t.Errorf("日志状态 = %s, 期望 success", log.Status)
	}
	if log.ResultsCount != 1 {
		t.Errorf("日志结果数 = %d, 期望 1", log.ResultsCount)
	}
}

// TestSearchArticlesSkipsEmptyKeywordLog 验证空关键词不写搜索日志。
func TestSearchArticlesSkipsEmptyKeywordLog(t *testing.T) {
	statsRepo := &fakeStatsRepo{}
	repo := &fakeArticleRepo{
		search: func(keyword string, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			return nil, 0, nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), statsRepo, &fakeNotificationRepo{}).(*ArticleService)

	if _, err := svc.SearchArticles("", &GetArticleListRequest{Page: 1, PageSize: 10}); err != nil {
		t.Fatalf("空关键词搜索失败: %v", err)
	}
	if len(statsRepo.searchLogs) != 0 {
		t.Errorf("搜索日志数 = %d, 期望 0", len(statsRepo.searchLogs))
	}
}

// TestSearchArticlesRecordsFailure 验证搜索失败时日志状态标记为 failed。
func TestSearchArticlesRecordsFailure(t *testing.T) {
	statsRepo := &fakeStatsRepo{}
	repo := &fakeArticleRepo{
		search: func(keyword string, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			return nil, 0, errors.New("查询超时")
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), statsRepo, &fakeNotificationRepo{}).(*ArticleService)

	if _, err := svc.SearchArticles("go", &GetArticleListRequest{Page: 1, PageSize: 10}); err == nil {
		t.Fatal("搜索失败应返回错误")
	}
	if len(statsRepo.searchLogs) != 1 {
		t.Fatalf("搜索日志数 = %d, 期望 1", len(statsRepo.searchLogs))
	}
	if statsRepo.searchLogs[0].Status != model.OperationStatusFailed {
		t.Errorf("日志状态 = %s, 期望 failed", statsRepo.searchLogs[0].Status)
	}
}
