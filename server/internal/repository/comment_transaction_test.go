package repository

import (
	"errors"
	"regexp"
	"testing"

	"MyBlog/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newMockCommentRepo 创建基于 sqlmock 的评论仓储，用于验证评论与关联计数的同事务落库。
func newMockCommentRepo(t *testing.T) (CommentRepositoryInterface, sqlmock.Sqlmock) {
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

	return NewCommentRepository(gormDB), mock
}

// expectArticleCommentCountSync 登记文章评论数按关联评论重算的期望语句。
// 文章与评论均为软删模型，GORM 自动追加 deleted_at 条件。
func expectArticleCommentCountSync(mock sqlmock.Sqlmock, articleID uint) {
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `articles` SET `comment_count`=(SELECT COUNT(*) FROM comments WHERE article_id = ? AND deleted_at IS NULL) WHERE id = ? AND `articles`.`deleted_at` IS NULL")).
		WithArgs(articleID, articleID).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// expectCommentSoftDelete 登记评论软删除的期望语句，软删除以写入 deleted_at 实现。
func expectCommentSoftDelete(mock sqlmock.Sqlmock, commentID uint) {
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `comments` SET `deleted_at`=? WHERE `comments`.`id` = ? AND `comments`.`deleted_at` IS NULL")).
		WithArgs(sqlmock.AnyArg(), commentID).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// TestCreateWithCountsSingleTransaction 验证创建评论时评论落库与两路计数维护在同一事务提交。
func TestCreateWithCountsSingleTransaction(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	rootID := uint(100)
	comment := testCommentForTransaction(1, &rootID)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `comments`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// 回复评论时父评论回复数递增与文章评论数重算都在同一事务内。
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `comments` SET `reply_count`=reply_count + 1 WHERE id = ?")).
		WithArgs(rootID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectArticleCommentCountSync(mock, comment.ArticleID)
	mock.ExpectCommit()

	if err := repo.CreateWithCounts(comment); err != nil {
		t.Fatalf("事务化创建评论失败: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("创建事务未按期望执行: %v", err)
	}
}

// TestCreateWithCountsRollsBackWhenCountSyncFails 验证计数维护失败时评论创建整体回滚。
func TestCreateWithCountsRollsBackWhenCountSyncFails(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	comment := testCommentForTransaction(1, nil)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `comments`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// 文章评论数重算失败，评论记录必须随事务回滚，不留下无计数的评论。
	mock.ExpectExec("UPDATE `articles` SET `comment_count`").
		WillReturnError(errors.New("模拟评论计数重算失败"))
	mock.ExpectRollback()

	if err := repo.CreateWithCounts(comment); err == nil {
		t.Fatal("计数重算失败应使创建整体失败")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("回滚事务未按期望执行: %v", err)
	}
}

// TestDeleteWithCountsSingleTransaction 验证删除评论时软删与两路计数回退在同一事务提交。
func TestDeleteWithCountsSingleTransaction(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	rootID := uint(100)
	comment := testCommentForTransaction(1, &rootID)

	mock.ExpectBegin()
	expectCommentSoftDelete(mock, 1)
	// 回退父评论回复数，CASE 条件保证不产生负值。
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `comments` SET `reply_count`=CASE WHEN reply_count > 0 THEN reply_count - 1 ELSE 0 END WHERE id = ?")).
		WithArgs(rootID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectArticleCommentCountSync(mock, comment.ArticleID)
	mock.ExpectCommit()

	if err := repo.DeleteWithCounts(comment); err != nil {
		t.Fatalf("事务化删除评论失败: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("删除事务未按期望执行: %v", err)
	}
}

// TestDeleteWithCountsRollsBackWhenCountSyncFails 验证计数回退失败时删除整体回滚。
func TestDeleteWithCountsRollsBackWhenCountSyncFails(t *testing.T) {
	repo, mock := newMockCommentRepo(t)
	comment := testCommentForTransaction(1, nil)

	mock.ExpectBegin()
	expectCommentSoftDelete(mock, 1)
	mock.ExpectExec("UPDATE `articles` SET `comment_count`").
		WillReturnError(errors.New("模拟评论计数重算失败"))
	mock.ExpectRollback()

	if err := repo.DeleteWithCounts(comment); err == nil {
		t.Fatal("计数重算失败应使删除整体失败")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("回滚事务未按期望执行: %v", err)
	}
}

// testCommentForTransaction 构造事务用例使用的评论实体，rootID 非空时模拟回复形态。
// 服务层对二级回复统一回挂根评论，回复形态下 ParentID 与 RootID 同值。
func testCommentForTransaction(articleID uint, rootID *uint) *model.Comment {
	comment := &model.Comment{ArticleID: articleID, ID: 1}
	if rootID != nil {
		comment.RootID = rootID
		comment.ParentID = rootID
	}
	return comment
}
