package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthRoutes 健康检查路由模块
type HealthRoutes struct{}

// NewHealthRoutes 创建健康检查路由模块
func NewHealthRoutes() *HealthRoutes {
	return &HealthRoutes{}
}

// RegisterRoutes 注册健康检查路由
func (hr *HealthRoutes) RegisterRoutes(api *gin.RouterGroup) {
	// 健康检查路由
	api.POST("/health", hr.healthCheck)

	// 可以添加更多的健康检查相关路由
	// api.GET("/version", hr.versionCheck)
	// api.GET("/status", hr.statusCheck)
}

// healthCheck 健康检查处理函数
func (hr *HealthRoutes) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "服务正常",
		"data": gin.H{
			"status":  "healthy",
			"service": "MyBlog API",
		},
	})
}

// versionCheck 版本检查处理函数（示例）
func (hr *HealthRoutes) versionCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "版本信息",
		"data": gin.H{
			"version": "1.0.0",
			"build":   "dev",
		},
	})
}

// statusCheck 状态检查处理函数（示例）
func (hr *HealthRoutes) statusCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "服务状态",
		"data": gin.H{
			"uptime":   "running",
			"database": "connected",
			"cache":    "available",
		},
	})
}
