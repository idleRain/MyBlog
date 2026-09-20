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

// newDictTestRepo 创建基于 sqlmock 的字典仓储，用于验证字典读写 SQL。
func newDictTestRepo(t *testing.T) (*DictRepository, sqlmock.Sqlmock) {
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

	return &DictRepository{db: gormDB}, mock
}

// TestDictDeleteTypeCascadesChildrenInOrder 验证删除字典类型时按序清理翻译行、字典项与类型本体。
func TestDictDeleteTypeCascadesChildrenInOrder(t *testing.T) {
	repo, mock := newDictTestRepo(t)

	// 事务开始后先取字典项ID，随后依次删除项翻译、字典项、类型翻译与类型本体。
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `dict_items` WHERE type_id = ?")).
		WithArgs(uint(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11).AddRow(12))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `dict_item_translations` WHERE item_id IN (?,?)")).
		WithArgs(uint(11), uint(12)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `dict_items` WHERE type_id = ?")).
		WithArgs(uint(7)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `dict_type_translations` WHERE type_id = ?")).
		WithArgs(uint(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `dict_types` WHERE `dict_types`.`id` = ?")).
		WithArgs(uint(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.DeleteDictType(7); err != nil {
		t.Fatalf("删除字典类型不应返回错误: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestDictDeleteTypeRollsBackOnItemFailure 验证字典项删除失败时整体事务回滚。
func TestDictDeleteTypeRollsBackOnItemFailure(t *testing.T) {
	repo, mock := newDictTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `dict_items` WHERE type_id = ?")).
		WithArgs(uint(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `dict_item_translations` WHERE item_id IN (?)")).
		WithArgs(uint(11)).
		WillReturnError(errors.New("删除失败"))
	mock.ExpectRollback()

	if err := repo.DeleteDictType(7); err == nil {
		t.Fatal("字典项删除失败时应返回错误")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestDictUpsertTypeTranslationsBackfillsTypeID 验证类型翻译写入前回填类型ID并携带冲突更新子句。
func TestDictUpsertTypeTranslationsBackfillsTypeID(t *testing.T) {
	repo, mock := newDictTestRepo(t)

	// 事务开始，翻译行命中唯一索引时整行更新。
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `dict_type_translations`.*ON DUPLICATE KEY UPDATE").
		WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectCommit()

	translations := []model.DictTypeTranslation{
		{Locale: "en", Name: "Tag Status"},
	}
	if err := repo.UpsertTypeTranslations(7, translations); err != nil {
		t.Fatalf("写入字典类型翻译不应返回错误: %v", err)
	}

	if translations[0].TypeID != 7 {
		t.Errorf("TypeID = %d, 期望回填为 7", translations[0].TypeID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestDictUpsertItemTranslationsBackfillsItemID 验证字典项翻译写入前回填字典项ID。
func TestDictUpsertItemTranslationsBackfillsItemID(t *testing.T) {
	repo, mock := newDictTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `dict_item_translations`.*ON DUPLICATE KEY UPDATE").
		WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectCommit()

	translations := []model.DictItemTranslation{
		{Locale: "en", Label: "Enabled"},
	}
	if err := repo.UpsertItemTranslations(9, translations); err != nil {
		t.Fatalf("写入字典项翻译不应返回错误: %v", err)
	}

	if translations[0].ItemID != 9 {
		t.Errorf("ItemID = %d, 期望回填为 9", translations[0].ItemID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestDictListItemsFiltersByTypeAndStatus 验证字典项列表按类型与状态过滤。
func TestDictListItemsFiltersByTypeAndStatus(t *testing.T) {
	repo, mock := newDictTestRepo(t)

	typeID := uint(7)
	status := model.DictStatusEnabled

	// 计数查询与列表查询均携带相同的过滤条件。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `dict_items` WHERE type_id = ? AND status = ?")).
		WithArgs(typeID, status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `dict_items` WHERE type_id = ? AND status = ? ORDER BY sort_order ASC, id ASC LIMIT ?")).
		WithArgs(typeID, status, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type_id", "value", "label"}).AddRow(1, 7, "1", "启用"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `dict_item_translations` WHERE `dict_item_translations`.`item_id` = ?")).
		WithArgs(uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "locale"}))

	items, total, err := repo.ListDictItems(&DictItemListParams{
		TypeID:   &typeID,
		Status:   &status,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("查询字典项列表不应返回错误: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("期望返回 1 条记录, 实际 total=%d len=%d", total, len(items))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestDictListEnabledItemsByTypesSkipsEmptyIDs 验证空类型集合不产生数据库查询。
func TestDictListEnabledItemsByTypesSkipsEmptyIDs(t *testing.T) {
	repo, mock := newDictTestRepo(t)

	items, err := repo.ListEnabledItemsByTypes(nil)
	if err != nil {
		t.Fatalf("空类型集合查询不应返回错误: %v", err)
	}
	if items != nil {
		t.Fatalf("空类型集合应返回 nil, 实际 %v", items)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}
