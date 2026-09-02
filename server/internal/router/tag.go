package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// TagRoutes 标签路由模块
type TagRoutes struct {
	tagHandler  handler.TagHandlerInterface
	jwtService  service.JWTService
	userRepo    repository.UserRepository
	rbacService service.RBACService
}

// NewTagRoutes 创建标签路由模块
func NewTagRoutes(
	tagHandler handler.TagHandlerInterface,
	jwtService service.JWTService,
	userRepo repository.UserRepository,
	rbacService service.RBACService,
) *TagRoutes {
	return &TagRoutes{
		tagHandler:  tagHandler,
		jwtService:  jwtService,
		userRepo:    userRepo,
		rbacService: rbacService,
	}
}

// RegisterRoutes 注册标签相关路由
func (tr *TagRoutes) RegisterRoutes(api *gin.RouterGroup, adminAPI *gin.RouterGroup) {
	// 公开标签接口，无需登录。
	publicTags := api.Group("/tags")
	{
		publicTags.POST("/get", tr.tagHandler.GetTag)               // 根据ID获取标签
		publicTags.POST("/popular", tr.tagHandler.GetPopularTags)   // 热门标签
	}

	// 标签管理接口，需要标签管理权限。
	adminTags := adminAPI.Group("/tags")
	adminTags.Use(middleware.RequirePermission(tr.jwtService, tr.userRepo, tr.rbacService, service.PermissionTagManage))
	{
		adminTags.POST("/create", tr.tagHandler.CreateTag) // 创建标签
		adminTags.POST("/update", tr.tagHandler.UpdateTag) // 更新标签
		adminTags.POST("/delete", tr.tagHandler.DeleteTag) // 删除标签
		adminTags.POST("/list", tr.tagHandler.ListTags)   // 标签列表
	}
}
