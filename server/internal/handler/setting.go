// Package handler HTTP请求处理层
package handler

import (
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// SettingHandlerInterface 设置处理器接口
type SettingHandlerInterface interface {
	GetPublicSettings(c *gin.Context)
	ListSettings(c *gin.Context)
	UpdateSettings(c *gin.Context)
}

// SettingHandler 设置处理器实现
type SettingHandler struct {
	settingService service.SettingServiceInterface
}

// NewSettingHandler 创建设置处理器实例
func NewSettingHandler(settingService service.SettingServiceInterface) SettingHandlerInterface {
	return &SettingHandler{
		settingService: settingService,
	}
}

// GetPublicSettings 获取公开设置项 POST /api/settings/public
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"settings": settings})
}

// ListSettings 获取全部设置项 POST /api/admin/settings/list
func (h *SettingHandler) ListSettings(c *gin.Context) {
	settings, err := h.settingService.ListSettings()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"settings": settings})
}

// UpdateSettings 批量更新设置项 POST /api/admin/settings/update
func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	type UpdateSettingsRequest struct {
		Items []service.UpdateSettingItem `json:"items" binding:"required,min=1,dive"`
	}

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	settings, err := h.settingService.UpdateSettings(req.Items, operatorID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "设置更新成功", gin.H{"settings": settings})
}
