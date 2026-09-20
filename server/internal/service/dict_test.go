package service

import (
	"encoding/json"
	"errors"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeDictRepo 字典仓储的测试替身，记录调用并返回可配置结果。
type fakeDictRepo struct {
	repository.DictRepositoryInterface
	types     []*model.DictType
	items     []*model.DictItem
	created   bool
	updated   bool
	deleted   bool
	typeIDSeq uint
}

// CreateDictType 显式覆写类型创建，分配自增ID并记录调用。
func (f *fakeDictRepo) CreateDictType(dictType *model.DictType) error {
	f.typeIDSeq++
	dictType.ID = f.typeIDSeq
	f.types = append(f.types, dictType)
	f.created = true
	return nil
}

// UpdateDictType 显式覆写类型更新，记录调用。
func (f *fakeDictRepo) UpdateDictType(dictType *model.DictType) error {
	f.updated = true
	return nil
}

// DeleteDictType 显式覆写类型删除，记录调用。
func (f *fakeDictRepo) DeleteDictType(id uint) error {
	f.deleted = true
	return nil
}

// GetDictTypeByID 按ID返回字典类型，未命中返回哨兵错误。
func (f *fakeDictRepo) GetDictTypeByID(id uint) (*model.DictType, error) {
	for _, dictType := range f.types {
		if dictType.ID == id {
			return dictType, nil
		}
	}
	return nil, repository.ErrDictTypeNotFound
}

// GetDictTypeByCode 按字典码返回字典类型，未命中返回哨兵错误。
func (f *fakeDictRepo) GetDictTypeByCode(code string) (*model.DictType, error) {
	for _, dictType := range f.types {
		if dictType.Code == code {
			return dictType, nil
		}
	}
	return nil, repository.ErrDictTypeNotFound
}

// UpsertTypeTranslations 显式覆写类型翻译写入，记录调用。
func (f *fakeDictRepo) UpsertTypeTranslations(typeID uint, translations []model.DictTypeTranslation) error {
	for _, dictType := range f.types {
		if dictType.ID == typeID {
			dictType.Translations = translations
		}
	}
	return nil
}

// CreateDictItem 显式覆写字典项创建，分配自增ID并记录调用。
func (f *fakeDictRepo) CreateDictItem(item *model.DictItem) error {
	f.typeIDSeq++
	item.ID = f.typeIDSeq
	f.items = append(f.items, item)
	f.created = true
	return nil
}

// UpdateDictItem 显式覆写字典项更新，记录调用。
func (f *fakeDictRepo) UpdateDictItem(item *model.DictItem) error {
	f.updated = true
	return nil
}

// DeleteDictItem 显式覆写字典项删除，记录调用。
func (f *fakeDictRepo) DeleteDictItem(id uint) error {
	f.deleted = true
	return nil
}

// GetDictItemByID 按ID返回字典项，未命中返回哨兵错误。
func (f *fakeDictRepo) GetDictItemByID(id uint) (*model.DictItem, error) {
	for _, item := range f.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, repository.ErrDictItemNotFound
}

// GetDictItemByValue 按类型与值返回字典项，未命中返回哨兵错误。
func (f *fakeDictRepo) GetDictItemByValue(typeID uint, value string) (*model.DictItem, error) {
	for _, item := range f.items {
		if item.TypeID == typeID && item.Value == value {
			return item, nil
		}
	}
	return nil, repository.ErrDictItemNotFound
}

// UpsertItemTranslations 显式覆写字典项翻译写入，记录调用。
func (f *fakeDictRepo) UpsertItemTranslations(itemID uint, translations []model.DictItemTranslation) error {
	for _, item := range f.items {
		if item.ID == itemID {
			item.Translations = translations
		}
	}
	return nil
}

// ListEnabledDictTypes 返回全部已生效字典类型。
func (f *fakeDictRepo) ListEnabledDictTypes() ([]*model.DictType, error) {
	enabled := make([]*model.DictType, 0, len(f.types))
	for _, dictType := range f.types {
		if dictType.IsEnabled() {
			enabled = append(enabled, dictType)
		}
	}
	return enabled, nil
}

// ListEnabledItemsByTypes 返回给定类型集合下的已生效字典项。
func (f *fakeDictRepo) ListEnabledItemsByTypes(typeIDs []uint) ([]*model.DictItem, error) {
	result := make([]*model.DictItem, 0)
	for _, item := range f.items {
		if !item.IsEnabled() {
			continue
		}
		for _, typeID := range typeIDs {
			if item.TypeID == typeID {
				result = append(result, item)
			}
		}
	}
	return result, nil
}

// ListEnabledItemsByType 返回单个类型下的已生效字典项。
func (f *fakeDictRepo) ListEnabledItemsByType(typeID uint) ([]*model.DictItem, error) {
	result := make([]*model.DictItem, 0)
	for _, item := range f.items {
		if item.TypeID == typeID && item.IsEnabled() {
			result = append(result, item)
		}
	}
	return result, nil
}

// TestCreateDictTypeValidatesCodeFormat 验证字典码格式非法时返回请求错误。
func TestCreateDictTypeValidatesCodeFormat(t *testing.T) {
	repo := &fakeDictRepo{}
	svc := NewDictService(repo)

	for _, code := range []string{"TagStatus", "1status", "tag-status", "tag status"} {
		req := &CreateDictTypeRequest{Code: code, Name: "标签状态"}
		if _, err := svc.CreateDictType(req); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("字典码 %q 应返回 ErrInvalidRequest, 实际: %v", code, err)
		}
	}
}

