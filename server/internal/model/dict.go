package model

import (
	"encoding/json"
	"time"
)

// 字典模块模型集合。
// 字典为读多写少的配置数据，类型与字典项均采用硬删除策略，
// 删除类型时经外键级联清理字典项与翻译行，避免软删除与唯一索引的占位冲突。

// DictStatusEnabled 字典生效状态常量，类型与字典项共用同一枚举。
const (
	DictStatusDisabled = 0 // 停用
	DictStatusEnabled  = 1 // 生效
)

// DictType 字典类型模型，按业务语义聚合一组可动态管理的状态类枚举项。
type DictType struct {
	ID          uint            `json:"id" gorm:"primaryKey;comment:字典类型ID"`
	Code        string          `json:"code" gorm:"not null;uniqueIndex;size:50;comment:字典码，业务侧唯一定位标识，如 tag_status"`
	Name        string          `json:"name" gorm:"not null;size:50;comment:字典类型名称，缺省语言"`
	Description string          `json:"description" gorm:"size:200;comment:字典类型描述，缺省语言"`
	Status      int             `json:"status" gorm:"type:tinyint;default:1;index;comment:生效状态：1-生效 0-停用"`
	SortOrder   int             `json:"sortOrder" gorm:"default:0;index;comment:排序权重，数值小的靠前"`
	Extra       json.RawMessage `json:"extra" gorm:"type:json;comment:预留扩展字段，存储颜色图标等展示元数据"`
	CreatedAt   time.Time       `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt   time.Time       `json:"updatedAt" gorm:"type:datetime(3);comment:最后更新时间"`

	// 关联关系
	Items        []DictItem            `json:"-" gorm:"foreignKey:TypeID;constraint:OnDelete:CASCADE"`
	Translations []DictTypeTranslation `json:"translations,omitempty" gorm:"foreignKey:TypeID"`
}

// TableName 指定表名
func (DictType) TableName() string {
	return "dict_types"
}

// IsEnabled 判断字典类型是否生效。
func (t *DictType) IsEnabled() bool {
	return t.Status == DictStatusEnabled
}

// DictItem 字典项模型，同一字典类型内 value 唯一，是业务取值的明细行。
type DictItem struct {
	ID          uint            `json:"id" gorm:"primaryKey;comment:字典项ID"`
	TypeID      uint            `json:"typeId" gorm:"not null;uniqueIndex:uk_dict_item,priority:1;index;comment:所属字典类型ID"`
	Value       string          `json:"value" gorm:"not null;uniqueIndex:uk_dict_item,priority:2;size:50;comment:字典项值，同类型内唯一"`
	Label       string          `json:"label" gorm:"not null;size:50;comment:字典项显示名，缺省语言"`
	Description string          `json:"description" gorm:"size:200;comment:字典项描述，缺省语言"`
	Status      int             `json:"status" gorm:"type:tinyint;default:1;index;comment:生效状态：1-生效 0-停用"`
	SortOrder   int             `json:"sortOrder" gorm:"default:0;index;comment:排序权重，数值小的靠前"`
	Extra       json.RawMessage `json:"extra" gorm:"type:json;comment:预留扩展字段，存储颜色图标等展示元数据"`
	CreatedAt   time.Time       `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt   time.Time       `json:"updatedAt" gorm:"type:datetime(3);comment:最后更新时间"`

	// 关联关系
	Type         DictType              `json:"-" gorm:"foreignKey:TypeID;constraint:OnDelete:CASCADE"`
	Translations []DictItemTranslation `json:"translations,omitempty" gorm:"foreignKey:ItemID"`
}

// TableName 指定表名
func (DictItem) TableName() string {
	return "dict_items"
}

// IsEnabled 判断字典项是否生效。
func (i *DictItem) IsEnabled() bool {
	return i.Status == DictStatusEnabled
}

// DictTypeTranslation 字典类型翻译模型，复合唯一索引确保同一类型同一语言仅一行翻译。
type DictTypeTranslation struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:翻译ID"`
	TypeID      uint      `json:"typeId" gorm:"not null;uniqueIndex:uk_dict_type_translation,priority:1;comment:字典类型ID"`
	Locale      string    `json:"locale" gorm:"not null;size:10;uniqueIndex:uk_dict_type_translation,priority:2;index;comment:语言标识，主子标签形式，如 zh、en"`
	Name        string    `json:"name" gorm:"size:50;comment:该语言字典类型名称"`
	Description string    `json:"description" gorm:"size:200;comment:该语言字典类型描述"`
	CreatedAt   time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:最后更新时间"`

	// 关联关系
	Type DictType `json:"-" gorm:"foreignKey:TypeID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (DictTypeTranslation) TableName() string {
	return "dict_type_translations"
}

// DictItemTranslation 字典项翻译模型，复合唯一索引确保同一字典项同一语言仅一行翻译。
type DictItemTranslation struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:翻译ID"`
	ItemID      uint      `json:"itemId" gorm:"not null;uniqueIndex:uk_dict_item_translation,priority:1;comment:字典项ID"`
	Locale      string    `json:"locale" gorm:"not null;size:10;uniqueIndex:uk_dict_item_translation,priority:2;index;comment:语言标识，主子标签形式，如 zh、en"`
	Label       string    `json:"label" gorm:"size:50;comment:该语言字典项显示名"`
	Description string    `json:"description" gorm:"size:200;comment:该语言字典项描述"`
	CreatedAt   time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:最后更新时间"`

	// 关联关系
	Item DictItem `json:"-" gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (DictItemTranslation) TableName() string {
	return "dict_item_translations"
}
