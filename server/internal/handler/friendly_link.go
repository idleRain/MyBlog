// Package handler HTTP请求处理层
package handler

import (
	"errors"

	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// FriendlyLinkHandlerInterface 友情链接处理器接口
type FriendlyLinkHandlerInterface interface {
	CreateLink(c *gin.Context)
	UpdateLink(c *gin.Context)
	DeleteLink(c *gin.Context)
	ApproveLink(c *gin.Context)
	HideLink(c *gin.Context)
	RejectLink(c *gin.Context)
	ListLinks(c *gin.Context)
	ListVisibleLinks(c *gin.Context)
}

// FriendlyLinkHandler 友情链接处理器实现
type FriendlyLinkHandler struct {
	linkService service.FriendlyLinkServiceInterface
}

// NewFriendlyLinkHandler 创建友情链接处理器实例
func NewFriendlyLinkHandler(linkService service.FriendlyLinkServiceInterface) FriendlyLinkHandlerInterface {
	return &FriendlyLinkHandler{
		linkService: linkService,
	}
}

// CreateLink 创建友情链接 POST /api/admin/friendly-links/create
func (h *FriendlyLinkHandler) CreateLink(c *gin.Context) {
	var req service.CreateFriendlyLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	link, err := h.linkService.CreateLink(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "友情链接创建成功", link)
}

// UpdateLink 更新友情链接 POST /api/admin/friendly-links/update
func (h *FriendlyLinkHandler) UpdateLink(c *gin.Context) {
	var req service.UpdateFriendlyLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	link, err := h.linkService.UpdateLink(&req)
	if err != nil {
		if errors.Is(err, repository.ErrFriendlyLinkNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "友情链接更新成功", link)
}

// DeleteLink 删除友情链接 POST /api/admin/friendly-links/delete
func (h *FriendlyLinkHandler) DeleteLink(c *gin.Context) {
	type DeleteLinkRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req DeleteLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.linkService.DeleteLink(req.ID, operatorID); err != nil {
		if errors.Is(err, repository.ErrFriendlyLinkNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "友情链接删除成功", nil)
}

// ApproveLink 审核通过友情链接 POST /api/admin/friendly-links/approve
func (h *FriendlyLinkHandler) ApproveLink(c *gin.Context) {
	h.transition(c, h.linkService.ApproveLink, "审核通过")
}

// HideLink 下架友情链接 POST /api/admin/friendly-links/hide
func (h *FriendlyLinkHandler) HideLink(c *gin.Context) {
	h.transition(c, h.linkService.HideLink, "下架成功")
}

// RejectLink 拒绝友情链接 POST /api/admin/friendly-links/reject
func (h *FriendlyLinkHandler) RejectLink(c *gin.Context) {
	h.transition(c, h.linkService.RejectLink, "已拒绝")
}

// ListLinks 分页查询友情链接列表 POST /api/admin/friendly-links/list
func (h *FriendlyLinkHandler) ListLinks(c *gin.Context) {
	var req service.ListFriendlyLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.linkService.ListLinks(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// ListVisibleLinks 获取对外展示的友情链接 POST /api/friendly-links/list
func (h *FriendlyLinkHandler) ListVisibleLinks(c *gin.Context) {
	links, err := h.linkService.ListVisibleLinks()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"links": links})
}

// transition 执行友情链接状态流转类操作的公共逻辑。
func (h *FriendlyLinkHandler) transition(c *gin.Context, action func(id, operatorID uint) error, successMessage string) {
	type TransitionLinkRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req TransitionLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := action(req.ID, operatorID); err != nil {
		if errors.Is(err, repository.ErrFriendlyLinkNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, successMessage, nil)
}
