// Package handler HTTP请求处理层
package handler

import (
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// DictHandlerInterface 字典处理器接口
type DictHandlerInterface interface {
	// app 端公开接口
	GetAllDicts(c *gin.Context)
	GetDictByType(c *gin.Context)

	// 管理端：字典类型
	CreateDictType(c *gin.Context)
	UpdateDictType(c *gin.Context)
	DeleteDictType(c *gin.Context)
	ListDictTypes(c *gin.Context)

	// 管理端：字典项
	CreateDictItem(c *gin.Context)
	UpdateDictItem(c *gin.Context)
	DeleteDictItem(c *gin.Context)
	ListDictItems(c *gin.Context)
}

// DictHandler 字典处理器实现
type DictHandler struct {
	dictService service.DictServiceInterface
}

// NewDictHandler 创建字典处理器实例
func NewDictHandler(dictService service.DictServiceInterface) DictHandlerInterface {
	return &DictHandler{
		dictService: dictService,
	}
}

// deleteDictRequest 按 ID 定位字典资源的请求体，类型与字典项删除共用。
type deleteDictRequest struct {
	ID uint `json:"id" binding:"required"`
}

// GetAllDicts 全量已生效字典 POST /api/dicts/all
func (h *DictHandler) GetAllDicts(c *gin.Context) {
	groups, err := h.dictService.ListEnabledDicts()
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	respondLocalizedDictGroups(c, groups)
}

// GetDictByType 按字典码查询单个已生效字典 POST /api/dicts/:type
func (h *DictHandler) GetDictByType(c *gin.Context) {
	code := c.Param("type")
	if code == "" {
		response.BadRequest(c, "字典码不能为空")
		return
	}

	group, err := h.dictService.ListEnabledItemsByTypeCode(code)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	respondLocalizedDictGroup(c, group)
}

// CreateDictType 创建字典类型 POST /api/admin/dicts/types/create
func (h *DictHandler) CreateDictType(c *gin.Context) {
	var req service.CreateDictTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	dictType, err := h.dictService.CreateDictType(&req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	respondLocalizedDictTypeMessage(c, "字典类型创建成功", dictType)
}

// UpdateDictType 更新字典类型 POST /api/admin/dicts/types/update
func (h *DictHandler) UpdateDictType(c *gin.Context) {
	var req service.UpdateDictTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	dictType, err := h.dictService.UpdateDictType(&req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	respondLocalizedDictTypeMessage(c, "字典类型更新成功", dictType)
}

// DeleteDictType 删除字典类型 POST /api/admin/dicts/types/delete
func (h *DictHandler) DeleteDictType(c *gin.Context) {
	var req deleteDictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.dictService.DeleteDictType(req.ID); err != nil {
		HandleServiceError(c, err)
		return
	}

	response.SuccessWithMessage(c, "字典类型删除成功", nil)
}

// ListDictTypes 分页查询字典类型列表 POST /api/admin/dicts/types/list
func (h *DictHandler) ListDictTypes(c *gin.Context) {
	var req service.ListDictTypesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.dictService.ListDictTypes(&req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	respondLocalizedDictTypeList(c, result)
}

// CreateDictItem 创建字典项 POST /api/admin/dicts/items/create
func (h *DictHandler) CreateDictItem(c *gin.Context) {
	var req service.CreateDictItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	item, err := h.dictService.CreateDictItem(&req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	respondLocalizedDictItemMessage(c, "字典项创建成功", item)
}

// UpdateDictItem 更新字典项 POST /api/admin/dicts/items/update
func (h *DictHandler) UpdateDictItem(c *gin.Context) {
	var req service.UpdateDictItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	item, err := h.dictService.UpdateDictItem(&req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	respondLocalizedDictItemMessage(c, "字典项更新成功", item)
}

// DeleteDictItem 删除字典项 POST /api/admin/dicts/items/delete
func (h *DictHandler) DeleteDictItem(c *gin.Context) {
	var req deleteDictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.dictService.DeleteDictItem(req.ID); err != nil {
		HandleServiceError(c, err)
		return
	}

	response.SuccessWithMessage(c, "字典项删除成功", nil)
}

// ListDictItems 分页查询字典项列表 POST /api/admin/dicts/items/list
func (h *DictHandler) ListDictItems(c *gin.Context) {
	var req service.ListDictItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.dictService.ListDictItems(&req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	respondLocalizedDictItemList(c, result)
}
