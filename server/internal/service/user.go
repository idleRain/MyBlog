// Package service 业务逻辑层
package service

import (
	"MyBlog/internal/repository"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT相关常量
const (
	JWTSecret   = "myblog_jwt_secret_key"
	TokenExpire = 24 * time.Hour // 24小时过期
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// UserService 用户服务接口
type UserService interface {
	CreateUser(req *repository.CreateUserRequest) (*repository.User, error)
	GetUserByID(id uint) (*repository.User, error)
	GetUserList(page, pageSize int) ([]*repository.User, int64, error)
	DeleteUser(id uint) error
	Login(username, password string) (*repository.User, string, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
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

	// 密码加密
	hashedPassword := s.hashPassword(req.Password)

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

// hashPassword 密码加密
func (s *userService) hashPassword(password string) string {
	h := md5.New()
	h.Write([]byte(password + "myblog_salt"))
	return hex.EncodeToString(h.Sum(nil))
}

// Login 用户登录
func (s *userService) Login(username, password string) (*repository.User, string, error) {
	// 先尝试通过用户名查找
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		// 如果用户名未找到，尝试通过邮箱查找
		user, err = s.userRepo.GetByEmail(username)
		if err != nil {
			return nil, "", fmt.Errorf("用户不存在")
		}
	}

	// 验证密码
	hashedPassword := s.hashPassword(password)
	if user.Password != hashedPassword {
		return nil, "", fmt.Errorf("密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, "", fmt.Errorf("用户已被禁用")
	}

	// 生成JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("生成token失败: %w", err)
	}

	return user, token, nil
}

// generateToken 生成JWT token
func (s *userService) generateToken(user *repository.User) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "myblog",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

// ValidateToken 验证JWT token
func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
