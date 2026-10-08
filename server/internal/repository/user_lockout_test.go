package repository

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newMockUserRepo 创建基于 sqlmock 的用户仓储，用于验证登录计数维护 SQL。
func newMockUserRepo(t *testing.T) (UserRepository, sqlmock.Sqlmock) {
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

	return NewUserRepository(gormDB), mock
}

// TestIncrementLoginFailuresUsesAtomicUpdate 验证失败计数为数据库侧自增语句，
// 并发登录各自生效，且更新语句不携带其他字段，防全字段覆盖回归。
func TestIncrementLoginFailuresUsesAtomicUpdate(t *testing.T) {
	repo, mock := newMockUserRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `failed_login_count`=failed_login_count + 1 WHERE id = ?")).
		WithArgs(uint(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.IncrementLoginFailures(7); err != nil {
		t.Fatalf("累计失败计数失败: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestLockUserAfterFailuresConditionalUpdate 验证锁定写入为阈值条件更新并透传命中结果。
func TestLockUserAfterFailuresConditionalUpdate(t *testing.T) {
	repo, mock := newMockUserRepo(t)

	lockedUntil := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)

	// 计数达阈值：条件命中，返回 true。
	// 软删除模型由 GORM 自动追加 deleted_at 条件，期望语句按生成实况登记。
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `locked_until`=? WHERE (id = ? AND failed_login_count >= ?) AND `users`.`deleted_at` IS NULL")).
		WithArgs(lockedUntil, uint(7), 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if locked, err := repo.LockUserAfterFailures(7, 5, lockedUntil); err != nil || !locked {
		t.Fatalf("达阈值的锁定写入应命中，locked=%v err=%v", locked, err)
	}

	// 计数未达阈值：条件未命中，返回 false。
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `locked_until`=? WHERE (id = ? AND failed_login_count >= ?) AND `users`.`deleted_at` IS NULL")).
		WithArgs(lockedUntil, uint(7), 5).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	if locked, err := repo.LockUserAfterFailures(7, 5, lockedUntil); err != nil || locked {
		t.Fatalf("未达阈值的锁定写入不应命中，locked=%v err=%v", locked, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestResetLoginFailuresUpdatesOnlyLoginColumns 验证清零为单条原子更新且仅触碰登录相关列。
// 期望 SQL 为整句精确匹配，子句中不出现昵称、邮箱等无关列即同时锚定防全字段覆盖的验收条款。
func TestResetLoginFailuresUpdatesOnlyLoginColumns(t *testing.T) {
	repo, mock := newMockUserRepo(t)

	// Select 强制更新两列，零值为空指针的列也应写 NULL 而非被跳过。
	// Updates 语义下 GORM 自动维护 updated_at 并追加软删除条件。
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `failed_login_count`=?,`locked_until`=?,`updated_at`=? WHERE id = ? AND `users`.`deleted_at` IS NULL")).
		WithArgs(uint(0), nil, sqlmock.AnyArg(), uint(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.ResetLoginFailures(7); err != nil {
		t.Fatalf("清零失败计数失败: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}
