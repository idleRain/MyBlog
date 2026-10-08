package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestListArchiveGroupsUsesAggregation 验证归档分组为 SQL 聚合查询而非全量列表。
func TestListArchiveGroupsUsesAggregation(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	mock.ExpectQuery("SELECT YEAR\\(published_at\\) AS year, MONTH\\(published_at\\) AS month, COUNT\\(\\*\\) AS total FROM `articles` WHERE status = \\? AND published_at IS NOT NULL AND `articles`\\.`deleted_at` IS NULL GROUP BY year, month ORDER BY year DESC, month DESC").
		WithArgs("published").
		WillReturnRows(sqlmock.NewRows([]string{"year", "month", "total"}).AddRow(2026, 3, 4))

	if _, err := repo.ListArchiveGroups(); err != nil {
		t.Fatalf("聚合归档分组查询失败: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestListArchiveRowsSelectsMinimalColumns 验证归档行仅取页面消费列，
// 且仅触发一次查询——任何 Preload 关联语句都会被未预期查询断言拦下。
func TestListArchiveRowsSelectsMinimalColumns(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT articles.id,articles.title,articles.slug,articles.summary,articles.published_at,articles.created_at FROM `articles`")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "summary", "published_at", "created_at"}))

	if _, err := repo.ListArchiveRows(); err != nil {
		t.Fatalf("归档行查询失败: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望（含未预期的关联查询）: %v", err)
	}
}

// TestListArchiveRowsFailsOnError 验证行查询失败时错误上抛。
func TestListArchiveRowsFailsOnError(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT articles.id,articles.title,articles.slug,articles.summary,articles.published_at,articles.created_at FROM `articles`")).
		WillReturnError(errors.New("模拟查询失败"))

	if _, err := repo.ListArchiveRows(); err == nil {
		t.Fatal("查询失败应返回错误")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestGetByTagsSingleQueryWithDistinct 验证标签集合为单条 IN 查询且按文章去重。
func TestGetByTagsSingleQueryWithDistinct(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// Count 与 Find 两阶段的期望：Count 由 GORM 生成 count(*)，Find 为 DISTINCT 去重。
	// 软删模型由 GORM 追加 deleted_at 条件。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `articles` JOIN article_tags ON article_tags.article_id = articles.id WHERE article_tags.tag_id IN (?,?) AND `articles`.`deleted_at` IS NULL")).
		WithArgs(uint(11), uint(12)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT DISTINCT `articles`\\..*WHERE article_tags.tag_id IN \\(\\?,\\?\\) AND `articles`\\.`deleted_at` IS NULL ORDER BY created_at DESC LIMIT \\?").
		WithArgs(uint(11), uint(12), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	if _, _, err := repo.GetByTags([]uint{11, 12}, &ArticleListParams{}); err != nil {
		t.Fatalf("标签集合查询失败: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestGetByTagsEmptyInputShortCircuits 验证空标签集合直接短路返回，不触达数据库。
func TestGetByTagsEmptyInputShortCircuits(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	articles, total, err := repo.GetByTags(nil, nil)
	if err != nil {
		t.Fatalf("空集合应成功短路: %v", err)
	}
	if articles != nil || total != 0 {
		t.Errorf("空集合应返回空结果，实际 %v / %d", articles, total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("空集合不应产生查询: %v", err)
	}
}