// TestCreateDictTypeUniqueCode 验证字典码重复时返回业务错误。
func TestCreateDictTypeUniqueCode(t *testing.T) {
	repo := &fakeDictRepo{
		types: []*model.DictType{{ID: 1, Code: "tag_status", Name: "标签状态"}},
	}
	svc := NewDictService(repo)

	req := &CreateDictTypeRequest{Code: "tag_status", Name: "重复字典"}
	if _, err := svc.CreateDictType(req); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("重复字典码应返回 ErrInvalidRequest, 实际: %v", err)
	}
}

// TestCreateDictTypeDefaults 验证创建字典类型时默认状态生效。
func TestCreateDictTypeDefaults(t *testing.T) {
	repo := &fakeDictRepo{}
	svc := NewDictService(repo)

	req := &CreateDictTypeRequest{Code: "tag_status", Name: "标签状态"}
	dictType, err := svc.CreateDictType(req)
	if err != nil {
		t.Fatalf("创建字典类型失败: %v", err)
	}
	if dictType.Status != model.DictStatusEnabled {
		t.Errorf("默认状态 = %d, 期望生效 1", dictType.Status)
	}
	if !repo.created {
		t.Error("创建字典类型应触发仓储写入")
	}
}

// TestCreateDictTypeRejectsInvalidI18nLocale 验证缺省语言翻译键被拒绝。
func TestCreateDictTypeRejectsInvalidI18nLocale(t *testing.T) {
	repo := &fakeDictRepo{}
	svc := NewDictService(repo)

	name := "Tag Status"
	req := &CreateDictTypeRequest{
		Code: "tag_status",
		Name: "标签状态",
		I18n: map[domain.Language]DictTypeI18nPayload{
			domain.DefaultLanguage: {Name: &name},
		},
	}
	if _, err := svc.CreateDictType(req); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("缺省语言翻译键应返回 ErrInvalidRequest, 实际: %v", err)
	}
}

// TestCreateDictItemRequiresExistingType 验证字典类型不存在时创建字典项被拒绝。
func TestCreateDictItemRequiresExistingType(t *testing.T) {
	repo := &fakeDictRepo{}
	svc := NewDictService(repo)

	req := &CreateDictItemRequest{TypeID: 99, Value: "1", Label: "启用"}
	if _, err := svc.CreateDictItem(req); !errors.Is(err, repository.ErrDictTypeNotFound) {
		t.Errorf("字典类型不存在应返回哨兵错误, 实际: %v", err)
	}
}

// TestCreateDictItemUniqueValueInType 验证同类型内字典项值唯一，跨类型可重复。
func TestCreateDictItemUniqueValueInType(t *testing.T) {
	repo := &fakeDictRepo{
		types: []*model.DictType{{ID: 1, Code: "tag_status", Name: "标签状态"}},
		items: []*model.DictItem{{ID: 1, TypeID: 1, Value: "1", Label: "启用"}},
	}
	svc := NewDictService(repo)

	duplicate := &CreateDictItemRequest{TypeID: 1, Value: "1", Label: "再次启用"}
	if _, err := svc.CreateDictItem(duplicate); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("同类型内重复值应返回 ErrInvalidRequest, 实际: %v", err)
	}

	otherType := &model.DictType{ID: 2, Code: "comment_status", Name: "评论状态"}
	repo.types = append(repo.types, otherType)
	crossType := &CreateDictItemRequest{TypeID: 2, Value: "1", Label: "显示"}
	if _, err := svc.CreateDictItem(crossType); err != nil {
		t.Errorf("跨类型同值不应返回错误: %v", err)
	}
}

// TestCreateDictItemValidatesExtraJSON 验证扩展字段非合法 JSON 时被拒绝。
func TestCreateDictItemValidatesExtraJSON(t *testing.T) {
	repo := &fakeDictRepo{
		types: []*model.DictType{{ID: 1, Code: "tag_status", Name: "标签状态"}},
	}
	svc := NewDictService(repo)

	req := &CreateDictItemRequest{TypeID: 1, Value: "2", Label: "待审核", Extra: json.RawMessage("{invalid}")}
	if _, err := svc.CreateDictItem(req); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("非法 JSON 扩展字段应返回 ErrInvalidRequest, 实际: %v", err)
	}
}

