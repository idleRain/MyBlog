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

// newMockSettingRepo 创建基于 sqlmock 的设置仓储，用于验证批量写回事务边界。
func newMockSettingRepo(t *testing.T) (SettingRepositoryInterface, sqlmock.Sqlmock) {
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

	return NewSettingRepository(gormDB), mock
}

// settingBatchForTest 构造两键批量写回条目，供事务边界用例复用。
func settingBatchForTest() []*model.Setting {
	return []*model.Setting{
		{KeyName: "site_name", Value: "新站名"},
		{KeyName: "site_desc", Value: "新站描述"},
	}
}

// expectSettingUpdate 登记一项按键名读取匹配行的写回期望。
// UPDATE 语句仅锚定表名前缀，列数由 FirstOrCreate 全字段写回决定，
// 本用例的锚点是事务边界而非列明细。
func expectSettingUpdate(mock sqlmock.Sqlmock, keyName string, execError error) {
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `settings` WHERE key_name = ? ORDER BY `settings`.`id` LIMIT ?")).
		WithArgs(keyName, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key_name"}).AddRow(1, keyName))
	exec := mock.ExpectExec(regexp.QuoteMeta("UPDATE `settings` SET"))
	if execError != nil {
		exec.WillReturnError(execError)
		return
	}
	exec.WillReturnResult(sqlmock.NewResult(0, 1))
}

// TestUpsertBatchRollsBackOnMidItemFailure 验证批量写回中途失败时整体回滚。
func TestUpsertBatchRollsBackOnMidItemFailure(t *testing.T) {
	repo, mock := newMockSettingRepo(t)

	items := settingBatchForTest()

	mock.ExpectBegin()
	expectSettingUpdate(mock, items[0].KeyName, nil)
	// 第二项写库失败，事务应整体回滚而非部分提交。
	expectSettingUpdate(mock, items[1].KeyName, errors.New("模拟第二项写库失败"))
	mock.ExpectRollback()

	if err := repo.UpsertBatch(items); err == nil {
		t.Fatal("中途失败应返回错误")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("回滚路径未满足: %v", err)
	}
}

// TestUpsertBatchCommitsAllOnSuccess 验证全部成功时批内各项同事务提交。
func TestUpsertBatchCommitsAllOnSuccess(t *testing.T) {
	repo, mock := newMockSettingRepo(t)

	items := settingBatchForTest()

	mock.ExpectBegin()
	for _, setting := range items {
		expectSettingUpdate(mock, setting.KeyName, nil)
	}
	mock.ExpectCommit()

	if err := repo.UpsertBatch(items); err != nil {
		t.Fatalf("批量写回失败: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("提交路径未满足: %v", err)
	}
}
