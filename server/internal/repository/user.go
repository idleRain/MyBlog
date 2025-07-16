// Package repository 数据访问层
package repository

import (
	"fmt"
	_ "time"

	"MyBlog/pkg/datetime"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint              `json:"id" gorm:"primaryKey"`
	Username  string            `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email     string            `json:"email" gorm:"uniqueIndex;not null;size:100"`
	Password  string            `json:"-" gorm:"not null;size:255"`
	Nickname  string            `json:"nickname" gorm:"size:50"`
	Avatar    string            `json:"avatar" gorm:"size:255"`
	Birthday  datetime.JSONDate `json:"birthday" gorm:"type:date"`
	Status    int               `json:"status" gorm:"default:1;comment:状态 1-正常 0-禁用"`
	CreatedAt datetime.JSONDate `json:"createdAt"`
	UpdatedAt datetime.JSONDate `json:"updatedAt"`
	DeletedAt gorm.DeletedAt    `json:"-" gorm:"index"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string            `json:"username" binding:"required,min=1,max=50"`
	Email    string            `json:"email" binding:"required,email"`
	Password string            `json:"password" binding:"required,min=6,max=100"`
	Nickname string            `json:"nickname" binding:"max=50"`
	Birthday datetime.JSONDate `json:"birthday" binding:"omitempty"`
}

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
	List(offset, limit int) ([]*User, int64, error)
}

// userRepository 用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(user *User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(id uint) (*User, error) {
	var user User
	if err := r.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(username string) (*User, error) {
	var user User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(user *User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}
	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	if err := r.db.Delete(&User{}, id).Error; err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	return nil
}

// List 获取用户列表
func (r *userRepository) List(offset, limit int) ([]*User, int64, error) {
	var users []*User
	var total int64

	// 获取总数
	if err := r.db.Model(&User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户总数失败: %w", err)
	}

	// 获取用户列表
	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	return users, total, nil
}
