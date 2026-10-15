package repository

import (
	"errors"
	"regexp"
	"testing"

	"MyBlog/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// mediaHashQuery 为按哈希与上传者查询单行的期望语句，取自实现实际生成的 SQL。
// 语句锚定 uploader_id 条件，条件被移除时期望不再匹配。
const mediaHashQuery = "SELECT * FROM `media_files` WHERE file_hash = ? AND uploader_id = ? AND `media_files`.`deleted_at` IS NULL ORDER BY `media_files`.`id` LIMIT ?"

// newMockMediaRepo 创建基于 sqlmock 的媒体仓储，用于锁定哈希查询的上传者限定条件。
func newMockMediaRepo(t *testing.T) (MediaRepositoryInterface, sqlmock.Sqlmock) {
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

	return NewMediaRepository(gormDB), mock
}

// expectMediaQueryByHashAndUploader 登记一项按哈希与上传者查询单行的期望。
// 末位参数为 First 追加的 LIMIT 值。
func expectMediaQueryByHashAndUploader(mock sqlmock.Sqlmock, hash string, uploaderID uint) *sqlmock.ExpectedQuery {
	return mock.ExpectQuery(regexp.QuoteMeta(mediaHashQuery)).WithArgs(hash, uploaderID, 1)
}

// TestGetByFileHashAndUploaderReturnsOwnRecord 验证哈希查询按上传者限定并回出命中行。
func TestGetByFileHashAndUploaderReturnsOwnRecord(t *testing.T) {
	repo, mock := newMockMediaRepo(t)

	const uploaderID uint = 2
	rows := sqlmock.NewRows([]string{"id", "filename", "file_path", "file_url", "file_hash", "uploader_id"}).
		AddRow(2, "mine.png", "/srv/uploads/2026/01/mine.png", "/uploads/2026/01/mine.png", "hash-value", uploaderID)
	expectMediaQueryByHashAndUploader(mock, "hash-value", uploaderID).WillReturnRows(rows)

	media, err := repo.GetByFileHashAndUploader("hash-value", uploaderID)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if media == nil {
		t.Fatal("命中时应返回实体")
	}
	if media.ID != 2 {
		t.Errorf("返回记录 ID 不符合预期，实际 %d", media.ID)
	}
	if media.UploaderID != uploaderID {
		t.Errorf("返回记录应归属于查询的上传者，实际 uploaderId = %d", media.UploaderID)
	}
	if media.Filename != "mine.png" || media.FilePath != "/srv/uploads/2026/01/mine.png" {
		t.Errorf("返回记录字段不符合预期: filename=%s filePath=%s", media.Filename, media.FilePath)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("查询期望未满足: %v", err)
	}
}

// TestGetByFileHashAndUploaderNotFound 验证未命中时返回不存在的哨兵错误。
func TestGetByFileHashAndUploaderNotFound(t *testing.T) {
	repo, mock := newMockMediaRepo(t)

	const uploaderID uint = 2
	// 空结果集模拟数据库无匹配行，GORM 据此返回 ErrRecordNotFound。
	emptyRows := sqlmock.NewRows([]string{"id", "filename", "file_path", "file_url", "file_hash", "uploader_id"})
	expectMediaQueryByHashAndUploader(mock, "absent-hash", uploaderID).WillReturnRows(emptyRows)

	if _, err := repo.GetByFileHashAndUploader("absent-hash", uploaderID); !errors.Is(err, domain.ErrMediaNotFound) {
		t.Errorf("未命中应返回 domain.ErrMediaNotFound，实际 %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("查询期望未满足: %v", err)
	}
}
