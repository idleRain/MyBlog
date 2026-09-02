// Package handler HTTP请求处理层
package handler

import (
	"errors"

	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// CommentHandlerInterface 评论处理器接口
type CommentHandlerInterface interface {
	CreateComment(c *gin.Context)
	GetCommentsByArticle(c *gin.Context)
	LikeComment(c *gin.Context)
	UnlikeComment(c *gin.Context)
	ApproveComment(c *gin.Context)
	RejectComment(c *gin.Context)
	MarkCommentSpam(c *gin.Context)
	TrashComment(c *gin.Context)
	DeleteComment(c *gin.Context)
	ListComments(c *gin.Context)
}

// CommentHandler 评论处理器实现
type CommentHandler struct {
	commentService service.CommentServiceInterface
}

// NewCommentHandler 创建评论处理器实例
func NewCommentHandler(commentService service.CommentServiceInterface) CommentHandlerInterface {
	return &CommentHandler{
		commentService: commentService,
	}
}

// CreateComment 创建评论 POST /api/comments/create
func (h *CommentHandler) CreateComment(c *gin.Context) {
	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	comment, err := h.commentService.CreateComment(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "评论提交成功", comment)
}

// GetCommentsByArticle 获取文章评论列表 POST /api/comments/list
func (h *CommentHandler) GetCommentsByArticle(c *gin.Context) {
	type GetCommentsByArticleRequest struct {
		ArticleID uint `json:"articleId" binding:"required"`
		service.ListCommentsRequest
	}

	var req GetCommentsByArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.commentService.GetCommentsByArticle(req.ArticleID, &req.ListCommentsRequest)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// LikeComment 点赞评论 POST /api/comments/like
func (h *CommentHandler) LikeComment(c *gin.Context) {
	type LikeCommentRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req LikeCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.commentService.LikeComment(req.ID, userID); err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "点赞成功", nil)
}

// UnlikeComment 取消点赞评论 POST /api/comments/unlike
func (h *CommentHandler) UnlikeComment(c *gin.Context) {
	type UnlikeCommentRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req UnlikeCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.commentService.UnlikeComment(req.ID, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "取消点赞成功", nil)
}

// ApproveComment 审核通过评论 POST /api/admin/comments/approve
func (h *CommentHandler) ApproveComment(c *gin.Context) {
	h.moderate(c, func(id, operatorID uint) error {
		return h.commentService.ApproveComment(id, operatorID)
	}, "审核通过")
}

// RejectComment 拒绝评论 POST /api/admin/comments/reject
func (h *CommentHandler) RejectComment(c *gin.Context) {
	h.moderate(c, func(id, operatorID uint) error {
		return h.commentService.RejectComment(id, operatorID)
	}, "拒绝评论")
}

// MarkCommentSpam 标记为垃圾评论 POST /api/admin/comments/spam
func (h *CommentHandler) MarkCommentSpam(c *gin.Context) {
	h.moderate(c, func(id, operatorID uint) error {
		return h.commentService.MarkCommentSpam(id, operatorID)
	}, "标记垃圾评论")
}

// TrashComment 移入回收站 POST /api/admin/comments/trash
func (h *CommentHandler) TrashComment(c *gin.Context) {
	h.moderate(c, func(id, operatorID uint) error {
		return h.commentService.TrashComment(id, operatorID)
	}, "移入回收站")
}

// DeleteComment 删除评论 POST /api/admin/comments/delete
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	type DeleteCommentRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req DeleteCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	operatorID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.commentService.DeleteComment(req.ID, operatorID); err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "评论删除成功", nil)
}

// ListComments 管理端评论列表 POST /api/admin/comments/list
func (h *CommentHandler) ListComments(c *gin.Context) {
	var req service.AdminListCommentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.commentService.ListComments(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// moderate 执行评论状态流转类操作的公共逻辑。
func (h *CommentHandler) moderate(c *gin.Context, action func(id, operatorID uint) error, successMessage string) {
	type ModerateCommentRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req ModerateCommentRequest
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
		if errors.Is(err, repository.ErrCommentNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, successMessage, nil)
}
