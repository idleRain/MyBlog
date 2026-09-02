// Package handler HTTP请求处理层
package handler

import (
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// StatsHandlerInterface 站点统计处理器接口
type StatsHandlerInterface interface {
	GetOverview(c *gin.Context)
	GetArticleViewsTrend(c *gin.Context)
}

// StatsHandler 站点统计处理器实现
type StatsHandler struct {
	statsService service.StatsServiceInterface
}

// NewStatsHandler 创建站点统计处理器实例
func NewStatsHandler(statsService service.StatsServiceInterface) StatsHandlerInterface {
	return &StatsHandler{
		statsService: statsService,
	}
}

// GetOverview 获取站点统计概览 POST /api/admin/stats/overview
func (h *StatsHandler) GetOverview(c *gin.Context) {
	overview, err := h.statsService.GetOverview()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, overview)
}

// GetArticleViewsTrend 获取文章浏览量趋势 POST /api/admin/stats/articles
func (h *StatsHandler) GetArticleViewsTrend(c *gin.Context) {
	type GetArticleViewsTrendRequest struct {
		Days int `json:"days" binding:"omitempty,min=1,max=90"`
	}

	var req GetArticleViewsTrendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	trend, err := h.statsService.GetArticleViewsTrend(req.Days)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, trend)
}
