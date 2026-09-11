package service

import (
	"errors"
	"testing"
	"time"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
)

// errStubView 浏览计数失败场景的桩错误。
var errStubView = errors.New("递增浏览量失败")

// TestViewArticleRecordsDetail 验证浏览上报落访客明细并累加日统计。
func TestViewArticleRecordsDetail(t *testing.T) {
	var recorded *model.ArticleView
	var statType string
	var statArticleID uint
	var statDate time.Time

	repo := &fakeArticleRepo{
		recordView: func(view *model.ArticleView) error {
			recorded = view
			return nil
		},
	}
	statsRepo := &fakeStatsRepo{
		upsertContentStat: func(contentType string, contentID uint, statKey string, date time.Time) error {
			statType = statKey
			statArticleID = contentID
			statDate = date
			return nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), statsRepo, &fakeNotificationRepo{}).(*ArticleService)

	if err := svc.ViewArticle(1, nil, "", "203.0.113.9"); err != nil {
		t.Fatalf("浏览上报失败: %v", err)
	}

	// 访客标识缺失时应以 IP 兜底写入明细。
	if recorded == nil {
		t.Fatal("应写入浏览明细")
	}
	if recorded.VisitorID != "203.0.113.9" {
		t.Errorf("VisitorID = %q, 期望以 IP 兜底", recorded.VisitorID)
	}
	if recorded.ViewCount != 1 {
		t.Errorf("ViewCount = %d, 期望 1", recorded.ViewCount)
	}

	// 日统计应针对当日零点的文章日浏览维度。
	if statType != model.StatTypeDailyViews || statArticleID != 1 {
		t.Errorf("日统计参数 = %s %d, 期望 %s 1", statType, statArticleID, model.StatTypeDailyViews)
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if !statDate.Equal(today) {
		t.Errorf("统计日期 = %v, 期望当日零点 %v", statDate, today)
	}
}

// TestViewArticleRejectsIncrementFailure 验证计数递增失败时不再写明细。
func TestViewArticleRejectsIncrementFailure(t *testing.T) {
	repo := &fakeArticleRepo{
		incrementView: func(id uint) error {
			return errStubView
		},
		recordView: func(view *model.ArticleView) error {
			t.Error("计数失败后不应写浏览明细")
			return nil
		},
	}
	statsRepo := &fakeStatsRepo{
		upsertContentStat: func(contentType string, contentID uint, statType string, statDate time.Time) error {
			t.Error("计数失败后不应写日统计")
			return nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), statsRepo, &fakeNotificationRepo{}).(*ArticleService)

	if err := svc.ViewArticle(1, nil, "visitor-1", "203.0.113.9"); err == nil {
		t.Fatal("计数递增失败应返回错误")
	}
}
