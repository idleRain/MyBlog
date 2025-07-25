// Package handler HTTP请求处理层
package handler

import (
	"MyBlog/internal/repository"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser 创建用户 POST /api/users/create
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req repository.CreateUserRequest

	// 绑定和验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 调用服务层创建用户
	user, err := h.userService.CreateUser(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "用户创建成功", user.ToResponse())
}

// GetUserByID 根据ID获取用户 POST /api/users/get
func (h *UserHandler) GetUserByID(c *gin.Context) {
	type GetUserByIDRequest struct {
		ID uint `json:"id" binding:"required,min=1"`
	}

	var req GetUserByIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user, err := h.userService.GetUserByID(req.ID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, user.ToResponse())
}

// GetUserList 获取用户列表 POST /api/users/list
func (h *UserHandler) GetUserList(c *gin.Context) {
	type GetUserListRequest struct {
		Page     int `json:"page" binding:"omitempty,min=1"`
		PageSize int `json:"pageSize" binding:"omitempty,min=1,max=100"`
	}

	var req GetUserListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	users, total, err := h.userService.GetUserList(req.Page, req.PageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 构建响应数据
	data := gin.H{
		"users":    repository.ToResponseList(users),
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
		"pages":    (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	response.Success(c, data)
}

// DeleteUser 删除用户 POST /api/users/delete
func (h *UserHandler) DeleteUser(c *gin.Context) {
	type DeleteUserRequest struct {
		ID uint `json:"id" binding:"required,min=1"`
	}

	var req DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.userService.DeleteUser(req.ID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "用户删除成功", nil)
}

// Login 用户登录 POST /api/users/login
func (h *UserHandler) Login(c *gin.Context) {
	type LoginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	loginResp, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	data := gin.H{
		"user":         loginResp.User.ToResponse(),
		"accessToken":  loginResp.AccessToken,
		"refreshToken": loginResp.RefreshToken,
		"expiresIn":    loginResp.ExpiresIn,
	}

	response.SuccessWithMessage(c, "登录成功", data)
}

// RefreshToken 刷新令牌 POST /api/auth/refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	type RefreshTokenRequest struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	tokenPair, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	data := gin.H{
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"expiresIn":    tokenPair.ExpiresIn,
	}

	response.SuccessWithMessage(c, "令牌刷新成功", data)
}

// Logout 用户登出 POST /api/auth/logout
func (h *UserHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		response.BadRequest(c, "未提供认证令牌")
		return
	}

	// 移除 Bearer 前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if err := h.userService.Logout(token); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "登出成功", nil)
}

// HealthCheck 健康检查 POST /api/health
func (h *UserHandler) HealthCheck(c *gin.Context) {
	data := gin.H{
		"status":  "ok",
		"message": "服务运行正常",
	}
	response.Success(c, data)
}
