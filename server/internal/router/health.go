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

	// 预留扩展更多健康检查相关路由的位置
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
