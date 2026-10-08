package service

import (
	"errors"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeSettingRepo 设置仓储的测试替身。
type fakeSettingRepo struct {
	repository.SettingRepositoryInterface
	settings []*model.Setting
	// batchErr 模拟批量写回失败，返回前不加任何改动，等价于事务回滚的整体失败。
	batchErr error
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
			// 返回独立副本，对齐真实查询一次一快照的语义，避免校验期间的赋值污染仓储数据。
			cloned := *setting
			return &cloned, nil
		}
	}
	return nil, domain.ErrSettingNotFound
}

func (f *fakeSettingRepo) Upsert(setting *model.Setting) error {
	return f.applySetting(setting)
}

// UpsertBatch 批量写回的替身实现，可注入整体失败验证半程不生效。
func (f *fakeSettingRepo) UpsertBatch(items []*model.Setting) error {
	if f.batchErr != nil {
		return f.batchErr
	}
	for _, setting := range items {
		if err := f.applySetting(setting); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeSettingRepo) applySetting(setting *model.Setting) error {
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
			KeyName:     model.SettingSiteName,
			Value:       "闲雨小筑",
			IsSensitive: false,
		},
		{
			KeyName:     model.SettingMailPassword,
			Value:       "secret123",
			IsSensitive: true,
		},
	}

	result := maskSensitiveSettings(settings)

	// 非敏感项保留原值。
	if result[0].Value != "闲雨小筑" {
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

// TestUpdateSettingsRollbackLeavesNothingApplied 验证写回失败时整批不生效。
// 校验先行意味着失败发生在写回前，任何设置项都不能被部分更新。
func TestUpdateSettingsRollbackLeavesNothingApplied(t *testing.T) {
	repo := &fakeSettingRepo{
		settings: []*model.Setting{
			{KeyName: model.SettingSiteName, Value: "旧站名"},
		},
		batchErr: errors.New("模拟批量写回失败"),
	}
	svc := NewSettingService(repo)

	items := []UpdateSettingItem{
		{KeyName: model.SettingSiteName, Value: "新站名"},
	}
	if _, err := svc.UpdateSettings(items, 1); err == nil {
		t.Fatal("批量写回失败应返回错误")
	}

	if repo.settings[0].Value != "旧站名" {
		t.Errorf("写回失败后值应保持原状 = %q, 期望 旧站名", repo.settings[0].Value)
	}
}

// TestUpdateSettingsMixedBatchValidatesAllBeforeWrite 验证混合批次中任一条目违规时整批被拒绝。
// 仅第一项合规、第二项只读，先校验后写回路径下两项都不应落库。
func TestUpdateSettingsMixedBatchValidatesAllBeforeWrite(t *testing.T) {
	repo := &fakeSettingRepo{
		settings: []*model.Setting{
			{KeyName: model.SettingSiteName, Value: "旧站名"},
			{KeyName: "readonly_key", Value: "旧值", IsReadonly: true},
		},
	}
	svc := NewSettingService(repo)

	items := []UpdateSettingItem{
		{KeyName: model.SettingSiteName, Value: "新站名"},
		{KeyName: "readonly_key", Value: "越权新值"},
	}
	if _, err := svc.UpdateSettings(items, 1); err == nil {
		t.Fatal("混合批次含只读项应整体被拒绝")
	}

	if repo.settings[0].Value != "旧站名" {
		t.Errorf("整批被拒绝时首项值不得更新 = %q, 期望 旧站名", repo.settings[0].Value)
	}
}
