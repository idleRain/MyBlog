// Package handler HTTP请求处理层
package handler

import (
	"errors"
	"io"
	"strings"

	"MyBlog/internal/domain"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// bearerTokenPrefix 认证令牌的 Bearer 前缀，用于从请求头解析令牌。
const bearerTokenPrefix = "Bearer "

// LogoutRequest 登出请求体，刷新令牌可选提交。
// 上限 512 字符为保护性约束，不透明令牌实际长度为 32 个字符。
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"omitempty,max=512"`
}

// UserHandlerInterface 用户处理器接口，由 router 层消费并注入。
type UserHandlerInterface interface {
	CreateUser(c *gin.Context)     // POST /api/users/create - JSON格式
	UpdateUser(c *gin.Context)     // POST /api/users/update - JSON格式
	GetUserByID(c *gin.Context)    // POST /api/users/get - JSON格式
	GetUserList(c *gin.Context)    // POST /api/users/list - JSON格式，用于复杂参数查询
	DeleteUser(c *gin.Context)     // POST /api/users/delete - JSON格式
	Login(c *gin.Context)          // POST /api/users/login - JSON格式
	CreateSession(c *gin.Context)  // POST /api/auth/session - 建立并续期会话 Cookie
	RefreshToken(c *gin.Context)   // POST /api/auth/refresh - JSON格式，Header 通道过渡保留
	Logout(c *gin.Context)         // POST /api/auth/logout - Header 或会话 Cookie
	GetProfile(c *gin.Context)     // POST /api/users/profile - 当前用户资料
	UpdateProfile(c *gin.Context)  // POST /api/users/profile/update - JSON格式
	ChangePassword(c *gin.Context) // POST /api/users/changePassword - JSON格式
}

// UserHandler 用户处理器
type UserHandler struct {
	userService   service.UserService
	sessionCookie SessionCookieConfig
}

// NewUserHandler 创建用户处理器实例，会话 Cookie 属性由组合根从令牌配置换算注入。
func NewUserHandler(userService service.UserService, sessionCookie SessionCookieConfig) UserHandlerInterface {
	return &UserHandler{
		userService:   userService,
		sessionCookie: sessionCookie,
	}
}

// CreateUser 创建用户 POST /api/users/create
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req domain.CreateUserRequest

	// 绑定和验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 调用服务层创建用户
	user, err := h.userService.CreateUser(&req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.SuccessWithMessage(c, "用户创建成功", user.ToResponse())
}

// UpdateUser 更新用户信息 POST /api/users/update
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req domain.UpdateUserRequest

	// 绑定和验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取当前操作用户信息
	currentUserRole, exists := c.Get("userRole")
	if !exists {
		response.Unauthorized(c, "无法获取用户权限信息")
		return
	}

	// 获取目标用户当前信息，用于角色权限验证。
	targetUser, err := h.userService.GetUserByID(req.ID)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	// 检查是否可以管理目标用户角色
	if !h.userService.CanUserManageRole(currentUserRole.(string), targetUser.Role) {
		response.Forbidden(c, "权限不足，无法管理该角色的用户")
		return
	}

	// 如果要修改角色，需要验证角色转换
	if req.Role != "" && req.Role != targetUser.Role {
		if err := h.userService.ValidateRoleTransition(targetUser.Role, req.Role); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		// 再次检查是否可以分配新角色
		if !h.userService.CanUserManageRole(currentUserRole.(string), req.Role) {
			response.Forbidden(c, "权限不足，无法分配该角色")
			return
		}
	}

	// 调用服务层更新用户
	user, err := h.userService.UpdateUser(&req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.SuccessWithMessage(c, "用户更新成功", user.ToResponse())
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
		HandleServiceError(c, err)
		return
	}

	response.Success(c, user.ToResponse())
}

// GetUserList 获取用户列表 POST /api/users/list
func (h *UserHandler) GetUserList(c *gin.Context) {
	type GetUserListRequest struct {
		Page     int    `json:"page" binding:"omitempty,min=1"`
		PageSize int    `json:"pageSize" binding:"omitempty,min=1,max=100"`
		Keyword  string `json:"keyword"`
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

	users, total, err := h.userService.GetUserList(req.Page, req.PageSize, req.Keyword)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 构建响应数据
	data := gin.H{
		"users":    domain.ToResponseList(users),
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

	// 获取当前操作用户信息
	currentUserRole, exists := c.Get("userRole")
	if !exists {
		response.Unauthorized(c, "无法获取用户权限信息")
		return
	}

	currentUserID, userIDExists := c.Get("userID")
	if !userIDExists {
		response.Unauthorized(c, "无法获取用户ID信息")
		return
	}

	// 防止用户删除自己
	if currentUserID.(uint) == req.ID {
		response.BadRequest(c, "不能删除自己的账户")
		return
	}

	// 获取目标用户信息
	targetUser, err := h.userService.GetUserByID(req.ID)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	// 检查是否可以删除目标用户
	if !h.userService.CanUserManageRole(currentUserRole.(string), targetUser.Role) {
		response.Forbidden(c, "权限不足，无法删除该角色的用户")
		return
	}

	if err := h.userService.DeleteUser(req.ID); err != nil {
		HandleServiceError(c, err)
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
		"permissions":  loginResp.Permissions,
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

	// TokenPair 的字段名即契约，直接序列化避免手工拼装与契约漂移。
	response.SuccessWithMessage(c, "令牌刷新成功", tokenPair)
}

// Logout 用户登出 POST /api/auth/logout
// 双轨撤销：Authorization 头令牌与会话 Cookie 任一存在即处理，过渡期两者可并存。
// 刷新令牌经请求体可选提交，Cookie 会话存在时随 Cookie 一并撤销并指示浏览器清除。
func (h *UserHandler) Logout(c *gin.Context) {
	headerToken := strings.TrimPrefix(c.GetHeader("Authorization"), bearerTokenPrefix)

	cookieAccess, cookieRefresh := sessionCookiesToken(c)
	if headerToken == "" && cookieAccess == "" {
		response.BadRequest(c, "未提供认证令牌")
		return
	}

	// 请求体可省略，客户端未提交请求体时跳过绑定错误，仅撤销已提交的令牌。
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// Header 通道按原语义撤销，Cookie 通道将访问与刷新令牌一并撤销。
	if headerToken != "" {
		if err := h.userService.Logout(headerToken, req.RefreshToken); err != nil {
			HandleServiceError(c, err)
			return
		}
	}
	if cookieAccess != "" {
		if err := h.userService.Logout(cookieAccess, cookieRefresh); err != nil {
			HandleServiceError(c, err)
			return
		}
	}

	h.clearSessionCookies(c)
	response.SuccessWithMessage(c, "登出成功", nil)
}

// GetProfile 获取当前登录用户资料 POST /api/users/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, user)
}

// UpdateProfile 更新当前登录用户资料 POST /api/users/profile/update
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, user)
}

// ChangePassword 修改当前登录用户密码 POST /api/users/changePassword
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	if err := h.userService.ChangePassword(userID, &req); err != nil {
		HandleServiceError(c, err)
		return
	}

	response.SuccessWithMessage(c, "密码修改成功", nil)
}

// HealthCheck 健康检查 POST /api/health
func (h *UserHandler) HealthCheck(c *gin.Context) {
	data := gin.H{
		"status":  "ok",
		"message": "服务运行正常",
	}
	response.Success(c, data)
}
