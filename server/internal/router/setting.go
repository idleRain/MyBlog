package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingRoutes 设置路由模块
type SettingRoutes struct {
	settingHandler handler.SettingHandlerInterface
	jwtService     service.JWTService
	userRepo       repository.UserRepository
	rbacService    service.RBACService
}

// NewSettingRoutes 创建设置路由模块
func NewSettingRoutes(
	settingHandler handler.SettingHandlerInterface,
	jwtService service.JWTService,
	userRepo repository.UserRepository,
	rbacService service.RBACService,
) *SettingRoutes {
	return &SettingRoutes{
		settingHandler: settingHandler,
		jwtService:     jwtService,
		userRepo:       userRepo,
		rbacService:    rbacService,
	}
}

// RegisterRoutes 注册设置相关路由
func (sr *SettingRoutes) RegisterRoutes(api *gin.RouterGroup, adminAPI *gin.RouterGroup) {
	// 公开设置接口，无需登录。
	api.POST("/settings/public", sr.settingHandler.GetPublicSettings)

	// 设置管理接口，需要系统配置权限。
	adminSettings := adminAPI.Group("/settings")
	adminSettings.Use(middleware.RequirePermission(sr.jwtService, sr.userRepo, sr.rbacService, service.PermissionSystemConfig))
	{
		adminSettings.POST("/list", sr.settingHandler.ListSettings)     // 设置项列表
		adminSettings.POST("/update", sr.settingHandler.UpdateSettings) // 批量更新设置项
	}
}
