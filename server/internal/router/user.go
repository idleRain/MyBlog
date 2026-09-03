package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// UserRoutes 用户路由模块
type UserRoutes struct {
	userHandler handler.UserHandlerInterface
	jwtService  service.JWTService
	userRepo    repository.UserRepository
	rbacService service.RBACService
}

// NewUserRoutes 创建用户路由模块，RBAC 服务由组合根注入。
func NewUserRoutes(userHandler handler.UserHandlerInterface, jwtService service.JWTService, userRepo repository.UserRepository, rbacService service.RBACService) *UserRoutes {
	return &UserRoutes{
		userHandler: userHandler,
		jwtService:  jwtService,
		userRepo:    userRepo,
		rbacService: rbacService,
	}
}

// RegisterRoutes 注册用户相关路由
func (ur *UserRoutes) RegisterRoutes(api *gin.RouterGroup) {
	// 用户路由分组
	userGroup := api.Group("/users")
	{
		// 认证相关路由，无需令牌验证。
		userGroup.POST("/login", ur.userHandler.Login)

		// 用户查看接口，需要基础认证。
		userGroup.POST("/get",
			middleware.Auth(ur.jwtService),
			ur.userHandler.GetUserByID)

		// 用户创建接口，需要用户创建权限。
		userGroup.POST("/create",
			middleware.RequirePermission(ur.jwtService, ur.userRepo, ur.rbacService, service.PermissionUserCreate),
			ur.userHandler.CreateUser)

		// 用户更新接口，需要用户更新权限。
		userGroup.POST("/update",
			middleware.RequirePermission(ur.jwtService, ur.userRepo, ur.rbacService, service.PermissionUserUpdate),
			ur.userHandler.UpdateUser)

		// 用户删除接口，需要用户删除权限。
		userGroup.POST("/delete",
			middleware.RequirePermission(ur.jwtService, ur.userRepo, ur.rbacService, service.PermissionUserDelete),
			ur.userHandler.DeleteUser)

		// 用户列表接口，需要用户列表权限。
		userGroup.POST("/list",
			middleware.RequirePermission(ur.jwtService, ur.userRepo, ur.rbacService, service.PermissionUserList),
			ur.userHandler.GetUserList)
	}

	// JWT相关路由
	authGroup := api.Group("/auth")
	{
		// 刷新令牌接口，无需认证。
		authGroup.POST("/refresh", ur.userHandler.RefreshToken)

		// 登出接口，需要认证。
		authGroup.POST("/logout", middleware.Auth(ur.jwtService), ur.userHandler.Logout)
	}
}
