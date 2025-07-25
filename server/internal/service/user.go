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
	GetUserByID(id uint) (*repository.User, error)
	GetUserList(page, pageSize int) ([]*repository.User, int64, error)
	DeleteUser(id uint) error
	Login(username, password string) (*LoginResponse, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
	Logout(accessToken string) error
}

// userService 用户服务实现
type userService struct {
	userRepo   repository.UserRepository
	jwtService JWTService
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, jwtService JWTService) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtService: jwtService,
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
		Status:   1,
	}

	// 如果昵称为空，使用用户名
	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	// 创建用户
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
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
