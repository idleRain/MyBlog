package repository

import (
	"errors"
	"testing"

	"MyBlog/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTranslationTestRepo 创建基于 sqlmock 的分类与标签仓储，用于验证翻译写入 SQL。
func newTranslationTestRepo(t *testing.T) (*CategoryRepository, *TagRepository, sqlmock.Sqlmock) {
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

	return &CategoryRepository{db: gormDB}, &TagRepository{db: gormDB}, mock
}

// TestArticleUpsertTranslationsInsertsOnDuplicate 验证文章翻译写入携带冲突整行更新子句并回填文章ID。
func TestArticleUpsertTranslationsInsertsOnDuplicate(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始，冲突时整行更新的多行写入，事务提交。
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `article_translations`.*ON DUPLICATE KEY UPDATE").
		WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectCommit()

	translations := []model.ArticleTranslation{
		{Locale: "en", Title: "Hello", Summary: "greeting"},
		{Locale: "ja", Title: "こんにちは"},
	}
	err := repo.UpsertTranslations(7, translations)
	if err != nil {
		t.Fatalf("翻译写入不应返回错误: %v", err)
	}

	// 写入前必须回填文章ID，服务层只负责构造语言字段。
	for index, row := range translations {
		if row.ArticleID != 7 {
			t.Errorf("第 %d 行 ArticleID = %d, 期望回填为 7", index, row.ArticleID)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestArticleUpsertTranslationsRollsBackOnError 验证翻译写入失败时事务整体回滚。
func TestArticleUpsertTranslationsRollsBackOnError(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `article_translations`").
		WillReturnError(errors.New("翻译写入失败"))
	mock.ExpectRollback()

	translations := []model.ArticleTranslation{{Locale: "en", Title: "Hello"}}
	err := repo.UpsertTranslations(1, translations)
	if err == nil {
		t.Fatal("翻译写入失败时应返回错误")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望，回滚未按预期执行: %v", err)
	}
}

// TestArticleUpsertTranslationsSkipsEmpty 验证空翻译列表不开启事务直接返回。
func TestArticleUpsertTranslationsSkipsEmpty(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	if err := repo.UpsertTranslations(1, nil); err != nil {
		t.Fatalf("空翻译列表不应返回错误: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("空列表不应产生任何 SQL 调用: %v", err)
	}
}

// TestCategoryUpsertTranslationsInsertsOnDuplicate 验证分类翻译写入携带冲突整行更新子句。
func TestCategoryUpsertTranslationsInsertsOnDuplicate(t *testing.T) {
	categoryRepo, _, mock := newTranslationTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `category_translations`.*ON DUPLICATE KEY UPDATE").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := categoryRepo.UpsertTranslations(3, []model.CategoryTranslation{{Locale: "en", Name: "Tech"}})
	if err != nil {
		t.Fatalf("分类翻译写入不应返回错误: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

// TestTagUpsertTranslationsInsertsOnDuplicate 验证标签翻译写入携带冲突整行更新子句。
func TestTagUpsertTranslationsInsertsOnDuplicate(t *testing.T) {
	_, tagRepo, mock := newTranslationTestRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `tag_translations`.*ON DUPLICATE KEY UPDATE").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := tagRepo.UpsertTranslations(5, []model.TagTranslation{{Locale: "en", Name: "golang"}})
	if err != nil {
		t.Fatalf("标签翻译写入不应返回错误: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}
