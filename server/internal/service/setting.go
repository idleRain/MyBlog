// Package service 业务逻辑层
package service

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// SettingServiceInterface 设置服务接口
type SettingServiceInterface interface {
	// 查询操作
	GetPublicSettings() ([]*model.Setting, error)
	ListSettings() ([]*model.Setting, error)

	// 更新操作
	UpdateSettings(items []UpdateSettingItem, operatorID uint) ([]*model.Setting, error)
}

// UpdateSettingItem 单个设置项的更新内容
type UpdateSettingItem struct {
	KeyName string `json:"keyName" binding:"required,max=100"`
	Value   string `json:"value"`
}

// SettingService 设置服务实现
type SettingService struct {
	settingRepo repository.SettingRepositoryInterface
}

// NewSettingService 创建设置服务实例
func NewSettingService(settingRepo repository.SettingRepositoryInterface) SettingServiceInterface {
	return &SettingService{
		settingRepo: settingRepo,
	}
}

// GetPublicSettings 获取全部公开设置项，输出前做脱敏处理。
func (s *SettingService) GetPublicSettings() ([]*model.Setting, error) {
	settings, err := s.settingRepo.GetPublic()
	if err != nil {
		return nil, err
	}
	return maskSensitiveSettings(settings), nil
}

// ListSettings 获取全部设置项，输出前做脱敏处理。
func (s *SettingService) ListSettings() ([]*model.Setting, error) {
	settings, err := s.settingRepo.List()
	if err != nil {
		return nil, err
	}
	return maskSensitiveSettings(settings), nil
}

// UpdateSettings 批量更新设置项，只读项禁止修改。
func (s *SettingService) UpdateSettings(items []UpdateSettingItem, operatorID uint) ([]*model.Setting, error) {
	if len(items) == 0 {
		return nil, errors.New("更新项不能为空")
	}

	for _, item := range items {
		// 读取现有设置项，不存在或只读时拒绝更新。
		setting, err := s.settingRepo.GetByKey(item.KeyName)
		if err != nil {
			if errors.Is(err, repository.ErrSettingNotFound) {
				return nil, fmt.Errorf("设置项 %s 不存在", item.KeyName)
			}
			return nil, err
		}

		if setting.IsReadonly {
			return nil, fmt.Errorf("设置项 %s 为只读，不可修改", item.KeyName)
		}

		setting.Value = item.Value
		setting.UpdatedBy = &operatorID
		if err := s.settingRepo.Upsert(setting); err != nil {
			return nil, err
		}
	}

	return s.ListSettings()
}

// maskSensitiveSettings 对敏感设置项输出掩码，避免泄露密码与密钥。
func maskSensitiveSettings(settings []*model.Setting) []*model.Setting {
	for _, setting := range settings {
		setting.Value = setting.GetDisplayValue()
	}
	return settings
}