// TestUpdateDictTypeKeepsOmittedFields 验证更新字典类型时省略字段保留原值。
func TestUpdateDictTypeKeepsOmittedFields(t *testing.T) {
	repo := &fakeDictRepo{
		types: []*model.DictType{{
			ID:          1,
			Code:        "tag_status",
			Name:        "标签状态",
			Description: "原始描述",
			SortOrder:   5,
		}},
	}
	svc := NewDictService(repo)

	name := "标签状态更新"
	req := &UpdateDictTypeRequest{ID: 1, Name: &name}
	dictType, err := svc.UpdateDictType(req)
	if err != nil {
		t.Fatalf("更新字典类型失败: %v", err)
	}
	if dictType.Description != "原始描述" {
		t.Errorf("描述 = %q, 期望保留原值", dictType.Description)
	}
	if dictType.SortOrder != 5 {
		t.Errorf("排序 = %d, 期望保留原值", dictType.SortOrder)
	}
}

// TestListEnabledDictsGroupsItems 验证全量字典按类型分组且只含已生效集合。
func TestListEnabledDictsGroupsItems(t *testing.T) {
	repo := &fakeDictRepo{
		types: []*model.DictType{
			{ID: 1, Code: "tag_status", Name: "标签状态", Status: model.DictStatusEnabled},
			{ID: 2, Code: "disabled_type", Name: "停用字典", Status: model.DictStatusDisabled},
		},
		items: []*model.DictItem{
			{ID: 1, TypeID: 1, Value: "1", Label: "启用", Status: model.DictStatusEnabled},
			{ID: 2, TypeID: 1, Value: "0", Label: "隐藏", Status: model.DictStatusDisabled},
		},
	}
	svc := NewDictService(repo)

	groups, err := svc.ListEnabledDicts()
	if err != nil {
		t.Fatalf("查询全量字典失败: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("期望仅返回 1 个已生效分组, 实际 %d", len(groups))
	}
	if groups[0].Code != "tag_status" {
		t.Errorf("分组字典码 = %q, 期望 tag_status", groups[0].Code)
	}
	// 已停用的字典项不得进入全量输出。
	if len(groups[0].Items) != 1 || groups[0].Items[0].Value != "1" {
		t.Errorf("期望仅含已生效字典项, 实际 %v", groups[0].Items)
	}
}

// TestListEnabledItemsByTypeCodeHidesDisabledType 验证停用类型对外等价于不存在。
func TestListEnabledItemsByTypeCodeHidesDisabledType(t *testing.T) {
	repo := &fakeDictRepo{
		types: []*model.DictType{
			{ID: 2, Code: "disabled_type", Name: "停用字典", Status: model.DictStatusDisabled},
		},
	}
	svc := NewDictService(repo)

	if _, err := svc.ListEnabledItemsByTypeCode("disabled_type"); !errors.Is(err, repository.ErrDictTypeNotFound) {
		t.Errorf("停用类型应返回类型不存在, 实际: %v", err)
	}
}

// TestLocalizeDictGroupFallsBackToDefault 验证分组本地化按字段回退缺省语言。
func TestLocalizeDictGroupFallsBackToDefault(t *testing.T) {
	group := &EnabledDictGroup{
		DictType: model.DictType{
			Name: "标签状态",
			Translations: []model.DictTypeTranslation{
				{Locale: "en", Name: "Tag Status"},
			},
		},
		Items: []*model.DictItem{
			{
				Label: "启用",
				Translations: []model.DictItemTranslation{
					{Locale: "en", Label: "Enabled"},
				},
			},
			{Label: "隐藏"},
		},
	}

	// 类型与字典项按字段独立回退，缺失翻译的项保留主列内容。
	if language := LocalizeDictType(&group.DictType, "en"); language != "en" {
		t.Errorf("类型命中翻译时实际语言 = %q, 期望 en", language)
	}
	for _, item := range group.Items {
		LocalizeDictItem(item, "en")
	}
	if group.Name != "Tag Status" {
		t.Errorf("分组名称 = %q, 期望 Tag Status", group.Name)
	}
	if group.Items[0].Label != "Enabled" {
		t.Errorf("字典项标签 = %q, 期望 Enabled", group.Items[0].Label)
	}
	if group.Items[1].Label != "隐藏" {
		t.Errorf("缺失翻译的字典项标签 = %q, 期望回退 隐藏", group.Items[1].Label)
	}
	if group.Items[0].Translations != nil {
		t.Error("本地化输出后应清空翻译行集合")
	}
}
