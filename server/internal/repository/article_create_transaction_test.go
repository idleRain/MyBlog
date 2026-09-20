package repository

import (
	"errors"
	"regexp"
	"testing"

	"MyBlog/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestCreateWithRelationsRollsBackOnCategoryFailure 验证创建文章时分类关联同步失败会整体回滚。
// 回滚语义锚定：文章写入与分类关联写入在同一事务内，任一环节失败即撤销全部写入。
func TestCreateWithRelationsRollsBackOnCategoryFailure(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始。
	mock.ExpectBegin()

	// 创建前的 slug 唯一性检查，返回 0 表示未占用。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `articles` WHERE slug = ?")).
		WithArgs("test-article").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// 文章写入成功。
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `articles`")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 分类同步：查询现有关联为空，删除后插入新关联时数据库故障。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_categories` WHERE article_id = ?")).
		WithArgs(uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "category_id"}))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article_categories` WHERE article_id = ?")).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `article_categories`")).
		WillReturnError(errors.New("分类关联写入失败"))

	// 事务回滚，撤销文章与关联的全部写入。
	mock.ExpectRollback()

	article := &model.Article{Title: "测试文章", Slug: "test-article", AuthorID: 1, Status: model.ArticleStatusDraft}
	err := repo.CreateWithRelations(article, []uint{1}, nil, nil)
	if err == nil {
		t.Fatal("分类同步失败时 CreateWithRelations 应返回错误")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望，回滚未按预期执行: %v", err)
	}
}

// TestCreateWithRelationsRollsBackOnTagFailure 验证创建文章时标签关联同步失败会整体回滚。
func TestCreateWithRelationsRollsBackOnTagFailure(t *testing.T) {
	repo, mock := newMockArticleRepo(t)

	// 事务开始。
	mock.ExpectBegin()

	// slug 唯一性检查通过。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `articles` WHERE slug = ?")).
		WithArgs("test-article").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// 文章写入成功。
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `articles`")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 分类同步成功：查询空关联后删除并写入新关联，随后递增分类文章计数。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_categories` WHERE article_id = ?")).
		WithArgs(uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "category_id"}))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article_categories` WHERE article_id = ?")).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `article_categories`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `categories` SET `article_count`=article_count + 1 WHERE id = ?")).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 标签同步：查询空关联后删除，插入新关联时数据库故障。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `article_tags` WHERE article_id = ?")).
		WithArgs(uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "article_id", "tag_id"}))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `article_tags` WHERE article_id = ?")).
		WithArgs(uint(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `article_tags`")).
		WillReturnError(errors.New("标签关联写入失败"))

	// 事务回滚，撤销文章与已写入的分类关联。
	mock.ExpectRollback()

	article := &model.Article{Title: "测试文章", Slug: "test-article", AuthorID: 1, Status: model.ArticleStatusDraft}
	err := repo.CreateWithRelations(article, []uint{1}, []uint{1}, nil)
	if err == nil {
		t.Fatal("标签同步失败时 CreateWithRelations 应返回错误")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望，回滚未按预期执行: %v", err)
	}
}
