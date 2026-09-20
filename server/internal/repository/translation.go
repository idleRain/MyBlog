// Package repository 数据访问层
package repository

import (
	"MyBlog/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 内容多语言翻译写入实现。
// 翻译行的字段合并、语言校验与派生值计算由服务层完成，本层只负责原子写入。

// UpsertTranslations 批量写入文章翻译行，命中唯一索引时整行更新。
// 单事务包裹保证多语言写入的原子性。
func (r *ArticleRepository) UpsertTranslations(articleID uint, translations []model.ArticleTranslation) error {
	if len(translations) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return upsertArticleTranslationRows(tx, articleID, translations)
	})
}

// upsertArticleTranslationRows 在给定句柄内写入文章翻译行，写入前回填文章ID。
func upsertArticleTranslationRows(tx *gorm.DB, articleID uint, translations []model.ArticleTranslation) error {
	for index := range translations {
		translations[index].ArticleID = articleID
	}
	return upsertTranslationRows(tx, &translations)
}

// UpsertTranslations 批量写入分类翻译行，命中唯一索引时整行更新。
// 单事务包裹保证多语言写入的原子性。
func (r *CategoryRepository) UpsertTranslations(categoryID uint, translations []model.CategoryTranslation) error {
	if len(translations) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return upsertCategoryTranslationRows(tx, categoryID, translations)
	})
}

// upsertCategoryTranslationRows 在给定句柄内写入分类翻译行，写入前回填分类ID。
func upsertCategoryTranslationRows(tx *gorm.DB, categoryID uint, translations []model.CategoryTranslation) error {
	for index := range translations {
		translations[index].CategoryID = categoryID
	}
	return upsertTranslationRows(tx, &translations)
}

// UpsertTranslations 批量写入标签翻译行，命中唯一索引时整行更新。
// 单事务包裹保证多语言写入的原子性。
func (r *TagRepository) UpsertTranslations(tagID uint, translations []model.TagTranslation) error {
	if len(translations) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return upsertTagTranslationRows(tx, tagID, translations)
	})
}

// upsertTagTranslationRows 在给定句柄内写入标签翻译行，写入前回填标签ID。
func upsertTagTranslationRows(tx *gorm.DB, tagID uint, translations []model.TagTranslation) error {
	for index := range translations {
		translations[index].TagID = tagID
	}
	return upsertTranslationRows(tx, &translations)
}

// upsertTranslationRows 以整行更新语义写入翻译行，供三个内容实体的翻译写入复用。
// 命中实体ID与 locale 组成的唯一索引时更新全部内容列，新行正常插入。
func upsertTranslationRows(tx *gorm.DB, rows interface{}) error {
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(rows).Error
}
