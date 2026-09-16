package service

import (
	"errors"
	"testing"
	"time"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
)

// errStubView 浏览事务失败场景的桩错误。
var errStubView = errors.New("浏览事务写入失败")

// TestViewArticleRecordsDetail 验证浏览上报在单次事务调用中携带完整的明细与统计参数。
func TestViewArticleRecordsDetail(t *testing.T) {
	var recorded *model.ArticleView
	var capturedType, statType string
	var statArticleID uint
	var statDate time.Time

	repo := &fakeArticleRepo{
		recordViewWithStats: func(view *model.ArticleView, contentType string, statTypeKey string, statDateArg time.Time) error {
			recorded = view
			capturedType = contentType
			statType = statTypeKey
			statArticleID = view.ArticleID
			statDate = statDateArg
			return nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	if err := svc.ViewArticle(1, nil, "", "203.0.113.9"); err != nil {
		t.Fatalf("浏览上报失败: %v", err)
	}

	// 访客标识缺失时应以 IP 后备处理写入明细。
	if recorded == nil {
		t.Fatal("应调用浏览事务写入")
	}
	if recorded.VisitorID != "203.0.113.9" {
		t.Errorf("VisitorID = %q, 期望以 IP 后备", recorded.VisitorID)
	}
	if recorded.ViewCount != 1 {
		t.Errorf("ViewCount = %d, 期望 1", recorded.ViewCount)
	}

	// 日统计应针对当日零点的文章日浏览维度。
	if capturedType != model.ContentTypeArticle || statType != model.StatTypeDailyViews || statArticleID != 1 {
		t.Errorf("统计维度 = %s %s %d, 期望 %s %s 1", capturedType, statType, statArticleID,
			model.ContentTypeArticle, model.StatTypeDailyViews)
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if !statDate.Equal(today) {
		t.Errorf("统计日期 = %v, 期望当日零点 %v", statDate, today)
	}
}

// TestViewArticlePropagatesTransactionFailure 事务写入失败时错误应原样透传给调用方。
// 回滚语义由仓储层的 sqlmock 失败注入测试锚定。
func TestViewArticlePropagatesTransactionFailure(t *testing.T) {
	repo := &fakeArticleRepo{
		recordViewWithStats: func(*model.ArticleView, string, string, time.Time) error {
			return errStubView
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	if err := svc.ViewArticle(1, nil, "visitor-1", "203.0.113.9"); !errors.Is(err, errStubView) {
		t.Fatalf("事务失败应透传错误，实际为 %v", err)
	}
}
