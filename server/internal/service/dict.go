// Package service 业务逻辑层
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// 字典业务逻辑。
// 字典类型与字典项的管理写入经本层校验，app 端查询只输出已生效集合；
// 字典码与字典项值在创建后不可变更，保护业务侧按码取值的契约稳定。

// dictCodePattern 字典码格式，小写字母开头，仅允许小写字母、数字与下划线。
var dictCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// DictServiceInterface 字典服务接口
type DictServiceInterface interface {
	// 字典类型管理
	CreateDictType(req *CreateDictTypeRequest) (*model.DictType, error)
	UpdateDictType(req *UpdateDictTypeRequest) (*model.DictType, error)
	DeleteDictType(id uint) error
	ListDictTypes(req *ListDictTypesRequest) (*DictTypeListResponse, error)

	// 字典项管理
	CreateDictItem(req *CreateDictItemRequest) (*model.DictItem, error)
	UpdateDictItem(req *UpdateDictItemRequest) (*model.DictItem, error)
	DeleteDictItem(id uint) error
	ListDictItems(req *ListDictItemsRequest) (*DictItemListResponse, error)

	// app 端查询
	ListEnabledDicts() ([]*EnabledDictGroup, error)
	ListEnabledItemsByTypeCode(code string) (*EnabledDictGroup, error)
}

// DictTypeI18nPayload 字典类型单语言翻译字段包，字段可选，提供即更新，缺省时保留既有翻译值。
type DictTypeI18nPayload struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=50"`
	Description *string `json:"description" binding:"omitempty,max=200"`
}

// DictItemI18nPayload 字典项单语言翻译字段包，字段可选，提供即更新，缺省时保留既有翻译值。
type DictItemI18nPayload struct {
	Label       *string `json:"label" binding:"omitempty,min=1,max=50"`
	Description *string `json:"description" binding:"omitempty,max=200"`
}

// CreateDictTypeRequest 创建字典类型请求
type CreateDictTypeRequest struct {
	Code        string                                  `json:"code" binding:"required,min=2,max=50"`
	Name        string                                  `json:"name" binding:"required,min=1,max=50"`
	Description string                                  `json:"description" binding:"omitempty,max=200"`
	Status      *int                                    `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int                                    `json:"sortOrder" binding:"omitempty,min=0,max=9999"`
	Extra       json.RawMessage                         `json:"extra"`
	I18n        map[domain.Language]DictTypeI18nPayload `json:"i18n" binding:"omitempty,dive"`
}

// UpdateDictTypeRequest 更新字典类型请求，字典码创建后不可变更。
type UpdateDictTypeRequest struct {
	ID          uint                                    `json:"id" binding:"required"`
	Name        *string                                 `json:"name" binding:"omitempty,min=1,max=50"`
	Description *string                                 `json:"description" binding:"omitempty,max=200"`
	Status      *int                                    `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int                                    `json:"sortOrder" binding:"omitempty,min=0,max=9999"`
	Extra       json.RawMessage                         `json:"extra"`
	I18n        map[domain.Language]DictTypeI18nPayload `json:"i18n" binding:"omitempty,dive"`
}

// ListDictTypesRequest 字典类型列表请求
type ListDictTypesRequest struct {
	Page     int    `json:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"pageSize" binding:"omitempty,min=1,max=100"`
	Status   *int   `json:"status"`
	Search   string `json:"search"`
}

// DictTypeListResponse 字典类型列表响应
type DictTypeListResponse struct {
	Types    []*model.DictType `json:"types"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// CreateDictItemRequest 创建字典项请求
type CreateDictItemRequest struct {
	TypeID      uint                                    `json:"typeId" binding:"required"`
	Value       string                                  `json:"value" binding:"required,max=50"`
	Label       string                                  `json:"label" binding:"required,min=1,max=50"`
	Description string                                  `json:"description" binding:"omitempty,max=200"`
	Status      *int                                    `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int                                    `json:"sortOrder" binding:"omitempty,min=0,max=9999"`
	Extra       json.RawMessage                         `json:"extra"`
	I18n        map[domain.Language]DictItemI18nPayload `json:"i18n" binding:"omitempty,dive"`
}

