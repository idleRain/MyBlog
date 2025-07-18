// Package service 业务逻辑层
package service

import (
	"MyBlog/internal/repository"
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

// UserService 用户服务接口
type UserService interface {
	CreateUser(req *repository.CreateUserRequest) (*repository.User, error)
	GetUserByID(id uint) (*repository.User, error)
	GetUserList(page, pageSize int) ([]*repository.User, int64, error)
	DeleteUser(id uint) error
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
