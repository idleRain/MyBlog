package router

import (
	"net/http"

	"MyBlog/internal/config"
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
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
		userRoutes := NewUserRoutes(deps.UserHandler, deps.JWTService, deps.IdentityProvider, deps.RBACService)
		userRoutes.RegisterRoutes(api)
	}

	// 注册文章相关路由
	if deps.ArticleHandler != nil {
		articleRoutes := NewArticleRoutes(deps.ArticleHandler, deps.JWTService, deps.IdentityProvider, deps.RBACService)
		articleRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册分类相关路由
	if deps.CategoryHandler != nil {
		categoryRoutes := NewCategoryRoutes(deps.CategoryHandler, deps.JWTService, deps.IdentityProvider, deps.RBACService)
		categoryRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册标签相关路由
	if deps.TagHandler != nil {
		tagRoutes := NewTagRoutes(deps.TagHandler, deps.JWTService, deps.IdentityProvider, deps.RBACService)
		tagRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册评论相关路由
	if deps.CommentHandler != nil {
		commentRoutes := NewCommentRoutes(deps.CommentHandler, deps.JWTService, deps.IdentityProvider, deps.RBACService)
		commentRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册媒体相关路由
	if deps.MediaHandler != nil {
		mediaRoutes := NewMediaRoutes(deps.MediaHandler, deps.JWTService, deps.IdentityProvider, deps.RBACService)
		mediaRoutes.RegisterRoutes(api)
	}

	// 注册设置相关路由
	if deps.SettingHandler != nil {
		settingRoutes := NewSettingRoutes(deps.SettingHandler, deps.JWTService, deps.IdentityProvider, deps.RBACService)
		settingRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册友情链接相关路由
	if deps.FriendlyLinkHandler != nil {
		linkRoutes := NewFriendlyLinkRoutes(deps.FriendlyLinkHandler, deps.JWTService, deps.IdentityProvider, deps.RBACService)
		linkRoutes.RegisterRoutes(api, adminAPI)
	}

	// 注册站点统计相关路由
	if deps.StatsHandler != nil {
		statsRoutes := NewStatsRoutes(deps.StatsHandler, deps.JWTService, deps.IdentityProvider, deps.RBACService)
		statsRoutes.RegisterRoutes(adminAPI)
	}

	// 注册通知相关路由
	if deps.NotificationHandler != nil {
		notificationRoutes := NewNotificationRoutes(deps.NotificationHandler, deps.JWTService)
		notificationRoutes.RegisterRoutes(api)
	}

	// 注册用户关注相关路由
	if deps.UserFollowHandler != nil {
		followRoutes := NewUserFollowRoutes(deps.UserFollowHandler, deps.JWTService)
		followRoutes.RegisterRoutes(api)
	}
}

// Dependencies 依赖注入结构，handler 字段使用 handler 包内的具体接口类型，编译期即可校验注入正确性。
type Dependencies struct {
	UserHandler         handler.UserHandlerInterface         // 用户处理器接口
	ArticleHandler      handler.ArticleHandlerInterface      // 文章处理器接口
	CategoryHandler     handler.CategoryHandlerInterface     // 分类处理器接口
	TagHandler          handler.TagHandlerInterface          // 标签处理器接口
	CommentHandler      handler.CommentHandlerInterface      // 评论处理器接口
	MediaHandler        handler.MediaHandlerInterface        // 媒体处理器接口
	SettingHandler      handler.SettingHandlerInterface      // 设置处理器接口
	FriendlyLinkHandler handler.FriendlyLinkHandlerInterface // 友情链接处理器接口
	StatsHandler        handler.StatsHandlerInterface        // 站点统计处理器接口
	NotificationHandler handler.NotificationHandlerInterface // 通知处理器接口
	UserFollowHandler   handler.UserFollowHandlerInterface   // 用户关注处理器接口
	JWTService          service.JWTService                   // JWT服务
	IdentityProvider    middleware.IdentityProvider          // 身份解析抽象
	RBACService         service.RBACService                  // RBAC权限服务
}
