package service

import (
	"testing"
	"time"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeStatsRepo 站点统计仓储的测试替身。
type fakeStatsRepo struct {
	repository.StatsRepositoryInterface
	articleCount  int64
	published     int64
	totalViews    uint64
	totalLikes    uint64
	commentCount  int64
	userCount     int64
	categoryCount int64
	tagCount      int64
	contentStats  []*model.ContentStats
}

func (f *fakeStatsRepo) CountArticles(status model.ArticleStatus) (int64, error) {
	if status == model.ArticleStatusPublished {
		return f.published, nil
	}
	return f.articleCount, nil
}

func (f *fakeStatsRepo) SumArticleViews() (uint64, error) { return f.totalViews, nil }
func (f *fakeStatsRepo) SumArticleLikes() (uint64, error) { return f.totalLikes, nil }
func (f *fakeStatsRepo) CountComments() (int64, error)    { return f.commentCount, nil }
func (f *fakeStatsRepo) CountUsers() (int64, error)       { return f.userCount, nil }
func (f *fakeStatsRepo) CountCategories() (int64, error)  { return f.categoryCount, nil }
func (f *fakeStatsRepo) CountTags() (int64, error)        { return f.tagCount, nil }

func (f *fakeStatsRepo) GetContentStats(contentType, statType string, startDate time.Time) ([]*model.ContentStats, error) {
	return f.contentStats, nil
}

// TestGetOverview 验证站点统计概览聚合各维度数据。
func TestGetOverview(t *testing.T) {
	repo := &fakeStatsRepo{
		articleCount:  8,
		published:     5,
		totalViews:    1000,
		totalLikes:    200,
		commentCount:  50,
		userCount:     10,
		categoryCount: 4,
		tagCount:      5,
	}
	svc := NewStatsService(repo)

	overview, err := svc.GetOverview()
	if err != nil {
		t.Fatalf("获取统计概览失败: %v", err)
	}

	if overview.ArticleCount != 8 || overview.PublishedCount != 5 {
		t.Errorf("文章统计错误: total=%d published=%d", overview.ArticleCount, overview.PublishedCount)
	}
	if overview.TotalViews != 1000 || overview.TotalLikes != 200 {
		t.Errorf("浏览量/点赞统计错误: views=%d likes=%d", overview.TotalViews, overview.TotalLikes)
	}
	if overview.CommentCount != 50 || overview.UserCount != 10 {
		t.Errorf("评论/用户统计错误: comments=%d users=%d", overview.CommentCount, overview.UserCount)
	}
}

// TestGetArticleViewsTrend 验证浏览量趋势按日期补零。
func TestGetArticleViewsTrend(t *testing.T) {
	repo := &fakeStatsRepo{
		contentStats: []*model.ContentStats{
			{
				ContentType: model.ContentTypeArticle,
				StatType:    model.StatTypeDailyViews,
				StatValue:   42,
				StatDate:    time.Now().AddDate(0, 0, -1),
			},
		},
	}
	svc := NewStatsService(repo)

	trend, err := svc.GetArticleViewsTrend(7)
	if err != nil {
		t.Fatalf("获取浏览量趋势失败: %v", err)
	}

	if len(trend.Dates) != 7 || len(trend.Values) != 7 {
		t.Fatalf("趋势长度 = %d/%d, 期望 7", len(trend.Dates), len(trend.Values))
	}

	// 校验前一天的浏览量已填入。
	for i, date := range trend.Dates {
		if date == time.Now().AddDate(0, 0, -1).Format("2006-01-02") {
			if trend.Values[i] != 42 {
				t.Errorf("前一天浏览量 = %d, 期望 42", trend.Values[i])
			}
		}
	}
}
