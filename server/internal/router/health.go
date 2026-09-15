package router

import (
	"net/http"

	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// serviceName 健康检查响应的服务标识，站点英文名变更时同步更新此处。
const serviceName = "IdleRain Blogs API"

// HealthRoutes 健康检查路由模块，数据库探针经组合根注入，路由层不直接依赖数据库实现。
type HealthRoutes struct {
	dbCheck func() error
}

// NewHealthRoutes 创建健康检查路由模块，dbCheck 为数据库连通性探测函数。
func NewHealthRoutes(dbCheck func() error) *HealthRoutes {
	return &HealthRoutes{dbCheck: dbCheck}
}

// RegisterRoutes 注册健康检查路由。
// POST-Only 规范约束业务接口，健康探针属基础设施端点，编排器只支持 GET 探测，故采用 GET 路由。
func (hr *HealthRoutes) RegisterRoutes(api *gin.RouterGroup) {
	// 存活探针，仅表明进程在运行，不探测数据库，供 liveness 检查使用。
	api.GET("/health", hr.livenessCheck)
	// 就绪探针，探测数据库连通性，供 readiness 检查使用，失败时返回 503。
	api.GET("/health/ready", hr.readinessCheck)
}

// livenessCheck 存活探针：进程可达即视为存活。
func (hr *HealthRoutes) livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    response.CodeSuccess,
		"message": "服务正常",
		"data": gin.H{
			"status":  "healthy",
			"service": serviceName,
		},
	})
}

// readinessCheck 就绪探针：数据库连通时视为就绪，失败时返回 503 供编排器摘除流量。
func (hr *HealthRoutes) readinessCheck(c *gin.Context) {
	if err := hr.dbCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    http.StatusServiceUnavailable,
			"message": "服务未就绪",
			"data": gin.H{
				"status":  "unready",
				"service": serviceName,
				"reason":  err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    response.CodeSuccess,
		"message": "服务就绪",
		"data": gin.H{
			"status":  "ready",
			"service": serviceName,
		},
	})
}
