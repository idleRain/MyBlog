// 数据库种子数据初始化
package database

import (
	"fmt"
	"log"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
	"MyBlog/internal/service"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminSeedOptions 超级管理员种子数据参数
type AdminSeedOptions struct {
	Username string
	Password string
	Email    string
}

// EnsureSuperAdmin 确保指定账号的超级管理员存在，已存在则将其角色提升为超级管理员
func EnsureSuperAdmin(db *gorm.DB, opts AdminSeedOptions) error {
	var existingUser repository.User
	queryErr := db.Where("username = ?", opts.Username).First(&existingUser).Error

	switch {
	case queryErr == nil:
		// 用户已存在，确保其角色为超级管理员
		if existingUser.Role != string(model.RoleSuperAdmin) {
			if err := db.Model(&existingUser).Update("role", string(model.RoleSuperAdmin)).Error; err != nil {
				return fmt.Errorf("提升用户 %s 为超级管理员失败: %w", opts.Username, err)
			}
			log.Printf("用户 %s 已提升为超级管理员", opts.Username)
		} else {
			log.Printf("超级管理员 %s 已存在，无需重复创建", opts.Username)
		}
		return nil

	case queryErr != gorm.ErrRecordNotFound:
		return fmt.Errorf("查询用户 %s 失败: %w", opts.Username, queryErr)
	}

	// 用户不存在，创建新的超级管理员账户
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(opts.Password), service.BcryptCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	superAdmin := repository.User{
		Username: opts.Username,
		Email:    opts.Email,
		Password: string(hashedPassword),
		Nickname: opts.Username,
		Role:     string(model.RoleSuperAdmin),
		Status:   model.UserStatusActive,
	}
	if err := db.Create(&superAdmin).Error; err != nil {
		return fmt.Errorf("创建超级管理员失败: %w", err)
	}
	log.Printf("超级管理员 %s 创建成功", opts.Username)
	return nil
}
