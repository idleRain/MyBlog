package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// CategoryRoutes 分类路由模块
type CategoryRoutes struct {
	categoryHandler handler.CategoryHandlerInterface
	jwtService      service.JWTService
	identity        middleware.IdentityProvider
	rbacService     service.RBACService
}

// NewCategoryRoutes 创建分类路由模块
func NewCategoryRoutes(
	categoryHandler handler.CategoryHandlerInterface,
	jwtService service.JWTService,
	identity middleware.IdentityProvider,
	rbacService service.RBACService,
) *CategoryRoutes {
	return &CategoryRoutes{
		categoryHandler: categoryHandler,
		jwtService:      jwtService,
		identity:        identity,
		rbacService:     rbacService,
	}
}

// RegisterRoutes 注册分类相关路由
func (cr *CategoryRoutes) RegisterRoutes(api *gin.RouterGroup, adminAPI *gin.RouterGroup) {
	// 公开分类接口，无需登录。
	publicCategories := api.Group("/categories")
	{
		publicCategories.POST("/get", cr.categoryHandler.GetCategory)      // 根据ID获取分类
		publicCategories.POST("/tree", cr.categoryHandler.GetCategoryTree) // 分类树
	}

	// 分类管理接口，需要分类管理权限。
	adminCategories := adminAPI.Group("/categories")
	adminCategories.Use(middleware.RequirePermission(cr.identity, cr.rbacService, service.PermissionCategoryManage))
	{
		adminCategories.POST("/create", cr.categoryHandler.CreateCategory) // 创建分类
		adminCategories.POST("/update", cr.categoryHandler.UpdateCategory) // 更新分类
		adminCategories.POST("/delete", cr.categoryHandler.DeleteCategory) // 删除分类
		adminCategories.POST("/list", cr.categoryHandler.ListCategories)   // 分类列表
	}
}
