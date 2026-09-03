// Package handler HTTP请求处理层
package handler

import (
	"errors"

	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// TagHandlerInterface 标签处理器接口
type TagHandlerInterface interface {
	CreateTag(c *gin.Context)
	UpdateTag(c *gin.Context)
	DeleteTag(c *gin.Context)
	GetTag(c *gin.Context)
	ListTags(c *gin.Context)
	GetPopularTags(c *gin.Context)
	ListAllTags(c *gin.Context)
}

// TagHandler 标签处理器实现
type TagHandler struct {
	tagService service.TagServiceInterface
}

// NewTagHandler 创建标签处理器实例
func NewTagHandler(tagService service.TagServiceInterface) TagHandlerInterface {
	return &TagHandler{
		tagService: tagService,
	}
}

// CreateTag 创建标签 POST /api/admin/tags/create
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req service.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "无法获取操作者信息")
		return
	}

	tag, err := h.tagService.CreateTag(&req, operatorID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "标签创建成功", tag)
}

// UpdateTag 更新标签 POST /api/admin/tags/update
func (h *TagHandler) UpdateTag(c *gin.Context) {
	var req service.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "无法获取操作者信息")
		return
	}

	tag, err := h.tagService.UpdateTag(&req, operatorID)
	if err != nil {
		if errors.Is(err, repository.ErrTagNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "标签更新成功", tag)
}

// DeleteTag 删除标签 POST /api/admin/tags/delete
func (h *TagHandler) DeleteTag(c *gin.Context) {
	type DeleteTagRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req DeleteTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "无法获取操作者信息")
		return
	}

	if err := h.tagService.DeleteTag(req.ID, operatorID); err != nil {
		if errors.Is(err, repository.ErrTagNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "标签删除成功", nil)
}

// GetTag 根据ID获取标签 POST /api/tags/get
func (h *TagHandler) GetTag(c *gin.Context) {
	type GetTagRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req GetTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	tag, err := h.tagService.GetTag(req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrTagNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, tag)
}

// ListTags 分页查询标签列表 POST /api/admin/tags/list
func (h *TagHandler) ListTags(c *gin.Context) {
	var req service.ListTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.tagService.ListTags(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetPopularTags 获取热门标签 POST /api/tags/popular
func (h *TagHandler) GetPopularTags(c *gin.Context) {
	type GetPopularTagsRequest struct {
		Limit int `json:"limit" binding:"omitempty,min=1,max=50"`
	}

	var req GetPopularTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	tags, err := h.tagService.GetPopularTags(req.Limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"tags": tags})
}

// ListAllTags 获取全部标签 POST /api/tags/list
// 需登录且具备文章读取权限，供文章编辑时选择标签使用。
func (h *TagHandler) ListAllTags(c *gin.Context) {
	tags, err := h.tagService.ListAllTags()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"tags": tags})
}
