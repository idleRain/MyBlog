package repository

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"MyBlog/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestRecordViewWithStatsRollsBackOnViewFailure 验证浏览明细写入失败时
// 计数递增与日统计整体回滚，杜绝三段写库中断导致的计数漂移。
func TestRecordViewWithStatsRollsBackOnViewFailure(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始。
	mock.ExpectBegin()

	// 浏览计数递增成功。
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `articles` SET `view_count`=view_count + 1 WHERE id = ?")).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 浏览明细：查询既有记录通过，随后插入时数据库故障。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_views` WHERE article_id = ? AND visitor_id = ? AND view_date = ?")).
		WithArgs(uint(1), "visitor-1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "visitor_id", "view_date"}))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `article_views`")).
		WillReturnError(errors.New("浏览明细写入失败"))

	// 事务回滚，撤销已执行的计数递增。
	mock.ExpectRollback()

	view := &model.ArticleView{ArticleID: 1, VisitorID: "visitor-1", ViewDate: today(), ViewCount: 1}
	err := repo.RecordViewWithStats(view, model.ContentTypeArticle, model.StatTypeDailyViews, view.ViewDate)
	if err == nil {
		t.Fatal("浏览明细写入失败时 RecordViewWithStats 应返回错误")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望，回滚未按预期执行: %v", err)
	}
}

// TestRecordViewWithStatsRollsBackOnStatFailure 验证日统计写入失败时整体回滚。
func TestRecordViewWithStatsRollsBackOnStatFailure(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始。
	mock.ExpectBegin()

	// 浏览计数递增成功。
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `articles` SET `view_count`=view_count + 1 WHERE id = ?")).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 浏览明细：无既有记录后插入成功。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_views` WHERE article_id = ? AND visitor_id = ? AND view_date = ?")).
		WithArgs(uint(1), "visitor-1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "visitor_id", "view_date"}))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `article_views`")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 日统计累加时数据库故障。
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `content_stats`")).
		WillReturnError(errors.New("日统计写入失败"))

	// 事务回滚，撤销计数递增与浏览明细。
	mock.ExpectRollback()

	view := &model.ArticleView{ArticleID: 1, VisitorID: "visitor-1", ViewDate: today(), ViewCount: 1}
	err := repo.RecordViewWithStats(view, model.ContentTypeArticle, model.StatTypeDailyViews, view.ViewDate)
	if err == nil {
		t.Fatal("日统计写入失败时 RecordViewWithStats 应返回错误")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望，回滚未按预期执行: %v", err)
	}
}

// TestRecordViewWithStatsCommitsOnSuccess 验证三段写入全部成功时事务提交。
func TestRecordViewWithStatsCommitsOnSuccess(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `articles` SET `view_count`=view_count + 1 WHERE id = ?")).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_views` WHERE article_id = ? AND visitor_id = ? AND view_date = ?")).
		WithArgs(uint(1), "visitor-1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "visitor_id", "view_date"}))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `article_views`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `content_stats`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	view := &model.ArticleView{ArticleID: 1, VisitorID: "visitor-1", ViewDate: today(), ViewCount: 1}
	if err := repo.RecordViewWithStats(view, model.ContentTypeArticle, model.StatTypeDailyViews, view.ViewDate); err != nil {
		t.Fatalf("三段写入全部成功时事务应提交: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// today 构造当日零点时间，与 service 层统计口径一致。
func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
