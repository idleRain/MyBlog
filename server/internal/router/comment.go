package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// CommentRoutes 评论路由模块
type CommentRoutes struct {
	commentHandler handler.CommentHandlerInterface
	jwtService     service.JWTService
	identity       middleware.IdentityProvider
	rbacService    service.RBACService
}

// NewCommentRoutes 创建评论路由模块
func NewCommentRoutes(
	commentHandler handler.CommentHandlerInterface,
	jwtService service.JWTService,
	identity middleware.IdentityProvider,
	rbacService service.RBACService,
) *CommentRoutes {
	return &CommentRoutes{
		commentHandler: commentHandler,
		jwtService:     jwtService,
		identity:       identity,
		rbacService:    rbacService,
	}
}

// RegisterRoutes 注册评论相关路由
func (cr *CommentRoutes) RegisterRoutes(api *gin.RouterGroup, adminAPI *gin.RouterGroup) {
	// 公开评论接口，游客与登录用户均可访问。
	publicComments := api.Group("/comments")
	{
		publicComments.POST("/list", cr.commentHandler.GetCommentsByArticle) // 文章评论列表

		// 评论创建接口，登录用户自动绑定身份，游客匿名提交。
		publicComments.POST("/create",
			middleware.OptionalAuth(cr.jwtService),
			cr.commentHandler.CreateComment)
	}

	// 评论点赞接口，需要登录。
	authComments := api.Group("/comments")
	authComments.Use(middleware.Auth(cr.jwtService))
	{
		authComments.POST("/like", cr.commentHandler.LikeComment)     // 点赞评论
		authComments.POST("/unlike", cr.commentHandler.UnlikeComment) // 取消点赞评论
	}

	// 评论审核接口，需要评论审核权限。
	adminComments := adminAPI.Group("/comments")
	adminComments.Use(middleware.RequirePermission(cr.identity, cr.rbacService, service.PermissionCommentModerate))
	{
		adminComments.POST("/approve", cr.commentHandler.ApproveComment) // 审核通过
		adminComments.POST("/reject", cr.commentHandler.RejectComment)   // 拒绝评论
		adminComments.POST("/spam", cr.commentHandler.MarkCommentSpam)   // 标记垃圾
		adminComments.POST("/trash", cr.commentHandler.TrashComment)     // 移入回收站
		adminComments.POST("/delete", cr.commentHandler.DeleteComment)   // 删除评论
		adminComments.POST("/list", cr.commentHandler.ListComments)      // 评论列表
	}
}
