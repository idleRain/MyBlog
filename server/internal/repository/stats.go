// Package repository 数据访问层
package repository

import (
	"time"

	"MyBlog/internal/model"

	"gorm.io/gorm"
)

// StatsRepositoryInterface 站点统计仓储接口
type StatsRepositoryInterface interface {
	// 文章维度统计
	CountArticles(status model.ArticleStatus) (int64, error)
	SumArticleViews() (uint64, error)
	SumArticleLikes() (uint64, error)

	// 内容聚合统计
	CountComments() (int64, error)
	CountUsers() (int64, error)
	CountCategories() (int64, error)
	CountTags() (int64, error)

	// 时间维度统计
	GetContentStats(contentType, statType string, startDate time.Time) ([]*model.ContentStats, error)
}

// StatsRepository 站点统计仓储实现
type StatsRepository struct {
	db *gorm.DB
}

// NewStatsRepository 创建站点统计仓储实例
func NewStatsRepository(db *gorm.DB) StatsRepositoryInterface {
	return &StatsRepository{db: db}
}

// CountArticles 统计指定状态的文章数量。
func (r *StatsRepository) CountArticles(status model.ArticleStatus) (int64, error) {
	var count int64
	query := r.db.Model(&model.Article{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// SumArticleViews 统计文章总浏览量。
func (r *StatsRepository) SumArticleViews() (uint64, error) {
	var total uint64
	if err := r.db.Model(&model.Article{}).Select("COALESCE(SUM(view_count), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// SumArticleLikes 统计文章总点赞数。
func (r *StatsRepository) SumArticleLikes() (uint64, error) {
	var total uint64
	if err := r.db.Model(&model.Article{}).Select("COALESCE(SUM(like_count), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountComments 统计评论总数。
func (r *StatsRepository) CountComments() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Comment{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountUsers 统计用户总数。
func (r *StatsRepository) CountUsers() (int64, error) {
	var count int64
	if err := r.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountCategories 统计分类总数。
func (r *StatsRepository) CountCategories() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Category{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountTags 统计标签总数。
func (r *StatsRepository) CountTags() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Tag{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetContentStats 查询指定内容类型与统计类型的时间序列数据。
func (r *StatsRepository) GetContentStats(contentType, statType string, startDate time.Time) ([]*model.ContentStats, error) {
	var stats []*model.ContentStats
	if err := r.db.Where("content_type = ? AND stat_type = ? AND stat_date >= ?",
		contentType, statType, startDate).
		Order("stat_date ASC").
		Find(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}
