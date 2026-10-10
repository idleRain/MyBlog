package router

import (
	"log"
	"net/http"

	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// serviceName 健康检查响应的服务标识，站点英文名变更时同步更新此处。
const serviceName = "IdleRain Blogs API"

// 未就绪原因码，故障细节只写日志。
const (
	reasonDatabaseUnavailable = "database_unavailable"
	reasonUploadsUnwritable   = "uploads_unwritable"
)

// readinessProbe 就绪检查项，name 用于日志定位，reason 为对外原因码。
type readinessProbe struct {
	name   string
	reason string
	probe  func() error
}

// HealthRoutes 健康检查路由模块，各就绪依赖经组合根注入探测函数，路由层不直接依赖具体实现。
type HealthRoutes struct {
	dbCheck      func() error
	uploadsCheck func() error
}

// NewHealthRoutes 创建健康检查路由模块。
// dbCheck 探测数据库连通性，uploadsCheck 探测本地上传目录可写性。
func NewHealthRoutes(dbCheck, uploadsCheck func() error) *HealthRoutes {
	return &HealthRoutes{dbCheck: dbCheck, uploadsCheck: uploadsCheck}
}

// RegisterRoutes 注册健康检查路由。
// POST-Only 规范约束业务接口，健康探针属基础设施端点，编排器只支持 GET 探测，故采用 GET 路由。
func (hr *HealthRoutes) RegisterRoutes(api *gin.RouterGroup) {
	// 存活探针，仅表明进程在运行，不探测下游依赖，供 liveness 检查使用。
	api.GET("/health", hr.livenessCheck)
	// 就绪探针，探测数据库连通性与上传目录可写性，供 readiness 检查使用，任一失败即返回 503。
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

// readinessCheck 就绪探针：数据库连通且上传目录可写时视为就绪，失败时返回 503 供编排器摘除流量。
func (hr *HealthRoutes) readinessCheck(c *gin.Context) {
	probes := []readinessProbe{
		{name: "数据库连通性", reason: reasonDatabaseUnavailable, probe: hr.dbCheck},
		{name: "上传目录可写性", reason: reasonUploadsUnwritable, probe: hr.uploadsCheck},
	}

	for _, item := range probes {
		if err := item.probe(); err != nil {
			log.Printf("[ERROR] 就绪探针检查失败：检查项=%s err=%v", item.name, err)
			hr.respondUnready(c, item.reason)
			return
		}
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

// respondUnready 返回未就绪响应。
func (hr *HealthRoutes) respondUnready(c *gin.Context, reason string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"code":    http.StatusServiceUnavailable,
		"message": "服务未就绪",
		"data": gin.H{
			"status":  "unready",
			"service": serviceName,
			"reason":  reason,
		},
	})
}
