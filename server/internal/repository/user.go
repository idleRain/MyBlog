// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"
	"time"

	"MyBlog/internal/domain"

	"gorm.io/gorm"
)

// UserRepository 用户仓库接口，实体类型统一使用 domain.User。
type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uint) (*domain.User, error)
	GetByUsername(username string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id uint) error
	List(offset, limit int, keyword string) ([]*domain.User, int64, error)
	IncrementLoginFailures(id uint) error
	LockUserAfterFailures(id uint, maxAttempts uint, lockedUntil time.Time) (bool, error)
	ResetLoginFailures(id uint) error
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
func (r *userRepository) Create(user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(user *domain.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}
	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	if err := r.db.Delete(&domain.User{}, id).Error; err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	return nil
}

// List 获取用户列表，keyword 非空时按用户名、邮箱或昵称模糊过滤。
func (r *userRepository) List(offset, limit int, keyword string) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	query := r.db.Model(&domain.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where(
			"username LIKE ? OR email LIKE ? OR nickname LIKE ?",
			like, like, like,
		)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户总数失败: %w", err)
	}

	// 获取用户列表
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	return users, total, nil
}

// IncrementLoginFailures 原子累计连续登录失败次数。
// 使用数据库侧自增表达式，并发登录时各次失败均被计入，不受读改写覆盖影响。
func (r *userRepository) IncrementLoginFailures(id uint) error {
	if err := r.db.Model(&domain.User{}).Where("id = ?", id).
		UpdateColumn("failed_login_count", gorm.Expr("failed_login_count + 1")).Error; err != nil {
		return fmt.Errorf("累计登录失败计数失败: %w", err)
	}
	return nil
}

// LockUserAfterFailures 在失败计数达到阈值时写入锁定截止时间。
// 条件更新自身完成阈值判断，计数未达阈值时不写锁定，取消读回值带来的竞态窗口。
// 返回是否真正写入锁定，供调用方选择对应响应文案。
func (r *userRepository) LockUserAfterFailures(id uint, maxAttempts uint, lockedUntil time.Time) (bool, error) {
	result := r.db.Model(&domain.User{}).
		Where("id = ? AND failed_login_count >= ?", id, maxAttempts).
		UpdateColumn("locked_until", lockedUntil)
	if result.Error != nil {
		return false, fmt.Errorf("写入账户锁定失败: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}

// ResetLoginFailures 登录成功后原子清零失败计数并清除锁定截止时间。
// 仅更新登录相关列，避免全字段覆盖写回其他字段的陈旧内存值。
func (r *userRepository) ResetLoginFailures(id uint) error {
	if err := r.db.Model(&domain.User{}).Where("id = ?", id).
		Select("failed_login_count", "locked_until").
		Updates(domain.User{FailedLoginCount: 0, LockedUntil: nil}).Error; err != nil {
		return fmt.Errorf("重置登录失败计数失败: %w", err)
	}
	return nil
}
