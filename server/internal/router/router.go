package router

import (
	"MyBlog/internal/middleware"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	engine *gin.Engine
}

// NewRouter 创建新的路由管理器
func NewRouter() *Router {
	engine := gin.New()

	// 设置全局中间件
	engine.Use(middleware.Logger())                                               // 自定义日志中间件
	engine.Use(gin.Recovery())                                                    // 恢复中间件
	engine.Use(middleware.RequestID())                                            // 请求ID中间件
	engine.Use(middleware.CORS())                                                 // CORS 中间件
	engine.Use(middleware.SecurityMiddleware(middleware.DefaultSecurityConfig())) // 综合安全中间件

	return &Router{
		engine: engine,
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

	// 注册健康检查路由
	healthRoutes := NewHealthRoutes()
	healthRoutes.RegisterRoutes(api)

	// 注册用户相关路由
	if deps.UserHandler != nil {
		userHandler := deps.UserHandler.(UserHandlerInterface)
		userRoutes := NewUserRoutes(userHandler, deps.JWTService, deps.UserRepository)
		userRoutes.RegisterRoutes(api)
	}

	// 可以在这里添加更多的路由模块
	// if deps.PostHandler != nil {
	//     postHandler := deps.PostHandler.(PostHandlerInterface)
	//     postRoutes := NewPostRoutes(postHandler)
	//     postRoutes.RegisterRoutes(api)
	// }
}

// Dependencies 依赖注入结构
type Dependencies struct {
	UserHandler    interface{}               // 用户处理器接口
	JWTService     service.JWTService        // JWT服务
	UserRepository repository.UserRepository // 用户仓库
	// 可以添加更多的依赖
	// PostHandler interface{}
	// AuthHandler interface{}
}

// UserHandlerInterface 用户处理器接口
type UserHandlerInterface interface {
	CreateUser(c *gin.Context)   // POST /api/users/create - JSON格式
	UpdateUser(c *gin.Context)   // POST /api/users/update - JSON格式
	GetUserByID(c *gin.Context)  // GET /api/users/:id - URL路径参数
	GetUserList(c *gin.Context)  // POST /api/users/list - JSON格式（复杂参数）
	DeleteUser(c *gin.Context)   // DELETE /api/users/:id - URL路径参数
	Login(c *gin.Context)        // POST /api/users/login - JSON格式
	RefreshToken(c *gin.Context) // POST /api/auth/refresh - JSON格式
	Logout(c *gin.Context)       // POST /api/auth/logout - Header中的Token
}
