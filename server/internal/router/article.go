package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// ArticleRoutes 文章路由
type ArticleRoutes struct {
	articleHandler handler.ArticleHandlerInterface
	jwtService     service.JWTService
	userRepo       repository.UserRepository
	rbacService    service.RBACService
}

// NewArticleRoutes 创建文章路由实例
func NewArticleRoutes(
	articleHandler handler.ArticleHandlerInterface,
	jwtService service.JWTService,
	userRepo repository.UserRepository,
	rbacService service.RBACService,
) *ArticleRoutes {
	return &ArticleRoutes{
		articleHandler: articleHandler,
		jwtService:     jwtService,
		userRepo:       userRepo,
		rbacService:    rbacService,
	}
}

// RegisterRoutes 注册文章相关路由
func (ar *ArticleRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	// 公开访问的文章路由
	publicArticles := rg.Group("/articles")
	{
		// 文章查看相关（无需登录）
		publicArticles.POST("/get", ar.articleHandler.GetArticle)                   // 根据ID获取文章
		publicArticles.POST("/getBySlug", ar.articleHandler.GetArticleBySlug)       // 根据Slug获取文章
		publicArticles.POST("/list", ar.articleHandler.GetArticleList)              // 文章列表（支持筛选）
		publicArticles.POST("/byAuthor", ar.articleHandler.GetArticlesByAuthor)     // 作者文章列表
		publicArticles.POST("/byCategory", ar.articleHandler.GetArticlesByCategory) // 分类文章列表
		publicArticles.POST("/byTag", ar.articleHandler.GetArticlesByTag)           // 标签文章列表
		publicArticles.POST("/search", ar.articleHandler.SearchArticles)            // 搜索文章
		publicArticles.POST("/popular", ar.articleHandler.GetPopularArticles)       // 热门文章
		publicArticles.POST("/recent", ar.articleHandler.GetRecentArticles)         // 最新文章
		publicArticles.POST("/related", ar.articleHandler.GetRelatedArticles)       // 相关文章

		// 文章统计（无需登录）
		publicArticles.POST("/view", ar.articleHandler.ViewArticle) // 记录浏览量
	}

	// 需要登录的文章操作
	authArticles := rg.Group("/articles")
	authArticles.Use(middleware.Auth(ar.jwtService))
	{
		// 文章互动操作（需要登录）
		authArticles.POST("/like", ar.articleHandler.LikeArticle)             // 点赞文章
		authArticles.POST("/unlike", ar.articleHandler.UnlikeArticle)         // 取消点赞
		authArticles.POST("/bookmark", ar.articleHandler.BookmarkArticle)     // 收藏文章
		authArticles.POST("/unbookmark", ar.articleHandler.UnbookmarkArticle) // 取消收藏

		// 文章管理操作（需要编辑权限）
		editorArticles := authArticles.Group("")
		editorArticles.Use(middleware.RequirePermission(ar.jwtService, ar.userRepo, ar.rbacService, service.PermissionArticleCreate))
		{
			editorArticles.POST("/create", ar.articleHandler.CreateArticle)       // 创建文章
			editorArticles.POST("/update", ar.articleHandler.UpdateArticle)       // 更新文章
			editorArticles.POST("/delete", ar.articleHandler.DeleteArticle)       // 删除文章（软删除）
			editorArticles.POST("/publish", ar.articleHandler.PublishArticle)     // 发布文章
			editorArticles.POST("/unpublish", ar.articleHandler.UnpublishArticle) // 取消发布
			editorArticles.POST("/archive", ar.articleHandler.ArchiveArticle)     // 归档文章
			editorArticles.POST("/private", ar.articleHandler.SetArticlePrivate)  // 设为私有
		}
	}

	// 管理员文章管理路由
	adminArticles := rg.Group("/admin/articles")
	adminArticles.Use(middleware.RequirePermission(ar.jwtService, ar.userRepo, ar.rbacService, service.PermissionArticleManage))
	{
		adminArticles.POST("/list", ar.articleHandler.GetArticleList)        // 管理员文章列表（所有状态）
		adminArticles.POST("/update", ar.articleHandler.UpdateArticle)       // 管理员更新任意文章
		adminArticles.POST("/delete", ar.articleHandler.DeleteArticle)       // 管理员删除任意文章
		adminArticles.POST("/publish", ar.articleHandler.PublishArticle)     // 管理员发布任意文章
		adminArticles.POST("/unpublish", ar.articleHandler.UnpublishArticle) // 管理员取消发布
		adminArticles.POST("/archive", ar.articleHandler.ArchiveArticle)     // 管理员归档文章
		adminArticles.POST("/private", ar.articleHandler.SetArticlePrivate)  // 管理员设为私有
	}
}
