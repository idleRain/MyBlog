// Package database 数据库初始化与迁移辅助。
package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// 文章全文索引的名称与建表目标列，搜索接口依赖该索引匹配标题、内容与摘要。
const articleFulltextIndexName = "ft_articles_search"

// EnsureArticleFulltextIndex 确保 articles 表存在 ngram 全文索引。
// GORM AutoMigrate 无法声明 FULLTEXT 索引，此处以幂等方式补建；
// ngram 分词器支持中文按双字切分，单字关键词受 ngram_token_size 限制无法命中。
func EnsureArticleFulltextIndex(db *gorm.DB) error {
	var count int64
	err := db.Raw(
		"SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'articles' AND index_name = ?",
		articleFulltextIndexName,
	).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("查询全文索引失败: %w", err)
	}
	if count > 0 {
		return nil
	}

	if err := db.Exec(
		"ALTER TABLE articles ADD FULLTEXT INDEX " + articleFulltextIndexName + " (title, content, summary) WITH PARSER ngram",
	).Error; err != nil {
		return fmt.Errorf("创建全文索引失败: %w", err)
	}

	log.Printf("已创建 articles 全文索引 %s（ngram）", articleFulltextIndexName)
	return nil
}
