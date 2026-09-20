// Package database 数据库初始化与迁移辅助。
package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// 文章全文索引的名称与建表目标列，搜索接口依赖该索引匹配标题、内容与摘要。
const articleFulltextIndexName = "ft_articles_search"

// 文章翻译表全文索引的名称与建表目标列，搜索接口依赖该索引匹配翻译后的标题、内容与摘要。
const articleTranslationFulltextIndexName = "ft_article_translations_search"

// EnsureArticleFulltextIndex 确保 articles 表存在 ngram 全文索引。
// GORM AutoMigrate 无法声明 FULLTEXT 索引，此处以幂等方式补建；
// ngram 分词器支持中文按双字切分，单字关键词受 ngram_token_size 限制无法命中。
func EnsureArticleFulltextIndex(db *gorm.DB) error {
	return ensureFulltextIndex(db, "articles", articleFulltextIndexName, "title, content, summary")
}

// EnsureArticleTranslationFulltextIndex 确保 article_translations 表存在 ngram 全文索引。
// 翻译行覆盖非默认语言的标题摘要正文，与主表索引共同支撑多语言检索。
func EnsureArticleTranslationFulltextIndex(db *gorm.DB) error {
	return ensureFulltextIndex(db, "article_translations", articleTranslationFulltextIndexName, "title, content, summary")
}

// ensureFulltextIndex 以幂等方式为目标表补建 ngram 全文索引，索引已存在时直接返回。
func ensureFulltextIndex(db *gorm.DB, table string, indexName string, columns string) error {
	var count int64
	err := db.Raw(
		"SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		table, indexName,
	).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("查询全文索引失败: %w", err)
	}
	if count > 0 {
		return nil
	}

	if err := db.Exec(
		"ALTER TABLE " + table + " ADD FULLTEXT INDEX " + indexName + " (" + columns + ") WITH PARSER ngram",
	).Error; err != nil {
		return fmt.Errorf("创建 %s 全文索引失败: %w", table, err)
	}

	log.Printf("已创建 %s 全文索引 %s（ngram）", table, indexName)
	return nil
}
