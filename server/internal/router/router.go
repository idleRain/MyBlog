package router

import (
	"net/http"

	"MyBlog/internal/config"
	"MyBlog/internal/middleware"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	engine *gin.Engine
	cfg    *config.Config
}

// NewRouter 创建新的路由管理器
func NewRouter(cfg *config.Config) *Router {
	engine := gin.New()

	// 设置全局中间件
	engine.Use(middleware.Logger())                          // 自定义日志中间件
	engine.Use(gin.Recovery())                               // 恢复中间件
	engine.Use(middleware.RequestID())                       // 请求ID中间件
	engine.Use(middleware.CORS())                            // CORS 中间件
	engine.Use(middleware.SecurityMiddlewareFromConfig(cfg)) // 综合安全中间件，规则与 config.yaml 保持一致

	// 统一处理未匹配路由，确保接口不存在时返回统一的 JSON 响应。
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    response.CodeNotFound,
			"message": "接口不存在",
		})
	})

	return &Router{
		engine: engine,
		cfg:    cfg,
	}
}

// GetEngine 获取 Gin 引擎实例
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// SetupRoutes 设置所有路由
func (r *Router) SetupRoutes(deps *Dependencies) {
	// API 根分组
	api := r.engine.Group("/api")

	// 管理员接口分组，统一应用更严格的安全中间件。
	adminAPI := api.Group("/admin")
	adminAPI.Use(middleware.AdminSecurityMiddlewareFromConfig(r.cfg))

	// 注册健康检查路由
	healthRoutes := NewHealthRoutes()
	healthRoutes.RegisterRoutes(api)

	// 注册用户相关路由
	if deps.UserHandler != nil {
		userHandler := deps.UserHandler.(UserHandlerInterface)
		userRoutes := NewUserRoutes(userHandler, deps.JWTService, deps.UserRepository)
		userRoutes.RegisterRoutes(api)
	}

	// 注册文章相关路由
	if deps.ArticleHandler != nil {
		articleHandler := deps.ArticleHandler.(ArticleHandlerInterface)
		articleRoutes := NewArticleRoutes(articleHandler, deps.JWTService, deps.UserRepository, deps.RBACService)
		articleRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册分类相关路由
	if deps.CategoryHandler != nil {
		categoryHandler := deps.CategoryHandler.(CategoryHandlerInterface)
		categoryRoutes := NewCategoryRoutes(categoryHandler, deps.JWTService, deps.UserRepository, deps.RBACService)
		categoryRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册标签相关路由
	if deps.TagHandler != nil {
		tagHandler := deps.TagHandler.(TagHandlerInterface)
		tagRoutes := NewTagRoutes(tagHandler, deps.JWTService, deps.UserRepository, deps.RBACService)
		tagRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册评论相关路由
	if deps.CommentHandler != nil {
		commentHandler := deps.CommentHandler.(CommentHandlerInterface)
		commentRoutes := NewCommentRoutes(commentHandler, deps.JWTService, deps.UserRepository, deps.RBACService)
		commentRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册媒体相关路由
	if deps.MediaHandler != nil {
		mediaHandler := deps.MediaHandler.(MediaHandlerInterface)
		mediaRoutes := NewMediaRoutes(mediaHandler, deps.JWTService, deps.UserRepository, deps.RBACService)
		mediaRoutes.RegisterRoutes(api)
	}
}

// Dependencies 依赖注入结构
type Dependencies struct {
	UserHandler     interface{}               // 用户处理器接口
	ArticleHandler  interface{}               // 文章处理器接口
	CategoryHandler interface{}               // 分类处理器接口
	TagHandler      interface{}               // 标签处理器接口
	CommentHandler  interface{}               // 评论处理器接口
	MediaHandler    interface{}               // 媒体处理器接口
	JWTService      service.JWTService        // JWT服务
	UserRepository  repository.UserRepository // 用户仓库
	RBACService     service.RBACService       // RBAC权限服务
}

// MediaHandlerInterface 媒体处理器接口
type MediaHandlerInterface interface {
	UploadFile(c *gin.Context)  // POST /api/media/upload - multipart
	GetMedia(c *gin.Context)    // POST /api/media/get
	ListMedia(c *gin.Context)   // POST /api/media/list
	DeleteMedia(c *gin.Context) // POST /api/media/delete
}

// CommentHandlerInterface 评论处理器接口
type CommentHandlerInterface interface {
	CreateComment(c *gin.Context)       // POST /api/comments/create
	GetCommentsByArticle(c *gin.Context) // POST /api/comments/list
	LikeComment(c *gin.Context)         // POST /api/comments/like
	UnlikeComment(c *gin.Context)       // POST /api/comments/unlike
	ApproveComment(c *gin.Context)      // POST /api/admin/comments/approve
	RejectComment(c *gin.Context)       // POST /api/admin/comments/reject
	MarkCommentSpam(c *gin.Context)     // POST /api/admin/comments/spam
	TrashComment(c *gin.Context)        // POST /api/admin/comments/trash
	DeleteComment(c *gin.Context)       // POST /api/admin/comments/delete
	ListComments(c *gin.Context)        // POST /api/admin/comments/list
}

// TagHandlerInterface 标签处理器接口
type TagHandlerInterface interface {
	CreateTag(c *gin.Context)     // POST /api/admin/tags/create
	UpdateTag(c *gin.Context)     // POST /api/admin/tags/update
	DeleteTag(c *gin.Context)     // POST /api/admin/tags/delete
	GetTag(c *gin.Context)        // POST /api/tags/get
	ListTags(c *gin.Context)      // POST /api/admin/tags/list
	GetPopularTags(c *gin.Context) // POST /api/tags/popular
}

// CategoryHandlerInterface 分类处理器接口
type CategoryHandlerInterface interface {
	CreateCategory(c *gin.Context)   // POST /api/admin/categories/create
	UpdateCategory(c *gin.Context)   // POST /api/admin/categories/update
	DeleteCategory(c *gin.Context)   // POST /api/admin/categories/delete
	GetCategory(c *gin.Context)      // POST /api/categories/get
	ListCategories(c *gin.Context)   // POST /api/admin/categories/list
	GetCategoryTree(c *gin.Context)  // POST /api/categories/tree
}

// UserHandlerInterface 用户处理器接口
type UserHandlerInterface interface {
	CreateUser(c *gin.Context)   // POST /api/users/create - JSON格式
	UpdateUser(c *gin.Context)   // POST /api/users/update - JSON格式
	GetUserByID(c *gin.Context)  // POST /api/users/get - JSON格式
	GetUserList(c *gin.Context)  // POST /api/users/list - JSON格式，用于复杂参数查询
	DeleteUser(c *gin.Context)   // POST /api/users/delete - JSON格式
	Login(c *gin.Context)        // POST /api/users/login - JSON格式
	RefreshToken(c *gin.Context) // POST /api/auth/refresh - JSON格式
	Logout(c *gin.Context)       // POST /api/auth/logout - Header中的Token
}

// ArticleHandlerInterface 文章处理器接口
type ArticleHandlerInterface interface {
	// 基础CRUD操作
	CreateArticle(c *gin.Context)
	GetArticle(c *gin.Context)
	GetArticleBySlug(c *gin.Context)
	UpdateArticle(c *gin.Context)
	DeleteArticle(c *gin.Context)

	// 查询操作
	GetArticleList(c *gin.Context)
	GetArticlesByAuthor(c *gin.Context)
	GetArticlesByCategory(c *gin.Context)
	GetArticlesByTag(c *gin.Context)
	SearchArticles(c *gin.Context)

	// 统计和推荐
	GetPopularArticles(c *gin.Context)
	GetRecentArticles(c *gin.Context)
	GetRelatedArticles(c *gin.Context)

	// 互动操作
	ViewArticle(c *gin.Context)
	LikeArticle(c *gin.Context)
	UnlikeArticle(c *gin.Context)
	BookmarkArticle(c *gin.Context)
	UnbookmarkArticle(c *gin.Context)

	// 状态管理
	PublishArticle(c *gin.Context)
	UnpublishArticle(c *gin.Context)
	ArchiveArticle(c *gin.Context)
	SetArticlePrivate(c *gin.Context)
}
