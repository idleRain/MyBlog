// Package service 业务逻辑层
package service

import (
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// NotificationServiceInterface 通知服务接口
type NotificationServiceInterface interface {
	// 查询操作
	ListNotifications(userID uint, req *ListNotificationsRequest) (*NotificationListResponse, error)
	GetUnreadCount(userID uint) (int64, error)

	// 已读操作
	MarkNotificationRead(id uint, userID uint) error
	MarkAllNotificationsRead(userID uint) error
}

// ListNotificationsRequest 通知列表请求
type ListNotificationsRequest struct {
	Page     int    `json:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"pageSize" binding:"omitempty,min=1,max=100"`
	Type     string `json:"type" binding:"omitempty,oneof=comment_reply article_like comment_like system follow article_new"`
	IsRead   *bool  `json:"isRead"`
}

// NotificationListResponse 通知列表响应
type NotificationListResponse struct {
	Notifications []*model.Notification `json:"notifications"`
	Total         int64                 `json:"total"`
	Page          int                   `json:"page"`
	PageSize      int                   `json:"pageSize"`
	UnreadCount   int64                 `json:"unreadCount"`
}

// NotificationService 通知服务实现
type NotificationService struct {
	notificationRepo repository.NotificationRepositoryInterface
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(notificationRepo repository.NotificationRepositoryInterface) NotificationServiceInterface {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

// ListNotifications 分页查询当前用户通知，附带未读数。
func (s *NotificationService) ListNotifications(userID uint, req *ListNotificationsRequest) (*NotificationListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	params := &repository.NotificationListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Type:     req.Type,
		IsRead:   req.IsRead,
	}

	notifications, total, err := s.notificationRepo.ListByUser(userID, params)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.notificationRepo.CountUnread(userID)
	if err != nil {
		return nil, err
	}

	return &NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		Page:          req.Page,
		PageSize:      req.PageSize,
		UnreadCount:   unreadCount,
	}, nil
}

// GetUnreadCount 获取当前用户未读通知数。
func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	return s.notificationRepo.CountUnread(userID)
}

// MarkNotificationRead 将指定通知标记为已读。
func (s *NotificationService) MarkNotificationRead(id uint, userID uint) error {
	return s.notificationRepo.MarkAsRead(id, userID)
}

// MarkAllNotificationsRead 将当前用户全部通知标记为已读。
func (s *NotificationService) MarkAllNotificationsRead(userID uint) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}
