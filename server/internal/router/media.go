package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// MediaRoutes 媒体路由模块
type MediaRoutes struct {
	mediaHandler handler.MediaHandlerInterface
	jwtService   service.JWTService
	identity     middleware.IdentityProvider
	rbacService  service.RBACService
}

// NewMediaRoutes 创建媒体路由模块
func NewMediaRoutes(
	mediaHandler handler.MediaHandlerInterface,
	jwtService service.JWTService,
	identity middleware.IdentityProvider,
	rbacService service.RBACService,
) *MediaRoutes {
	return &MediaRoutes{
		mediaHandler: mediaHandler,
		jwtService:   jwtService,
		identity:     identity,
		rbacService:  rbacService,
	}
}

// RegisterRoutes 注册媒体相关路由
func (mr *MediaRoutes) RegisterRoutes(api *gin.RouterGroup) {
	// 媒体接口需要登录，通过权限中间件同时注入操作者身份。
	media := api.Group("/media")
	{
		// 上传文件需要上传权限。
		media.POST("/upload",
			middleware.RequirePermission(mr.identity, mr.rbacService, service.PermissionFileUpload),
			mr.mediaHandler.UploadFile)

		// 查看媒体列表需要读取权限。
		media.POST("/list",
			middleware.RequirePermission(mr.identity, mr.rbacService, service.PermissionFileRead),
			mr.mediaHandler.ListMedia)

		// 查看媒体详情需要读取权限。
		media.POST("/get",
			middleware.RequirePermission(mr.identity, mr.rbacService, service.PermissionFileRead),
			mr.mediaHandler.GetMedia)

		// 删除文件需要删除权限。
		media.POST("/delete",
			middleware.RequirePermission(mr.identity, mr.rbacService, service.PermissionFileDelete),
			mr.mediaHandler.DeleteMedia)
	}
}
