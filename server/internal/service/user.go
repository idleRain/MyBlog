// Package service 业务逻辑层
package service

import (
	"MyBlog/internal/repository"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// 密码加密相关常量
const (
	// bcrypt 加密成本，推荐值为 12
	BcryptCost = 12
)

// LoginResponse 登录响应
type LoginResponse struct {
	User         *repository.User `json:"user"`
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	ExpiresIn    int64            `json:"expires_in"`
}

// UserService 用户服务接口
type UserService interface {
	CreateUser(req *repository.CreateUserRequest) (*repository.User, error)
	UpdateUser(req *repository.UpdateUserRequest) (*repository.User, error)
	GetUserByID(id uint) (*repository.User, error)
	GetUserList(page, pageSize int) ([]*repository.User, int64, error)
	DeleteUser(id uint) error
	Login(username, password string) (*LoginResponse, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
	Logout(accessToken string) error
	// 权限相关方法
	CanUserManageRole(managerRole, targetRole string) bool
	ValidateRoleTransition(currentRole, newRole string) error
}

// userService 用户服务实现
type userService struct {
	userRepo    repository.UserRepository
	jwtService  JWTService
	rbacService RBACService
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, jwtService JWTService) UserService {
	return &userService{
		userRepo:    userRepo,
		jwtService:  jwtService,
		rbacService: NewRBACService(),
	}
}

// CreateUser 创建用户
func (s *userService) CreateUser(req *repository.CreateUserRequest) (*repository.User, error) {
	// 检查用户名是否已存在
	if existUser, _ := s.userRepo.GetByUsername(req.Username); existUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查邮箱是否已存在
	if existUser, _ := s.userRepo.GetByEmail(req.Email); existUser != nil {
		return nil, fmt.Errorf("邮箱已存在")
	}

	// 验证密码强度
	if err := s.validatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// 使用bcrypt加密密码
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 构建用户对象
	user := &repository.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Nickname: req.Nickname,
		Birthday: req.Birthday,
		Role:     req.Role,
		Status:   1,
	}

	// 如果昵称为空，使用用户名
	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	// 如果角色为空，设置默认角色为 user
	if user.Role == "" {
		user.Role = "user"
	}

	// 验证角色是否有效
	if !s.rbacService.IsValidRole(user.Role) {
		return nil, fmt.Errorf("无效的用户角色: %s", user.Role)
	}

	// 创建用户
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(req *repository.UpdateUserRequest) (*repository.User, error) {
	// 获取现有用户信息
	existingUser, err := s.userRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 检查用户名是否被其他用户占用
	if userByName, _ := s.userRepo.GetByUsername(req.Username); userByName != nil && userByName.ID != req.ID {
		return nil, fmt.Errorf("用户名已被其他用户使用")
	}

	// 检查邮箱是否被其他用户占用
	if userByEmail, _ := s.userRepo.GetByEmail(req.Email); userByEmail != nil && userByEmail.ID != req.ID {
		return nil, fmt.Errorf("邮箱已被其他用户使用")
	}

	// 更新基本信息
	existingUser.Username = req.Username
	existingUser.Email = req.Email
	existingUser.Nickname = req.Nickname
	existingUser.Birthday = req.Birthday
	existingUser.Role = req.Role

	// 更新状态（如果提供了状态字段）
	if req.Status == 0 || req.Status == 1 {
		existingUser.Status = req.Status
	}

	// 如果昵称为空，使用用户名
	if existingUser.Nickname == "" {
		existingUser.Nickname = existingUser.Username
	}

	// 如果角色为空，设置默认角色为 user
	if existingUser.Role == "" {
		existingUser.Role = "user"
	}

	// 验证角色是否有效
	if !s.rbacService.IsValidRole(existingUser.Role) {
		return nil, fmt.Errorf("无效的用户角色: %s", existingUser.Role)
	}

	// 如果提供了新密码，则更新密码
	if req.Password != "" {
		// 验证密码强度
		if err := s.validatePasswordStrength(req.Password); err != nil {
			return nil, err
		}

		// 使用bcrypt加密密码
		hashedPassword, err := s.hashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败: %w", err)
		}

		existingUser.Password = hashedPassword
	}

	// 更新用户
	if err := s.userRepo.Update(existingUser); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return existingUser, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*repository.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserList 获取用户列表
func (s *userService) GetUserList(page, pageSize int) ([]*repository.User, int64, error) {
	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, total, err := s.userRepo.List(offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 删除用户
	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// hashPassword 使用bcrypt加密密码
func (s *userService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(hashedBytes), nil
}

// verifyPassword 验证密码
func (s *userService) verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// validatePasswordStrength 验证密码强度
func (s *userService) validatePasswordStrength(password string) error {
	// 最小长度检查
	if len(password) < 6 {
		return fmt.Errorf("密码长度不能少于6位")
	}

	// 最大长度检查
	if len(password) > 100 {
		return fmt.Errorf("密码长度不能超过100位")
	}

	// 检查是否包含数字
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// 检查是否包含字母
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)

	// 要求至少包含数字和字母
	if !hasNumber || !hasLetter {
		return fmt.Errorf("密码必须包含字母和数字")
	}

	// 检查常见弱密码
	weakPasswords := []string{
		"123456", "password", "123456789", "12345678", "12345",
		"1234567", "1234567890", "qwerty", "abc123", "admin",
		"root", "guest", "test", "user", "demo", "login",
	}

	for _, weak := range weakPasswords {
		if password == weak {
			return fmt.Errorf("密码过于简单，请使用更复杂的密码")
		}
	}

	return nil
}

// Login 用户登录
func (s *userService) Login(username, password string) (*LoginResponse, error) {
	// 先尝试通过用户名查找
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		// 如果用户名未找到，尝试通过邮箱查找
		user, err = s.userRepo.GetByEmail(username)
		if err != nil {
			return nil, fmt.Errorf("用户不存在")
		}
	}

	// 验证密码
	if !s.verifyPassword(password, user.Password) {
		return nil, fmt.Errorf("密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 生成JWT令牌对
	tokenPair, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RefreshToken 刷新令牌
func (s *userService) RefreshToken(refreshToken string) (*TokenPair, error) {
	return s.jwtService.RefreshAccessToken(refreshToken)
}

// Logout 用户登出
func (s *userService) Logout(accessToken string) error {
	return s.jwtService.RevokeToken(accessToken)
}

// CanUserManageRole 检查用户是否可以管理指定角色
func (s *userService) CanUserManageRole(managerRole, targetRole string) bool {
	return s.rbacService.CanManageUser(managerRole, targetRole)
}

// ValidateRoleTransition 验证角色转换是否有效
func (s *userService) ValidateRoleTransition(currentRole, newRole string) error {
	// 检查新角色是否有效
	if !s.rbacService.IsValidRole(newRole) {
		return fmt.Errorf("无效的目标角色: %s", newRole)
	}

	// 检查当前角色是否有效
	if !s.rbacService.IsValidRole(currentRole) {
		return fmt.Errorf("无效的当前角色: %s", currentRole)
	}

	// 超级管理员角色只能由系统设置，不能通过API变更
	if newRole == string(RoleSuperAdmin) {
		return fmt.Errorf("超级管理员角色只能通过系统管理员设置")
	}

	// 从超级管理员降级需要特殊处理
	if currentRole == string(RoleSuperAdmin) {
		return fmt.Errorf("超级管理员角色不能被降级")
	}

	return nil
}
