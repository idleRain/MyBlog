package service

import (
	"testing"
	"time"

	"MyBlog/internal/model"
)

// archiveArticle 构造指定发布时间的已发布文章。
func archiveArticle(id uint, publishedAt time.Time) *model.Article {
	return &model.Article{
		ID:          id,
		Title:       "归档文章",
		Status:      model.ArticleStatusPublished,
		PublishedAt: &publishedAt,
	}
}

// TestGroupArticlesByYearMonth 验证文章按年月两级分组并统计年总量。
func TestGroupArticlesByYearMonth(t *testing.T) {
	articles := []*model.Article{
		archiveArticle(1, time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC)),
		archiveArticle(2, time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC)),
		archiveArticle(3, time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)),
		archiveArticle(4, time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC)),
	}

	groups := groupArticlesByYearMonth(articles)

	if len(groups) != 2 {
		t.Fatalf("年份分组数 = %d, 期望 2", len(groups))
	}

	first := groups[0]
	if first.Year != 2026 || first.Total != 3 {
		t.Errorf("首组 = %d 年 %d 篇, 期望 2026 年 3 篇", first.Year, first.Total)
	}
	if len(first.Months) != 2 {
		t.Fatalf("2026 年月份分组数 = %d, 期望 2", len(first.Months))
	}
	if first.Months[0].Month != 2 || len(first.Months[0].Articles) != 1 {
		t.Errorf("2 月分组 = %d 月 %d 篇, 期望 2 月 1 篇", first.Months[0].Month, len(first.Months[0].Articles))
	}
	if first.Months[1].Month != 1 || len(first.Months[1].Articles) != 2 {
		t.Errorf("1 月分组 = %d 月 %d 篇, 期望 1 月 2 篇", first.Months[1].Month, len(first.Months[1].Articles))
	}

	second := groups[1]
	if second.Year != 2025 || second.Total != 1 {
		t.Errorf("次组 = %d 年 %d 篇, 期望 2025 年 1 篇", second.Year, second.Total)
	}
}

// TestGroupArticlesSkipsMissingPublishedAt 验证缺失发布时间的记录不进入归档分组。
func TestGroupArticlesSkipsMissingPublishedAt(t *testing.T) {
	articles := []*model.Article{
		{ID: 1, Title: "未发布文章", Status: model.ArticleStatusDraft},
		archiveArticle(2, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)),
	}

	groups := groupArticlesByYearMonth(articles)

	if len(groups) != 1 {
		t.Fatalf("年份分组数 = %d, 期望 1", len(groups))
	}
	if groups[0].Total != 1 {
		t.Errorf("2026 年总量 = %d, 期望 1", groups[0].Total)
	}
}
