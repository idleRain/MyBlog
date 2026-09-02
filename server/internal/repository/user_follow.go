// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrFollowNotFound 关注关系不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrFollowNotFound = errors.New("关注关系不存在")

// UserFollowRepositoryInterface 用户关注仓储接口
type UserFollowRepositoryInterface interface {
	// 关注操作
	Follow(followerID, followingID uint) (bool, error)
	Unfollow(followerID, followingID uint) (bool, error)

	// 查询操作
	ListFollowers(userID uint, params *FollowListParams) ([]*model.UserFollow, int64, error)
	ListFollowing(userID uint, params *FollowListParams) ([]*model.UserFollow, int64, error)
	IsFollowing(followerID, followingID uint) (bool, error)
}

// FollowListParams 关注列表查询参数
type FollowListParams struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// UserFollowRepository 用户关注仓储实现
type UserFollowRepository struct {
	db *gorm.DB
}

// NewUserFollowRepository 创建用户关注仓储实例
func NewUserFollowRepository(db *gorm.DB) UserFollowRepositoryInterface {
	return &UserFollowRepository{db: db}
}

// Follow 建立关注关系，依赖唯一索引防重复，返回是否新增。
func (r *UserFollowRepository) Follow(followerID, followingID uint) (bool, error) {
	var added bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.UserFollow{
			FollowerID:  followerID,
			FollowingID: followingID,
		})
		if result.Error != nil {
			return result.Error
		}
		added = result.RowsAffected > 0
		return nil
	})
	return added, err
}

// Unfollow 取消关注关系，返回是否实际删除。
func (r *UserFollowRepository) Unfollow(followerID, followingID uint) (bool, error) {
	var removed bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("follower_id = ? AND following_id = ?", followerID, followingID).
			Delete(&model.UserFollow{})
		if result.Error != nil {
			return result.Error
		}
		removed = result.RowsAffected > 0
		return nil
	})
	return removed, err
}

// ListFollowers 分页查询用户的粉丝列表。
func (r *UserFollowRepository) ListFollowers(userID uint, params *FollowListParams) ([]*model.UserFollow, int64, error) {
	query := r.db.Model(&model.UserFollow{}).Where("following_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询粉丝总数失败: %w", err)
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var follows []*model.UserFollow
	offset := (params.Page - 1) * params.PageSize
	if err := query.Preload("Follower").
		Order("created_at DESC").
		Offset(offset).Limit(params.PageSize).
		Find(&follows).Error; err != nil {
		return nil, 0, fmt.Errorf("查询粉丝列表失败: %w", err)
	}

	return follows, total, nil
}

// ListFollowing 分页查询用户的关注列表。
func (r *UserFollowRepository) ListFollowing(userID uint, params *FollowListParams) ([]*model.UserFollow, int64, error) {
	query := r.db.Model(&model.UserFollow{}).Where("follower_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询关注总数失败: %w", err)
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var follows []*model.UserFollow
	offset := (params.Page - 1) * params.PageSize
	if err := query.Preload("Following").
		Order("created_at DESC").
		Offset(offset).Limit(params.PageSize).
		Find(&follows).Error; err != nil {
		return nil, 0, fmt.Errorf("查询关注列表失败: %w", err)
	}

	return follows, total, nil
}

// IsFollowing 检查是否已关注目标用户。
func (r *UserFollowRepository) IsFollowing(followerID, followingID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.UserFollow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("查询关注关系失败: %w", err)
	}
	return count > 0, nil
}
