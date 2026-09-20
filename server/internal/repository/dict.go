// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"

	"gorm.io/gorm"
)

// ErrDictTypeNotFound 字典类型不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrDictTypeNotFound = errors.New("字典类型不存在")

// ErrDictItemNotFound 字典项不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrDictItemNotFound = errors.New("字典项不存在")

// DictRepositoryInterface 字典仓储接口
type DictRepositoryInterface interface {
	// 字典类型
	CreateDictType(dictType *model.DictType) error
	UpdateDictType(dictType *model.DictType) error
	DeleteDictType(id uint) error
	GetDictTypeByID(id uint) (*model.DictType, error)
	GetDictTypeByCode(code string) (*model.DictType, error)
	ListDictTypes(params *DictTypeListParams) ([]*model.DictType, int64, error)
	ListEnabledDictTypes() ([]*model.DictType, error)

	// 字典项
	CreateDictItem(item *model.DictItem) error
	UpdateDictItem(item *model.DictItem) error
	DeleteDictItem(id uint) error
	GetDictItemByID(id uint) (*model.DictItem, error)
	ListDictItems(params *DictItemListParams) ([]*model.DictItem, int64, error)
	ListEnabledItemsByType(typeID uint) ([]*model.DictItem, error)
	ListEnabledItemsByTypes(typeIDs []uint) ([]*model.DictItem, error)

	// 多语言翻译
	UpsertTypeTranslations(typeID uint, translations []model.DictTypeTranslation) error
	UpsertItemTranslations(itemID uint, translations []model.DictItemTranslation) error
}

// DictTypeListParams 字典类型列表查询参数
type DictTypeListParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Status   *int   `json:"status"`
	Search   string `json:"search"`
}

