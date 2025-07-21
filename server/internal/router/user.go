package router

import (
	"github.com/gin-gonic/gin"
)

// UserRoutes 用户路由模块
type UserRoutes struct {
	handler UserHandlerInterface
}

// NewUserRoutes 创建用户路由模块
func NewUserRoutes(handler UserHandlerInterface) *UserRoutes {
	return &UserRoutes{
		handler: handler,
	}
}

// RegisterRoutes 注册用户相关路由
func (ur *UserRoutes) RegisterRoutes(api *gin.RouterGroup) {
	// 用户路由分组
	userGroup := api.Group("/users")
	{
		// 认证相关路由（无需token验证）
		userGroup.POST("/create", ur.handler.CreateUser)
		userGroup.POST("/login", ur.handler.Login)

		// 用户管理相关路由
		userGroup.POST("/get", ur.handler.GetUserByID)
		userGroup.POST("/list", ur.handler.GetUserList)
		userGroup.POST("/delete", ur.handler.DeleteUser)

		// 可以在这里添加更多用户相关的路由
	}
}
