package router

import (
	"MyBlog/internal/handler"
	"MyBlog/internal/middleware"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

// StatsRoutes 站点统计路由模块
type StatsRoutes struct {
	statsHandler handler.StatsHandlerInterface
	jwtService   service.JWTService
	identity     middleware.IdentityProvider
	rbacService  service.RBACService
}

// NewStatsRoutes 创建站点统计路由模块
func NewStatsRoutes(
	statsHandler handler.StatsHandlerInterface,
	jwtService service.JWTService,
	identity middleware.IdentityProvider,
	rbacService service.RBACService,
) *StatsRoutes {
	return &StatsRoutes{
		statsHandler: statsHandler,
		jwtService:   jwtService,
		identity:     identity,
		rbacService:  rbacService,
	}
}

// RegisterRoutes 注册站点统计相关路由
func (sr *StatsRoutes) RegisterRoutes(adminAPI *gin.RouterGroup) {
	// 站点统计接口，需要统计权限。
	adminStats := adminAPI.Group("/stats")
	adminStats.Use(middleware.RequirePermission(sr.identity, sr.rbacService, service.PermissionSystemStats))
	{
		adminStats.POST("/overview", sr.statsHandler.GetOverview)          // 站点统计概览
		adminStats.POST("/articles", sr.statsHandler.GetArticleViewsTrend) // 文章浏览量趋势
	}
}
