package service

import (
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// TestIsArticleLiked 验证点赞状态查询透传仓储结果。
func TestIsArticleLiked(t *testing.T) {
	repo := &fakeArticleRepo{
		existsLike: func(articleID, userID uint) (bool, error) {
			return articleID == 1 && userID == 1, nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	isLiked, err := svc.IsArticleLiked(1, 1)
	if err != nil {
		t.Fatalf("查询点赞状态失败: %v", err)
	}
	if !isLiked {
		t.Error("点赞状态应为 true")
	}
}

// TestIsArticleBookmarked 验证收藏状态查询透传仓储结果。
func TestIsArticleBookmarked(t *testing.T) {
	repo := &fakeArticleRepo{
		existsBookmark: func(articleID, userID uint) (bool, error) {
			return false, nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	isBookmarked, err := svc.IsArticleBookmarked(1, 1)
	if err != nil {
		t.Fatalf("查询收藏状态失败: %v", err)
	}
	if isBookmarked {
		t.Error("收藏状态应为 false")
	}
}

// TestGetArticleBookmarksAppliesPaginationDefaults 验证收藏列表的分页默认值。
func TestGetArticleBookmarksAppliesPaginationDefaults(t *testing.T) {
	var captured *repository.ArticleListParams
	repo := &fakeArticleRepo{
		listBookmarks: func(userID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			captured = params
			return []*model.Article{publishedArticle(1)}, 1, nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	result, err := svc.GetArticleBookmarks(1, &GetArticleListRequest{})
	if err != nil {
		t.Fatalf("查询收藏列表失败: %v", err)
	}
	if captured.Page != 1 || captured.PageSize != 10 {
		t.Errorf("分页参数 = %d/%d, 期望 1/10", captured.Page, captured.PageSize)
	}
	if result.Total != 1 || len(result.Articles) != 1 {
		t.Errorf("收藏列表 = %d 篇 (total %d), 期望 1 篇", len(result.Articles), result.Total)
	}
}
