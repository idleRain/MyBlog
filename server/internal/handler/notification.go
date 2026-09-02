// Package handler HTTP请求处理层
package handler

import (
	"errors"

	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// NotificationHandlerInterface 通知处理器接口
type NotificationHandlerInterface interface {
	ListNotifications(c *gin.Context)
	GetUnreadCount(c *gin.Context)
	MarkNotificationRead(c *gin.Context)
	MarkAllNotificationsRead(c *gin.Context)
}

// NotificationHandler 通知处理器实现
type NotificationHandler struct {
	notificationService service.NotificationServiceInterface
}

// NewNotificationHandler 创建通知处理器实例
func NewNotificationHandler(notificationService service.NotificationServiceInterface) NotificationHandlerInterface {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// ListNotifications 分页查询当前用户通知 POST /api/notifications/list
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	var req service.ListNotificationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	result, err := h.notificationService.ListNotifications(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetUnreadCount 获取未读通知数 POST /api/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"unreadCount": count})
}

// MarkNotificationRead 标记单条通知已读 POST /api/notifications/read
func (h *NotificationHandler) MarkNotificationRead(c *gin.Context) {
	type MarkNotificationReadRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req MarkNotificationReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.notificationService.MarkNotificationRead(req.ID, userID); err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "通知已标记为已读", nil)
}

// MarkAllNotificationsRead 标记全部通知已读 POST /api/notifications/read-all
func (h *NotificationHandler) MarkAllNotificationsRead(c *gin.Context) {
	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.notificationService.MarkAllNotificationsRead(userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "全部通知已标记为已读", nil)
}
