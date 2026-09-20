package database

import (
	"encoding/json"
	"fmt"

	"MyBlog/internal/model"

	"gorm.io/gorm"
)

// 字典种子数据定义。
// 种子仅写入主列缺省语言内容，各语言翻译由管理端按需配置；
// 展示元数据存入 extra 字段，渲染层经字典动态读取，不引入第三处硬编码。

// dictTypeSeed 单个字典类型的种子数据，携带其全部字典项。
type dictTypeSeed struct {
	Type  model.DictType
	Items []model.DictItem
}

// DictTypeCodeTagStatus 标签状态字典码，标签启用状态的展示文案与样式经该字典读取。
const DictTypeCodeTagStatus = "tag_status"

// dictTypeSeeds 全部字典种子数据，新增种子在此追加并保持幂等语义。
var dictTypeSeeds = []dictTypeSeed{
	{
		Type: model.DictType{
			Code:        DictTypeCodeTagStatus,
			Name:        "标签状态",
			Description: "标签启用与隐藏状态的展示配置",
			Status:      model.DictStatusEnabled,
			SortOrder:   1,
		},
		Items: []model.DictItem{
			{
				Value:       "1",
				Label:       "启用",
				Description: "标签对外展示并可用于文章挂载",
				Status:      model.DictStatusEnabled,
				SortOrder:   1,
				Extra:       json.RawMessage(`{"variant":"default"}`),
			},
			{
				Value:       "0",
				Label:       "隐藏",
				Description: "标签不对外展示，已挂载文章不可再选择",
				Status:      model.DictStatusEnabled,
				SortOrder:   2,
				Extra:       json.RawMessage(`{"variant":"secondary"}`),
			},
		},
	},
}

// SeedDicts 幂等写入字典种子数据，已存在的字典码整组跳过，重复执行无副作用。
func SeedDicts(db *gorm.DB) error {
	for _, seed := range dictTypeSeeds {
		if err := seedDictType(db, seed); err != nil {
			return err
		}
	}
	return nil
}

// seedDictType 写入单个字典类型种子，类型与字典项在同一事务内提交。
func seedDictType(db *gorm.DB, seed dictTypeSeed) error {
	var count int64
	if err := db.Model(&model.DictType{}).Where("code = ?", seed.Type.Code).Count(&count).Error; err != nil {
		return fmt.Errorf("查询字典种子 %s 失败: %w", seed.Type.Code, err)
	}
	if count > 0 {
		return nil
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&seed.Type).Error; err != nil {
			return fmt.Errorf("写入字典类型种子 %s 失败: %w", seed.Type.Code, err)
		}
		if len(seed.Items) == 0 {
			return nil
		}
		for index := range seed.Items {
			seed.Items[index].TypeID = seed.Type.ID
		}
		if err := tx.Create(&seed.Items).Error; err != nil {
			return fmt.Errorf("写入字典项种子 %s 失败: %w", seed.Type.Code, err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("字典种子 %s 写入成功\n", seed.Type.Code)
	return nil
}
