package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// FriendlyLinkRoutes 友情链接路由模块
type FriendlyLinkRoutes struct {
	linkHandler handler.FriendlyLinkHandlerInterface
	jwtService  service.JWTService
	userRepo    repository.UserRepository
	rbacService service.RBACService
}

// NewFriendlyLinkRoutes 创建友情链接路由模块
func NewFriendlyLinkRoutes(
	linkHandler handler.FriendlyLinkHandlerInterface,
	jwtService service.JWTService,
	userRepo repository.UserRepository,
	rbacService service.RBACService,
) *FriendlyLinkRoutes {
	return &FriendlyLinkRoutes{
		linkHandler: linkHandler,
		jwtService:  jwtService,
		userRepo:    userRepo,
		rbacService: rbacService,
	}
}

// RegisterRoutes 注册友情链接相关路由
func (fr *FriendlyLinkRoutes) RegisterRoutes(api *gin.RouterGroup, adminAPI *gin.RouterGroup) {
	// 公开友情链接接口，无需登录。
	api.POST("/friendly-links/list", fr.linkHandler.ListVisibleLinks)

	// 友情链接管理接口，需要系统配置权限。
	adminLinks := adminAPI.Group("/friendly-links")
	adminLinks.Use(middleware.RequirePermission(fr.jwtService, fr.userRepo, fr.rbacService, service.PermissionSystemConfig))
	{
		adminLinks.POST("/create", fr.linkHandler.CreateLink)   // 创建友情链接
		adminLinks.POST("/update", fr.linkHandler.UpdateLink)   // 更新友情链接
		adminLinks.POST("/delete", fr.linkHandler.DeleteLink)   // 删除友情链接
		adminLinks.POST("/approve", fr.linkHandler.ApproveLink) // 审核通过
		adminLinks.POST("/hide", fr.linkHandler.HideLink)       // 下架
		adminLinks.POST("/reject", fr.linkHandler.RejectLink)   // 拒绝
		adminLinks.POST("/list", fr.linkHandler.ListLinks)      // 链接列表
	}
}
