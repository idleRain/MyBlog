// Package handler HTTP请求处理层
package handler

import (
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserFollowHandlerInterface 用户关注处理器接口
type UserFollowHandlerInterface interface {
	Follow(c *gin.Context)
	Unfollow(c *gin.Context)
	ListFollowers(c *gin.Context)
	ListFollowing(c *gin.Context)
}

// UserFollowHandler 用户关注处理器实现
type UserFollowHandler struct {
	followService service.UserFollowServiceInterface
}

// NewUserFollowHandler 创建用户关注处理器实例
func NewUserFollowHandler(followService service.UserFollowServiceInterface) UserFollowHandlerInterface {
	return &UserFollowHandler{
		followService: followService,
	}
}

// Follow 关注用户 POST /api/users/follow
func (h *UserFollowHandler) Follow(c *gin.Context) {
	type FollowRequest struct {
		FollowingID uint `json:"followingId" binding:"required"`
	}

	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.followService.Follow(userID, req.FollowingID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "关注成功", nil)
}

// Unfollow 取消关注 POST /api/users/unfollow
func (h *UserFollowHandler) Unfollow(c *gin.Context) {
	type UnfollowRequest struct {
		FollowingID uint `json:"followingId" binding:"required"`
	}

	var req UnfollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.followService.Unfollow(userID, req.FollowingID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "取消关注成功", nil)
}

// ListFollowers 分页查询粉丝列表 POST /api/users/followers
func (h *UserFollowHandler) ListFollowers(c *gin.Context) {
	type ListFollowersRequest struct {
		UserID uint `json:"userId" binding:"required"`
		service.ListFollowsRequest
	}

	var req ListFollowersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.followService.ListFollowers(req.UserID, &req.ListFollowsRequest)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// ListFollowing 分页查询关注列表 POST /api/users/following
func (h *UserFollowHandler) ListFollowing(c *gin.Context) {
	type ListFollowingRequest struct {
		UserID uint `json:"userId" binding:"required"`
		service.ListFollowsRequest
	}

	var req ListFollowingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.followService.ListFollowing(req.UserID, &req.ListFollowsRequest)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}
