package service

import (
	"strings"
	"testing"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeSettingRepo 设置仓储的测试替身。
type fakeSettingRepo struct {
	repository.SettingRepositoryInterface
	settings []*model.Setting
}

func (f *fakeSettingRepo) GetPublic() ([]*model.Setting, error) {
	var result []*model.Setting
	for _, setting := range f.settings {
		if setting.IsPublic {
			result = append(result, setting)
		}
	}
	return result, nil
}

func (f *fakeSettingRepo) List() ([]*model.Setting, error) {
	return f.settings, nil
}

func (f *fakeSettingRepo) GetByKey(keyName string) (*model.Setting, error) {
	for _, setting := range f.settings {
		if setting.KeyName == keyName {
			return setting, nil
		}
	}
	return nil, repository.ErrSettingNotFound
}

func (f *fakeSettingRepo) Upsert(setting *model.Setting) error {
	for i, existing := range f.settings {
		if existing.KeyName == setting.KeyName {
			f.settings[i] = setting
			return nil
		}
	}
	f.settings = append(f.settings, setting)
	return nil
}

// TestMaskSensitiveSettings 验证敏感设置项输出掩码。
func TestMaskSensitiveSettings(t *testing.T) {
	settings := []*model.Setting{
		{
			KeyName:    model.SettingSiteName,
			Value:      "MyBlog",
			IsSensitive: false,
		},
		{
			KeyName:    model.SettingMailPassword,
			Value:      "secret123",
			IsSensitive: true,
		},
	}

	result := maskSensitiveSettings(settings)

	// 非敏感项保留原值。
	if result[0].Value != "MyBlog" {
		t.Errorf("非敏感项值 = %q, 期望保留原值", result[0].Value)
	}
	// 敏感项输出掩码。
	if !strings.Contains(result[1].Value, "*") {
		t.Errorf("敏感项值应输出掩码，实际为 %q", result[1].Value)
	}
}

// TestUpdateSettingsRejectsReadonly 验证只读设置项禁止更新。
func TestUpdateSettingsRejectsReadonly(t *testing.T) {
	repo := &fakeSettingRepo{
		settings: []*model.Setting{
			{KeyName: "readonly_key", Value: "旧值", IsReadonly: true},
		},
	}
	svc := NewSettingService(repo)

	items := []UpdateSettingItem{
		{KeyName: "readonly_key", Value: "新值"},
	}
	_, err := svc.UpdateSettings(items, 1)
	if err == nil {
		t.Fatal("更新只读设置项应返回错误")
	}
}

// TestUpdateSettingsMissingKey 验证更新不存在的设置项返回错误。
func TestUpdateSettingsMissingKey(t *testing.T) {
	repo := &fakeSettingRepo{settings: []*model.Setting{}}
	svc := NewSettingService(repo)

	items := []UpdateSettingItem{
		{KeyName: "no_such_key", Value: "值"},
	}
	_, err := svc.UpdateSettings(items, 1)
	if err == nil {
		t.Fatal("更新不存在的设置项应返回错误")
	}
}

// TestUpdateSettingsSucceeds 验证批量更新设置项成功。
func TestUpdateSettingsSucceeds(t *testing.T) {
	repo := &fakeSettingRepo{
		settings: []*model.Setting{
			{KeyName: model.SettingSiteName, Value: "旧站名"},
		},
	}
	svc := NewSettingService(repo)

	items := []UpdateSettingItem{
		{KeyName: model.SettingSiteName, Value: "新站名"},
	}
	if _, err := svc.UpdateSettings(items, 1); err != nil {
		t.Fatalf("更新设置失败: %v", err)
	}

	if repo.settings[0].Value != "新站名" {
		t.Errorf("更新后值 = %q, 期望 新站名", repo.settings[0].Value)
	}
}