// DictItemListParams 字典项列表查询参数，typeID 为空时跨类型查询。
type DictItemListParams struct {
	TypeID   *uint  `json:"typeId"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Status   *int   `json:"status"`
	Search   string `json:"search"`
}

// DictRepository 字典仓储实现
type DictRepository struct {
	db *gorm.DB
}

// NewDictRepository 创建字典仓储实例
func NewDictRepository(db *gorm.DB) DictRepositoryInterface {
	return &DictRepository{db: db}
}

// CreateDictType 创建字典类型
func (r *DictRepository) CreateDictType(dictType *model.DictType) error {
	if err := r.db.Create(dictType).Error; err != nil {
		return fmt.Errorf("创建字典类型失败: %w", err)
	}
	return nil
}

// UpdateDictType 更新字典类型
func (r *DictRepository) UpdateDictType(dictType *model.DictType) error {
	if err := r.db.Save(dictType).Error; err != nil {
		return fmt.Errorf("更新字典类型失败: %w", err)
	}
	return nil
}

// DeleteDictType 删除字典类型，事务内按序清理翻译行、字典项与类型本体，
// 不依赖数据库外键级联，约束缺失的环境下仍能保持数据一致。
func (r *DictRepository) DeleteDictType(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var itemIDs []uint
		if err := tx.Model(&model.DictItem{}).Where("type_id = ?", id).Pluck("id", &itemIDs).Error; err != nil {
			return fmt.Errorf("查询字典项失败: %w", err)
		}

		if len(itemIDs) > 0 {
			if err := tx.Where("item_id IN ?", itemIDs).Delete(&model.DictItemTranslation{}).Error; err != nil {
				return fmt.Errorf("删除字典项翻译失败: %w", err)
			}
			if err := tx.Where("type_id = ?", id).Delete(&model.DictItem{}).Error; err != nil {
				return fmt.Errorf("删除字典项失败: %w", err)
			}
		}

		if err := tx.Where("type_id = ?", id).Delete(&model.DictTypeTranslation{}).Error; err != nil {
			return fmt.Errorf("删除字典类型翻译失败: %w", err)
		}
		if err := tx.Delete(&model.DictType{}, id).Error; err != nil {
			return fmt.Errorf("删除字典类型失败: %w", err)
		}
		return nil
	})
}

// GetDictTypeByID 根据ID获取字典类型
func (r *DictRepository) GetDictTypeByID(id uint) (*model.DictType, error) {
	var dictType model.DictType
	if err := r.db.Preload("Translations").First(&dictType, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDictTypeNotFound
		}
		return nil, fmt.Errorf("查询字典类型失败: %w", err)
	}
	return &dictType, nil
}

// GetDictTypeByCode 根据字典码获取字典类型
func (r *DictRepository) GetDictTypeByCode(code string) (*model.DictType, error) {
	var dictType model.DictType
	if err := r.db.Preload("Translations").Where("code = ?", code).First(&dictType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDictTypeNotFound
		}
		return nil, fmt.Errorf("查询字典类型失败: %w", err)
	}
	return &dictType, nil
}

// ListDictTypes 分页查询字典类型列表
func (r *DictRepository) ListDictTypes(params *DictTypeListParams) ([]*model.DictType, int64, error) {
	query := r.db.Model(&model.DictType{}).Preload("Translations")

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("code LIKE ? OR name LIKE ? OR description LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询字典类型总数失败: %w", err)
	}

	normalized := normalizeDictPageParams(params.Page, params.PageSize)
	var types []*model.DictType
	if err := query.Order("sort_order ASC, id ASC").Offset(normalized.offset).Limit(normalized.limit).Find(&types).Error; err != nil {
		return nil, 0, fmt.Errorf("查询字典类型列表失败: %w", err)
	}

	return types, total, nil
}

// ListEnabledDictTypes 查询全部已生效字典类型，供 app 端全量字典接口使用。
func (r *DictRepository) ListEnabledDictTypes() ([]*model.DictType, error) {
	var types []*model.DictType
	if err := r.db.Preload("Translations").
		Where("status = ?", model.DictStatusEnabled).
		Order("sort_order ASC, id ASC").
		Find(&types).Error; err != nil {
		return nil, fmt.Errorf("查询已生效字典类型失败: %w", err)
	}
	return types, nil
}

// CreateDictItem 创建字典项
func (r *DictRepository) CreateDictItem(item *model.DictItem) error {
	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("创建字典项失败: %w", err)
	}
	return nil
}

// UpdateDictItem 更新字典项
func (r *DictRepository) UpdateDictItem(item *model.DictItem) error {
	if err := r.db.Save(item).Error; err != nil {
		return fmt.Errorf("更新字典项失败: %w", err)
	}
	return nil
}

// DeleteDictItem 删除字典项，事务内先清翻译行再删本体。
func (r *DictRepository) DeleteDictItem(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("item_id = ?", id).Delete(&model.DictItemTranslation{}).Error; err != nil {
			return fmt.Errorf("删除字典项翻译失败: %w", err)
		}
		if err := tx.Delete(&model.DictItem{}, id).Error; err != nil {
			return fmt.Errorf("删除字典项失败: %w", err)
		}
		return nil
	})
}

// GetDictItemByID 根据ID获取字典项
func (r *DictRepository) GetDictItemByID(id uint) (*model.DictItem, error) {
	var item model.DictItem
	if err := r.db.Preload("Translations").First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDictItemNotFound
		}
		return nil, fmt.Errorf("查询字典项失败: %w", err)
	}
	return &item, nil
}

// ListDictItems 分页查询字典项列表，typeID 为空时跨类型查询。
func (r *DictRepository) ListDictItems(params *DictItemListParams) ([]*model.DictItem, int64, error) {
	query := r.db.Model(&model.DictItem{}).Preload("Translations")

	if params.TypeID != nil {
		query = query.Where("type_id = ?", *params.TypeID)
	}
	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("value LIKE ? OR label LIKE ? OR description LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询字典项总数失败: %w", err)
	}

	normalized := normalizeDictPageParams(params.Page, params.PageSize)
	var items []*model.DictItem
	if err := query.Order("sort_order ASC, id ASC").Offset(normalized.offset).Limit(normalized.limit).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("查询字典项列表失败: %w", err)
	}

	return items, total, nil
}

// ListEnabledItemsByType 查询单个字典类型下的已生效字典项，供 app 端单字典接口使用。
func (r *DictRepository) ListEnabledItemsByType(typeID uint) ([]*model.DictItem, error) {
	var items []*model.DictItem
	if err := r.db.Preload("Translations").
		Where("type_id = ? AND status = ?", typeID, model.DictStatusEnabled).
		Order("sort_order ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("查询字典项失败: %w", err)
	}
	return items, nil
}

// ListEnabledItemsByTypes 批量查询多个字典类型下的已生效字典项，供 app 端全量字典接口使用。
func (r *DictRepository) ListEnabledItemsByTypes(typeIDs []uint) ([]*model.DictItem, error) {
	if len(typeIDs) == 0 {
		return nil, nil
	}

	var items []*model.DictItem
	if err := r.db.Preload("Translations").
		Where("type_id IN ? AND status = ?", typeIDs, model.DictStatusEnabled).
		Order("type_id ASC, sort_order ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("批量查询字典项失败: %w", err)
	}
	return items, nil
}

// UpsertTypeTranslations 批量写入字典类型翻译行，命中唯一索引时整行更新。
func (r *DictRepository) UpsertTypeTranslations(typeID uint, translations []model.DictTypeTranslation) error {
	if len(translations) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for index := range translations {
			translations[index].TypeID = typeID
		}
		return upsertTranslationRows(tx, &translations)
	})
}

// UpsertItemTranslations 批量写入字典项翻译行，命中唯一索引时整行更新。
func (r *DictRepository) UpsertItemTranslations(itemID uint, translations []model.DictItemTranslation) error {
	if len(translations) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for index := range translations {
			translations[index].ItemID = itemID
		}
		return upsertTranslationRows(tx, &translations)
	})
}

// dictPageParams 归一化后的分页参数。
type dictPageParams struct {
	offset int
	limit  int
}

// normalizeDictPageParams 归一化分页参数，非法值回退默认页大小。
func normalizeDictPageParams(page, pageSize int) dictPageParams {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return dictPageParams{offset: (page - 1) * pageSize, limit: pageSize}
}
