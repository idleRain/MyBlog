package router

import (
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// UserRoutes 用户路由模块
type UserRoutes struct {
	handler    UserHandlerInterface
	jwtService service.JWTService
}

// NewUserRoutes 创建用户路由模块
func NewUserRoutes(handler UserHandlerInterface, jwtService service.JWTService) *UserRoutes {
	return &UserRoutes{
		handler:    handler,
		jwtService: jwtService,
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

		// 用户管理相关路由（需要认证）
		authenticatedGroup := userGroup.Group("")
		authenticatedGroup.Use(middleware.Auth(ur.jwtService))
		{
			// 根据CLAUDE.md规范：统一使用POST方法
			authenticatedGroup.POST("/get", ur.handler.GetUserByID)
			authenticatedGroup.POST("/delete", ur.handler.DeleteUser)
			authenticatedGroup.POST("/list", ur.handler.GetUserList)
			{

			}
		}
	}

	// JWT相关路由
	authGroup := api.Group("/auth")
	{
		// 刷新令牌（无需认证）
		authGroup.POST("/refresh", ur.handler.RefreshToken)

		// 登出（需要认证）
		authGroup.POST("/logout", middleware.Auth(ur.jwtService), ur.handler.Logout)
	}
}
