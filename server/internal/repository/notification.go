// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"

	"gorm.io/gorm"
)

// ErrNotificationNotFound 通知不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrNotificationNotFound = errors.New("通知不存在")

// NotificationRepositoryInterface 通知仓储接口
type NotificationRepositoryInterface interface {
	// 基础操作
	Create(notification *model.Notification) error
	GetByID(id uint) (*model.Notification, error)
	Update(notification *model.Notification) error

	// 查询操作
	ListByUser(userID uint, params *NotificationListParams) ([]*model.Notification, int64, error)
	CountUnread(userID uint) (int64, error)

	// 已读操作
	MarkAsRead(id uint, userID uint) error
	MarkAllAsRead(userID uint) error
}

// NotificationListParams 通知列表查询参数
type NotificationListParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Type     string `json:"type"`
	IsRead   *bool  `json:"isRead"`
}

// NotificationRepository 通知仓储实现
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓储实例
func NewNotificationRepository(db *gorm.DB) NotificationRepositoryInterface {
	return &NotificationRepository{db: db}
}

// Create 创建通知
func (r *NotificationRepository) Create(notification *model.Notification) error {
	return r.db.Create(notification).Error
}

// GetByID 根据ID获取通知
func (r *NotificationRepository) GetByID(id uint) (*model.Notification, error) {
	var notification model.Notification
	if err := r.db.First(&notification, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	return &notification, nil
}

// Update 更新通知
func (r *NotificationRepository) Update(notification *model.Notification) error {
	return r.db.Save(notification).Error
}

// ListByUser 分页查询用户通知。
func (r *NotificationRepository) ListByUser(userID uint, params *NotificationListParams) ([]*model.Notification, int64, error) {
	query := r.db.Model(&model.Notification{}).Where("user_id = ?", userID)

	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.IsRead != nil {
		query = query.Where("is_read = ?", *params.IsRead)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询通知总数失败: %w", err)
	}

	// 设置分页默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var notifications []*model.Notification
	offset := (params.Page - 1) * params.PageSize
	if err := query.Preload("Sender").
		Order("created_at DESC").
		Offset(offset).Limit(params.PageSize).
		Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("查询通知列表失败: %w", err)
	}

	return notifications, total, nil
}

// CountUnread 统计用户未读通知数。
func (r *NotificationRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计未读通知失败: %w", err)
	}
	return count, nil
}

// MarkAsRead 将指定通知标记为已读，仅本人通知可操作。
func (r *NotificationRepository) MarkAsRead(id uint, userID uint) error {
	result := r.db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)
	if result.Error != nil {
		return fmt.Errorf("标记通知已读失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

// MarkAllAsRead 将用户全部通知标记为已读。
func (r *NotificationRepository) MarkAllAsRead(userID uint) error {
	if err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("标记全部通知已读失败: %w", err)
	}
	return nil
}
