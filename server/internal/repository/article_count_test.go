package repository

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newMockArticleRepo 创建基于 sqlmock 的文章仓储，用于验证计数维护 SQL。
func newMockArticleRepo(t *testing.T) (*ArticleRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 sqlmock 失败: %v", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("创建 GORM 实例失败: %v", err)
	}

	return NewArticleRepository(gormDB).(*ArticleRepository), mock
}

// TestSyncTagsMaintainsCountSQL 验证标签同步时对新增与移除标签执行计数增减。
func TestSyncTagsMaintainsCountSQL(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始。
	mock.ExpectBegin()

	// 现有标签关联：标签 1。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_tags`")).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "tag_id"}).
			AddRow(1, 100, 1))

	// 删除现有关联。
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article_tags` WHERE article_id = ?")).
		WithArgs(100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 回滚标签 1 的使用计数。
	mock.ExpectExec("UPDATE `tags` SET `usage_count`=CASE WHEN usage_count > 0 THEN usage_count - 1 ELSE 0 END WHERE id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 新增标签 2 关联并递增计数。
	mock.ExpectExec("INSERT INTO `article_tags` \\(`article_id`,`tag_id`,`created_at`\\) VALUES \\(\\?,\\?,\\?\\)").
		WithArgs(100, 2, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("UPDATE `tags` SET `usage_count`=usage_count \\+ 1 WHERE id = \\?").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	if err := repo.SyncTags(100, []uint{2}); err != nil {
		t.Fatalf("SyncTags 失败: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestSyncCategoriesMaintainsCountSQL 验证分类同步时对新增与移除分类执行计数增减。
func TestSyncCategoriesMaintainsCountSQL(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始。
	mock.ExpectBegin()

	// 现有分类关联：分类 3。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_categories`")).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "category_id"}).
			AddRow(1, 100, 3))

	// 删除现有关联。
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article_categories` WHERE article_id = ?")).
		WithArgs(100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 回滚分类 3 的文章计数。
	mock.ExpectExec("UPDATE `categories` SET `article_count`=CASE WHEN article_count > 0 THEN article_count - 1 ELSE 0 END WHERE id = \\? AND `categories`.`deleted_at` IS NULL").
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 新增分类 4 关联并递增计数。
	mock.ExpectExec("INSERT INTO `article_categories` \\(`article_id`,`category_id`,`created_at`\\) VALUES \\(\\?,\\?,\\?\\)").
		WithArgs(100, 4, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("UPDATE `categories` SET `article_count`=article_count \\+ 1 WHERE id = \\? AND `categories`.`deleted_at` IS NULL").
		WithArgs(4).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	if err := repo.SyncCategories(100, []uint{4}); err != nil {
		t.Fatalf("SyncCategories 失败: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestDeleteRollsBackCountsSQL 验证软删文章时清理关联并回滚计数。
func TestDeleteRollsBackCountsSQL(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始。
	mock.ExpectBegin()

	// 查询文章关联的分类与标签。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_categories`")).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "category_id"}).
			AddRow(1, 100, 3))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_tags`")).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "tag_id"}).
			AddRow(1, 100, 1))

	// 删除分类与标签关联。
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article_categories` WHERE article_id = ?")).
		WithArgs(100).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article_tags` WHERE article_id = ?")).
		WithArgs(100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 回滚分类 3 与标签 1 的计数。
	mock.ExpectExec("UPDATE `categories` SET `article_count`=CASE WHEN article_count > 0 THEN article_count - 1 ELSE 0 END WHERE id = \\? AND `categories`.`deleted_at` IS NULL").
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `tags` SET `usage_count`=CASE WHEN usage_count > 0 THEN usage_count - 1 ELSE 0 END WHERE id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 软删文章本体。
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `articles` SET `deleted_at`=?")).
		WithArgs(sqlmock.AnyArg(), 100).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	if err := repo.Delete(100); err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestArchiveWritesArchivedAt 验证归档时写入状态与归档时间。
func TestArchiveWritesArchivedAt(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 归档通过事务执行更新，updated_at 由 GORM 自动维护。
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `articles` SET `archived_at`=\\?,`status`=\\?,`updated_at`=\\? WHERE id = \\? AND `articles`.`deleted_at` IS NULL").
		WithArgs(sqlmock.AnyArg(), "archived", sqlmock.AnyArg(), 100).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.Archive(100); err != nil {
		t.Fatalf("Archive 失败: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}
