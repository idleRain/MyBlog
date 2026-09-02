package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// NotificationRoutes 通知路由模块
type NotificationRoutes struct {
	notificationHandler handler.NotificationHandlerInterface
	jwtService          service.JWTService
}

// NewNotificationRoutes 创建通知路由模块
func NewNotificationRoutes(
	notificationHandler handler.NotificationHandlerInterface,
	jwtService service.JWTService,
) *NotificationRoutes {
	return &NotificationRoutes{
		notificationHandler: notificationHandler,
		jwtService:          jwtService,
	}
}

// RegisterRoutes 注册通知相关路由
func (nr *NotificationRoutes) RegisterRoutes(api *gin.RouterGroup) {
	// 通知接口需要登录，且仅能操作本人通知。
	notifications := api.Group("/notifications")
	notifications.Use(middleware.Auth(nr.jwtService))
	{
		notifications.POST("/list", nr.notificationHandler.ListNotifications)            // 通知列表
		notifications.POST("/unread-count", nr.notificationHandler.GetUnreadCount)       // 未读数
		notifications.POST("/read", nr.notificationHandler.MarkNotificationRead)         // 标记单条已读
		notifications.POST("/read-all", nr.notificationHandler.MarkAllNotificationsRead) // 标记全部已读
	}
}
