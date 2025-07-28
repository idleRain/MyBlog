package router

import (
	"MyBlog/internal/middleware"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// UserRoutes 用户路由模块
type UserRoutes struct {
	handler     UserHandlerInterface
	jwtService  service.JWTService
	userRepo    repository.UserRepository
	rbacService service.RBACService
}

// NewUserRoutes 创建用户路由模块
func NewUserRoutes(handler UserHandlerInterface, jwtService service.JWTService, userRepo repository.UserRepository) *UserRoutes {
	return &UserRoutes{
		handler:     handler,
		jwtService:  jwtService,
		userRepo:    userRepo,
		rbacService: service.NewRBACService(),
	}
}

// RegisterRoutes 注册用户相关路由
func (ur *UserRoutes) RegisterRoutes(api *gin.RouterGroup) {
	// 用户路由分组
	userGroup := api.Group("/users")
	{
		// 认证相关路由（无需token验证）
		userGroup.POST("/login", ur.handler.Login)

		// 用户查看接口（需要基础认证）
		userGroup.POST("/get",
			middleware.Auth(ur.jwtService),
			ur.handler.GetUserByID)

		// 用户创建（需要用户创建权限）
		userGroup.POST("/create",
			middleware.RequirePermission(ur.jwtService, ur.userRepo, ur.rbacService, service.PermissionUserCreate),
			ur.handler.CreateUser)

		// 用户更新（需要用户更新权限）
		userGroup.POST("/update",
			middleware.RequirePermission(ur.jwtService, ur.userRepo, ur.rbacService, service.PermissionUserUpdate),
			ur.handler.UpdateUser)

		// 用户删除（需要用户删除权限）
		userGroup.POST("/delete",
			middleware.RequirePermission(ur.jwtService, ur.userRepo, ur.rbacService, service.PermissionUserDelete),
			ur.handler.DeleteUser)

		// 用户列表（需要用户列表权限）
		userGroup.POST("/list",
			middleware.RequirePermission(ur.jwtService, ur.userRepo, ur.rbacService, service.PermissionUserList),
			ur.handler.GetUserList)
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
