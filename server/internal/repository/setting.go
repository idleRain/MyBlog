// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"

	"gorm.io/gorm"
)

// ErrSettingNotFound 设置项不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrSettingNotFound = errors.New("设置项不存在")

// SettingRepositoryInterface 设置仓储接口
type SettingRepositoryInterface interface {
	// 基础操作
	GetByKey(keyName string) (*model.Setting, error)
	GetPublic() ([]*model.Setting, error)
	List() ([]*model.Setting, error)
	Upsert(setting *model.Setting) error
}

// SettingRepository 设置仓储实现
type SettingRepository struct {
	db *gorm.DB
}

// NewSettingRepository 创建设置仓储实例
func NewSettingRepository(db *gorm.DB) SettingRepositoryInterface {
	return &SettingRepository{db: db}
}

// GetByKey 根据键名获取设置项。
func (r *SettingRepository) GetByKey(keyName string) (*model.Setting, error) {
	var setting model.Setting
	if err := r.db.Where("key_name = ?", keyName).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSettingNotFound
		}
		return nil, fmt.Errorf("查询设置项失败: %w", err)
	}
	return &setting, nil
}

// GetPublic 获取全部公开设置项。
func (r *SettingRepository) GetPublic() ([]*model.Setting, error) {
	var settings []*model.Setting
	if err := r.db.Where("is_public = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("查询公开设置项失败: %w", err)
	}
	return settings, nil
}

// List 获取全部设置项。
func (r *SettingRepository) List() ([]*model.Setting, error) {
	var settings []*model.Setting
	if err := r.db.Order("group_name ASC, sort_order ASC, id ASC").Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("查询设置项失败: %w", err)
	}
	return settings, nil
}

// Upsert 按键名新增或更新设置项。
func (r *SettingRepository) Upsert(setting *model.Setting) error {
	if err := r.db.Where("key_name = ?", setting.KeyName).
		Assign(setting).
		FirstOrCreate(setting).Error; err != nil {
		return fmt.Errorf("保存设置项失败: %w", err)
	}
	return nil
}