// UpdateDictItemRequest 更新字典项请求，字典项值创建后不可变更。
type UpdateDictItemRequest struct {
	ID          uint                                    `json:"id" binding:"required"`
	Label       *string                                 `json:"label" binding:"omitempty,min=1,max=50"`
	Description *string                                 `json:"description" binding:"omitempty,max=200"`
	Status      *int                                    `json:"status" binding:"omitempty,oneof=0 1"`
	SortOrder   *int                                    `json:"sortOrder" binding:"omitempty,min=0,max=9999"`
	Extra       json.RawMessage                         `json:"extra"`
	I18n        map[domain.Language]DictItemI18nPayload `json:"i18n" binding:"omitempty,dive"`
}

// ListDictItemsRequest 字典项列表请求，管理端按类型查询。
type ListDictItemsRequest struct {
	TypeID   *uint  `json:"typeId" binding:"required"`
	Page     int    `json:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"pageSize" binding:"omitempty,min=1,max=100"`
	Status   *int   `json:"status"`
	Search   string `json:"search"`
}

// DictItemListResponse 字典项列表响应
type DictItemListResponse struct {
	Items    []*model.DictItem `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
}

// EnabledDictGroup 已生效字典分组，内嵌字典类型本体并附加已生效字典项集合，
// JSON 输出时类型字段与 items 平铺为同一层级。
type EnabledDictGroup struct {
	model.DictType
	Items []*model.DictItem `json:"items"`
}

// DictService 字典服务实现
type DictService struct {
	dictRepo repository.DictRepositoryInterface
	languagePolicyHolder
}

// NewDictService 创建字典服务实例
func NewDictService(dictRepo repository.DictRepositoryInterface, options ...DictServiceOption) DictServiceInterface {
	service := &DictService{
		dictRepo: dictRepo,
		languagePolicyHolder: languagePolicyHolder{
			languagePolicy: domain.DefaultLanguagePolicy,
		},
	}
	for _, option := range options {
		option(service)
	}
	return service
}

// CreateDictType 创建字典类型
func (s *DictService) CreateDictType(req *CreateDictTypeRequest) (*model.DictType, error) {
	code, err := validateDictCode(req.Code)
	if err != nil {
		return nil, err
	}
	if err := s.requireDictTypeCodeAvailable(code, 0); err != nil {
		return nil, err
	}
	if err := validateDictExtra(req.Extra); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("%w：字典类型名称不能为空", ErrInvalidRequest)
	}

	dictType := buildNewDictType(req, code, name)
	translations, err := s.buildDictTypeTranslationRows(nil, req.I18n)
	if err != nil {
		return nil, err
	}

	if err := s.dictRepo.CreateDictType(dictType); err != nil {
		return nil, err
	}

	// 类型翻译行与主表写入分两步提交，翻译写入失败不影响类型本体的创建结果。
	if err := s.dictRepo.UpsertTypeTranslations(dictType.ID, translations); err != nil {
		return nil, fmt.Errorf("写入字典类型翻译失败: %w", err)
	}
	return s.dictRepo.GetDictTypeByID(dictType.ID)
}

// UpdateDictType 更新字典类型，字典码不参与更新以保护业务取值契约。
func (s *DictService) UpdateDictType(req *UpdateDictTypeRequest) (*model.DictType, error) {
	dictType, err := s.dictRepo.GetDictTypeByID(req.ID)
	if err != nil {
		return nil, err
	}
	if err := validateDictExtra(req.Extra); err != nil {
		return nil, err
	}

	applyDictTypePatch(dictType, req)
	if dictType.Name == "" {
		return nil, fmt.Errorf("%w：字典类型名称不能为空", ErrInvalidRequest)
	}

	// 以既有翻译行为底合并补丁，提供即更新，未提供的字段保留既有翻译值。
	translations, err := s.buildDictTypeTranslationRows(dictType.Translations, req.I18n)
	if err != nil {
		return nil, err
	}

	if err := s.dictRepo.UpdateDictType(dictType); err != nil {
		return nil, err
	}

	// 类型翻译行与主表更新分两步提交，翻译写入失败不影响类型本体的更新结果。
	if err := s.dictRepo.UpsertTypeTranslations(dictType.ID, translations); err != nil {
		return nil, fmt.Errorf("写入字典类型翻译失败: %w", err)
	}
	return s.dictRepo.GetDictTypeByID(dictType.ID)
}

// DeleteDictType 删除字典类型，仓储层在事务内级联清理字典项与翻译行。
func (s *DictService) DeleteDictType(id uint) error {
	if _, err := s.dictRepo.GetDictTypeByID(id); err != nil {
		return err
	}
	return s.dictRepo.DeleteDictType(id)
}

// ListDictTypes 分页查询字典类型列表
func (s *DictService) ListDictTypes(req *ListDictTypesRequest) (*DictTypeListResponse, error) {
	page, pageSize := normalizeDictListPage(req.Page, req.PageSize)

	types, total, err := s.dictRepo.ListDictTypes(&repository.DictTypeListParams{
		Page:     page,
		PageSize: pageSize,
		Status:   req.Status,
		Search:   req.Search,
	})
	if err != nil {
		return nil, err
	}

	return &DictTypeListResponse{Types: types, Total: total, Page: page, PageSize: pageSize}, nil
}

// CreateDictItem 创建字典项
func (s *DictService) CreateDictItem(req *CreateDictItemRequest) (*model.DictItem, error) {
	// 所属字典类型必须存在，停用类型仍允许维护字典项。
	if _, err := s.dictRepo.GetDictTypeByID(req.TypeID); err != nil {
		return nil, err
	}
	value, err := validateDictValue(req.Value)
	if err != nil {
		return nil, err
	}
	if err := s.requireDictItemValueAvailable(req.TypeID, value, 0); err != nil {
		return nil, err
	}
	if err := validateDictExtra(req.Extra); err != nil {
		return nil, err
	}

	label := strings.TrimSpace(req.Label)
	if label == "" {
		return nil, fmt.Errorf("%w：字典项显示名不能为空", ErrInvalidRequest)
	}

	item := buildNewDictItem(req, value, label)
	translations, err := s.buildDictItemTranslationRows(nil, req.I18n)
	if err != nil {
		return nil, err
	}

	if err := s.dictRepo.CreateDictItem(item); err != nil {
		return nil, err
	}

	// 字典项翻译行与主表写入分两步提交，翻译写入失败不影响字典项本体的创建结果。
	if err := s.dictRepo.UpsertItemTranslations(item.ID, translations); err != nil {
		return nil, fmt.Errorf("写入字典项翻译失败: %w", err)
	}
	return s.dictRepo.GetDictItemByID(item.ID)
}

// UpdateDictItem 更新字典项，字典项值不参与更新以保护业务取值契约。
func (s *DictService) UpdateDictItem(req *UpdateDictItemRequest) (*model.DictItem, error) {
	item, err := s.dictRepo.GetDictItemByID(req.ID)
	if err != nil {
		return nil, err
	}
	if err := validateDictExtra(req.Extra); err != nil {
		return nil, err
	}

	applyDictItemPatch(item, req)
	if item.Label == "" {
		return nil, fmt.Errorf("%w：字典项显示名不能为空", ErrInvalidRequest)
	}

	// 以既有翻译行为底合并补丁，提供即更新，未提供的字段保留既有翻译值。
	translations, err := s.buildDictItemTranslationRows(item.Translations, req.I18n)
	if err != nil {
		return nil, err
	}

	if err := s.dictRepo.UpdateDictItem(item); err != nil {
		return nil, err
	}

	// 字典项翻译行与主表更新分两步提交，翻译写入失败不影响字典项本体的更新结果。
	if err := s.dictRepo.UpsertItemTranslations(item.ID, translations); err != nil {
		return nil, fmt.Errorf("写入字典项翻译失败: %w", err)
	}
	return s.dictRepo.GetDictItemByID(item.ID)
}

// DeleteDictItem 删除字典项，仓储层在事务内级联清理翻译行。
func (s *DictService) DeleteDictItem(id uint) error {
	if _, err := s.dictRepo.GetDictItemByID(id); err != nil {
		return err
	}
	return s.dictRepo.DeleteDictItem(id)
}

// ListDictItems 分页查询字典项列表，管理端按类型查询。
func (s *DictService) ListDictItems(req *ListDictItemsRequest) (*DictItemListResponse, error) {
	page, pageSize := normalizeDictListPage(req.Page, req.PageSize)

	items, total, err := s.dictRepo.ListDictItems(&repository.DictItemListParams{
		TypeID:   req.TypeID,
		Page:     page,
		PageSize: pageSize,
		Status:   req.Status,
		Search:   req.Search,
	})
	if err != nil {
		return nil, err
	}

	return &DictItemListResponse{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// ListEnabledDicts 输出全部已生效字典及其已生效字典项，供 app 端全量字典接口使用。
func (s *DictService) ListEnabledDicts() ([]*EnabledDictGroup, error) {
	types, err := s.dictRepo.ListEnabledDictTypes()
	if err != nil {
		return nil, err
	}

	typeIDs := make([]uint, 0, len(types))
	for _, dictType := range types {
		typeIDs = append(typeIDs, dictType.ID)
	}

	items, err := s.dictRepo.ListEnabledItemsByTypes(typeIDs)
	if err != nil {
		return nil, err
	}

	return buildDictGroups(types, groupDictItemsByType(items)), nil
}

// ListEnabledItemsByTypeCode 按字典码输出单个已生效字典及其已生效字典项。
// 类型不存在或已停用时统一返回类型不存在错误，由 handler 层映射 404。
func (s *DictService) ListEnabledItemsByTypeCode(code string) (*EnabledDictGroup, error) {
	dictType, err := s.dictRepo.GetDictTypeByCode(strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	if !dictType.IsEnabled() {
		return nil, repository.ErrDictTypeNotFound
	}

	items, err := s.dictRepo.ListEnabledItemsByType(dictType.ID)
	if err != nil {
		return nil, err
	}

	return &EnabledDictGroup{DictType: *dictType, Items: items}, nil
}

// validateDictCode 校验并归一化字典码，返回去除首尾空格后的合法字典码。
func validateDictCode(raw string) (string, error) {
	code := strings.TrimSpace(raw)
	if !dictCodePattern.MatchString(code) {
		return "", fmt.Errorf("%w：字典码需以小写字母开头，仅含小写字母、数字与下划线", ErrInvalidRequest)
	}
	return code, nil
}

// validateDictValue 校验并归一化字典项值。
func validateDictValue(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("%w：字典项值不能为空", ErrInvalidRequest)
	}
	return value, nil
}

// validateDictExtra 校验扩展字段为合法 JSON，未提供时直接通过。
func validateDictExtra(extra json.RawMessage) error {
	if len(extra) == 0 {
		return nil
	}
	if !json.Valid(extra) {
		return fmt.Errorf("%w：扩展字段必须是合法 JSON", ErrInvalidRequest)
	}
	return nil
}

// requireDictTypeCodeAvailable 校验字典码未被占用，更新场景以 excludeID 排除自身，新建场景传 0。
func (s *DictService) requireDictTypeCodeAvailable(code string, excludeID uint) error {
	existing, err := s.dictRepo.GetDictTypeByCode(code)
	if err != nil {
		if errors.Is(err, repository.ErrDictTypeNotFound) {
			return nil
		}
		return err
	}
	if existing.ID != excludeID {
		return fmt.Errorf("%w：字典码已存在", ErrInvalidRequest)
	}
	return nil
}

// requireDictItemValueAvailable 校验字典项值在类型内未被占用，更新场景以 excludeID 排除自身，新建场景传 0。
func (s *DictService) requireDictItemValueAvailable(typeID uint, value string, excludeID uint) error {
	existing, err := s.dictRepo.GetDictItemByValue(typeID, value)
	if err != nil {
		if errors.Is(err, repository.ErrDictItemNotFound) {
			return nil
		}
		return err
	}
	if existing.ID != excludeID {
		return fmt.Errorf("%w：字典项值已存在", ErrInvalidRequest)
	}
	return nil
}

// buildNewDictType 按请求构造字典类型实体，未指定状态时默认生效。
func buildNewDictType(req *CreateDictTypeRequest, code string, name string) *model.DictType {
	dictType := &model.DictType{
		Code:        code,
		Name:        name,
		Description: req.Description,
		Status:      model.DictStatusEnabled,
		Extra:       req.Extra,
	}
	if req.Status != nil {
		dictType.Status = *req.Status
	}
	if req.SortOrder != nil {
		dictType.SortOrder = *req.SortOrder
	}
	return dictType
}

// applyDictTypePatch 将可选字段合并到既有字典类型，未提供的字段保留原值。
func applyDictTypePatch(dictType *model.DictType, req *UpdateDictTypeRequest) {
	if req.Name != nil {
		dictType.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		dictType.Description = *req.Description
	}
	if req.Status != nil {
		dictType.Status = *req.Status
	}
	if req.SortOrder != nil {
		dictType.SortOrder = *req.SortOrder
	}
	if req.Extra != nil {
		dictType.Extra = req.Extra
	}
}

// buildNewDictItem 按请求构造字典项实体，未指定状态时默认生效。
func buildNewDictItem(req *CreateDictItemRequest, value string, label string) *model.DictItem {
	item := &model.DictItem{
		TypeID:      req.TypeID,
		Value:       value,
		Label:       label,
		Description: req.Description,
		Status:      model.DictStatusEnabled,
		Extra:       req.Extra,
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	return item
}

// applyDictItemPatch 将可选字段合并到既有字典项，未提供的字段保留原值。
func applyDictItemPatch(item *model.DictItem, req *UpdateDictItemRequest) {
	if req.Label != nil {
		item.Label = strings.TrimSpace(*req.Label)
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	if req.Extra != nil {
		item.Extra = req.Extra
	}
}

// buildDictTypeTranslationRows 校验语言键并构造待写入的字典类型翻译行集合。
func (s *DictService) buildDictTypeTranslationRows(existing []model.DictTypeTranslation, patches map[domain.Language]DictTypeI18nPayload) ([]model.DictTypeTranslation, error) {
	if len(patches) == 0 {
		return nil, nil
	}

	rows := make([]model.DictTypeTranslation, 0, len(patches))
	for language, patch := range patches {
		if err := s.requireWritableTranslation(language); err != nil {
			return nil, err
		}

		// 以既有翻译行为底，无既有行时从零值开始合并补丁。
		row := model.DictTypeTranslation{Locale: string(language)}
		if existingRow := findDictTypeTranslation(existing, language); existingRow != nil {
			row = *existingRow
		}
		if patch.Name != nil {
			row.Name = *patch.Name
		}
		if patch.Description != nil {
			row.Description = *patch.Description
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// buildDictItemTranslationRows 校验语言键并构造待写入的字典项翻译行集合。
func (s *DictService) buildDictItemTranslationRows(existing []model.DictItemTranslation, patches map[domain.Language]DictItemI18nPayload) ([]model.DictItemTranslation, error) {
	if len(patches) == 0 {
		return nil, nil
	}

	rows := make([]model.DictItemTranslation, 0, len(patches))
	for language, patch := range patches {
		if err := s.requireWritableTranslation(language); err != nil {
			return nil, err
		}

		// 以既有翻译行为底，无既有行时从零值开始合并补丁。
		row := model.DictItemTranslation{Locale: string(language)}
		if existingRow := findDictItemTranslation(existing, language); existingRow != nil {
			row = *existingRow
		}
		if patch.Label != nil {
			row.Label = *patch.Label
		}
		if patch.Description != nil {
			row.Description = *patch.Description
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// groupDictItemsByType 将字典项集合按所属类型ID分组。
func groupDictItemsByType(items []*model.DictItem) map[uint][]*model.DictItem {
	itemsByType := make(map[uint][]*model.DictItem, len(items))
	for _, item := range items {
		itemsByType[item.TypeID] = append(itemsByType[item.TypeID], item)
	}
	return itemsByType
}

// buildDictGroups 按类型顺序组装已生效字典分组，无字典项的类型输出空集合保证前端迭代安全。
func buildDictGroups(types []*model.DictType, itemsByType map[uint][]*model.DictItem) []*EnabledDictGroup {
	groups := make([]*EnabledDictGroup, 0, len(types))
	for _, dictType := range types {
		items := itemsByType[dictType.ID]
		if items == nil {
			items = []*model.DictItem{}
		}
		groups = append(groups, &EnabledDictGroup{DictType: *dictType, Items: items})
	}
	return groups
}

// normalizeDictListPage 归一化字典列表分页参数，非法值回退默认页大小。
func normalizeDictListPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
