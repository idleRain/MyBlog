package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// ArticleRoutes 文章路由
type ArticleRoutes struct {
	articleHandler handler.ArticleHandlerInterface
	jwtService     service.JWTService
	identity       middleware.IdentityProvider
	rbacService    service.RBACService
}

// NewArticleRoutes 创建文章路由实例
func NewArticleRoutes(
	articleHandler handler.ArticleHandlerInterface,
	jwtService service.JWTService,
	identity middleware.IdentityProvider,
	rbacService service.RBACService,
) *ArticleRoutes {
	return &ArticleRoutes{
		articleHandler: articleHandler,
		jwtService:     jwtService,
		identity:       identity,
		rbacService:    rbacService,
	}
}

// RegisterRoutes 注册文章相关路由
func (ar *ArticleRoutes) RegisterRoutes(rg *gin.RouterGroup, _ *gin.RouterGroup) {
	// 公开访问的文章路由，携带有效令牌时注入用户身份用于角色化可见性判断。
	publicArticles := rg.Group("/articles")
	publicArticles.Use(middleware.OptionalAuth(ar.jwtService))
	{
		// 文章查看接口，无需登录。
		publicArticles.POST("/get", ar.articleHandler.GetArticle)                   // 根据ID获取文章
		publicArticles.POST("/getBySlug", ar.articleHandler.GetArticleBySlug)       // 根据Slug获取文章
		publicArticles.POST("/list", ar.articleHandler.GetArticleList)              // 文章列表接口，支持多种筛选条件。
		publicArticles.POST("/byAuthor", ar.articleHandler.GetArticlesByAuthor)     // 作者文章列表
		publicArticles.POST("/byCategory", ar.articleHandler.GetArticlesByCategory) // 分类文章列表
		publicArticles.POST("/byTag", ar.articleHandler.GetArticlesByTag)           // 标签文章列表
		publicArticles.POST("/search", ar.articleHandler.SearchArticles)            // 搜索文章
		publicArticles.POST("/popular", ar.articleHandler.GetPopularArticles)       // 热门文章
		publicArticles.POST("/recent", ar.articleHandler.GetRecentArticles)         // 最新文章
		publicArticles.POST("/related", ar.articleHandler.GetRelatedArticles)       // 相关文章

		// 文章统计接口，无需登录。
		publicArticles.POST("/view", ar.articleHandler.ViewArticle) // 记录浏览量
	}

	// 需要登录的文章操作
	authArticles := rg.Group("/articles")
	authArticles.Use(middleware.Auth(ar.jwtService))
	{
		// 文章互动操作接口，需要登录。
		authArticles.POST("/like", ar.articleHandler.LikeArticle)             // 点赞文章
		authArticles.POST("/unlike", ar.articleHandler.UnlikeArticle)         // 取消点赞
		authArticles.POST("/bookmark", ar.articleHandler.BookmarkArticle)     // 收藏文章
		authArticles.POST("/unbookmark", ar.articleHandler.UnbookmarkArticle) // 取消收藏

		// 文章管理操作接口，需要编辑权限。
		// 授权由服务层统一判定：作者（article:create）或具备 article:manage 的管理员均可操作。
		editorArticles := authArticles.Group("")
		editorArticles.Use(middleware.RequirePermission(ar.identity, ar.rbacService, service.PermissionArticleCreate))
		{
			editorArticles.POST("/create", ar.articleHandler.CreateArticle)       // 创建文章
			editorArticles.POST("/update", ar.articleHandler.UpdateArticle)       // 更新文章
			editorArticles.POST("/delete", ar.articleHandler.DeleteArticle)       // 删除文章，采用软删除。
			editorArticles.POST("/publish", ar.articleHandler.PublishArticle)     // 发布文章
			editorArticles.POST("/unpublish", ar.articleHandler.UnpublishArticle) // 取消发布
			editorArticles.POST("/archive", ar.articleHandler.ArchiveArticle)     // 归档文章
			editorArticles.POST("/private", ar.articleHandler.SetArticlePrivate)  // 设为私有
		}
	}
}
