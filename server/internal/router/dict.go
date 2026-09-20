package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// DictRoutes 字典路由模块
type DictRoutes struct {
	dictHandler handler.DictHandlerInterface
	jwtService  service.JWTService
	identity    middleware.IdentityProvider
	rbacService service.RBACService
}

// NewDictRoutes 创建字典路由模块
func NewDictRoutes(
	dictHandler handler.DictHandlerInterface,
	jwtService service.JWTService,
	identity middleware.IdentityProvider,
	rbacService service.RBACService,
) *DictRoutes {
	return &DictRoutes{
		dictHandler: dictHandler,
		jwtService:  jwtService,
		identity:    identity,
		rbacService: rbacService,
	}
}

// RegisterRoutes 注册字典相关路由
func (dr *DictRoutes) RegisterRoutes(api *gin.RouterGroup, adminAPI *gin.RouterGroup) {
	// 公开字典接口，无需登录，供 app 端读取已生效字典。
	publicDicts := api.Group("/dicts")
	{
		publicDicts.POST("/all", dr.dictHandler.GetAllDicts)     // 全量已生效字典
		publicDicts.POST("/:type", dr.dictHandler.GetDictByType) // 按字典码查询单列表
	}

	// 字典管理接口，需要字典管理权限。
	adminDicts := adminAPI.Group("/dicts")
	adminDicts.Use(middleware.RequirePermission(dr.identity, dr.rbacService, service.PermissionDictManage))
	{
		adminDicts.POST("/types/create", dr.dictHandler.CreateDictType) // 创建字典类型
		adminDicts.POST("/types/update", dr.dictHandler.UpdateDictType) // 更新字典类型
		adminDicts.POST("/types/delete", dr.dictHandler.DeleteDictType) // 删除字典类型
		adminDicts.POST("/types/list", dr.dictHandler.ListDictTypes)    // 字典类型列表
		adminDicts.POST("/items/create", dr.dictHandler.CreateDictItem) // 创建字典项
		adminDicts.POST("/items/update", dr.dictHandler.UpdateDictItem) // 更新字典项
		adminDicts.POST("/items/delete", dr.dictHandler.DeleteDictItem) // 删除字典项
		adminDicts.POST("/items/list", dr.dictHandler.ListDictItems)    // 字典项列表
	}
}
