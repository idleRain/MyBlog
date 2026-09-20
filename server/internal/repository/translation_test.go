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

// TestCreateWithRelationsUpsertsTranslations 验证创建文章事务内写入翻译行并回填文章ID。
func TestCreateWithRelationsUpsertsTranslations(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始，slug 唯一性检查通过，文章写入成功。
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `articles` WHERE slug = ?")).
		WithArgs("test-article").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `articles`")).
		WillReturnResult(sqlmock.NewResult(7, 1))

	// 翻译行随主事务写入，携带冲突整行更新子句。
	mock.ExpectExec("INSERT INTO `article_translations`.*ON DUPLICATE KEY UPDATE").
		WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectCommit()

	translations := []model.ArticleTranslation{
		{Locale: "en", Title: "Hello", Summary: "greeting"},
		{Locale: "ja", Title: "こんにちは"},
	}
	article := &model.Article{Title: "测试文章", Slug: "test-article", AuthorID: 1, Status: model.ArticleStatusDraft}
	err := repo.CreateWithRelations(article, nil, nil, translations)
	if err != nil {
		t.Fatalf("创建文章不应返回错误: %v", err)
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

// TestUpdateWithRelationsRollsBackOnTranslationFailure 验证翻译行写入失败时更新事务整体回滚。
func TestUpdateWithRelationsRollsBackOnTranslationFailure(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始，slug 检查与文章更新成功。
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `articles` WHERE slug = ? AND id != ?")).
		WithArgs("test-article", uint(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `articles`")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 翻译行写入失败，事务回滚撤销文章更新。
	mock.ExpectExec("INSERT INTO `article_translations`").
		WillReturnError(errors.New("翻译写入失败"))
	mock.ExpectRollback()

	translations := []model.ArticleTranslation{{Locale: "en", Title: "Hello"}}
	article := &model.Article{ID: 7, Title: "测试文章", Slug: "test-article", AuthorID: 1, Status: model.ArticleStatusDraft}
	err := repo.UpdateWithRelations(article, nil, nil, translations)
	if err == nil {
		t.Fatal("翻译写入失败时更新文章应返回错误")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望，回滚未按预期执行: %v", err)
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
