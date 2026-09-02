// Package handler HTTP请求处理层
package handler

import (
	"errors"

	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// CategoryHandlerInterface 分类处理器接口
type CategoryHandlerInterface interface {
	CreateCategory(c *gin.Context)
	UpdateCategory(c *gin.Context)
	DeleteCategory(c *gin.Context)
	GetCategory(c *gin.Context)
	ListCategories(c *gin.Context)
	GetCategoryTree(c *gin.Context)
}

// CategoryHandler 分类处理器实现
type CategoryHandler struct {
	categoryService service.CategoryServiceInterface
}

// NewCategoryHandler 创建分类处理器实例
func NewCategoryHandler(categoryService service.CategoryServiceInterface) CategoryHandlerInterface {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// CreateCategory 创建分类 POST /api/admin/categories/create
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req service.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "无法获取操作者信息")
		return
	}

	category, err := h.categoryService.CreateCategory(&req, operatorID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "分类创建成功", category)
}

// UpdateCategory 更新分类 POST /api/admin/categories/update
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	var req service.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "无法获取操作者信息")
		return
	}

	category, err := h.categoryService.UpdateCategory(&req, operatorID)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "分类更新成功", category)
}

// DeleteCategory 删除分类 POST /api/admin/categories/delete
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	type DeleteCategoryRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req DeleteCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "无法获取操作者信息")
		return
	}

	if err := h.categoryService.DeleteCategory(req.ID, operatorID); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "分类删除成功", nil)
}

// GetCategory 根据ID获取分类 POST /api/categories/get
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	type GetCategoryRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req GetCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	category, err := h.categoryService.GetCategory(req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, category)
}

// ListCategories 分页查询分类列表 POST /api/admin/categories/list
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	var req service.ListCategoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.categoryService.ListCategories(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetCategoryTree 获取分类树 POST /api/categories/tree
func (h *CategoryHandler) GetCategoryTree(c *gin.Context) {
	tree, err := h.categoryService.GetCategoryTree()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"tree": tree})
}
