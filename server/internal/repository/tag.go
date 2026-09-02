// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"
	"MyBlog/pkg/slug"

	"gorm.io/gorm"
)

// ErrTagNotFound 标签不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrTagNotFound = errors.New("标签不存在")

// TagRepositoryInterface 标签仓储接口
type TagRepositoryInterface interface {
	// 基础CRUD操作
	Create(tag *model.Tag) error
	GetByID(id uint) (*model.Tag, error)
	GetByName(name string) (*model.Tag, error)
	Update(tag *model.Tag) error
	Delete(id uint) error

	// 查询操作
	List(params *TagListParams) ([]*model.Tag, int64, error)
	GetPopular(limit int) ([]*model.Tag, error)

	// 工具方法
	EnsureUniqueSlug(tag *model.Tag) error
}

// TagListParams 标签列表查询参数
type TagListParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Status   *int   `json:"status"`
	Search   string `json:"search"`
	IsHot    *bool  `json:"isHot"`
}

// TagRepository 标签仓储实现
type TagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓储实例
func NewTagRepository(db *gorm.DB) TagRepositoryInterface {
	return &TagRepository{db: db}
}

// Create 创建标签
func (r *TagRepository) Create(tag *model.Tag) error {
	// 标签 slug 为空时按名称生成，确保唯一索引不写入空值。
	if tag.Slug == "" {
		tag.Slug = slug.Generate("tag", tag.Name)
	}

	if err := r.EnsureUniqueSlug(tag); err != nil {
		return err
	}

	return r.db.Create(tag).Error
}

// GetByID 根据ID获取标签
func (r *TagRepository) GetByID(id uint) (*model.Tag, error) {
	var tag model.Tag
	if err := r.db.First(&tag, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}
	return &tag, nil
}

// GetByName 根据名称获取标签
func (r *TagRepository) GetByName(name string) (*model.Tag, error) {
	var tag model.Tag
	if err := r.db.Where("name = ?", name).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}
	return &tag, nil
}

// Update 更新标签
func (r *TagRepository) Update(tag *model.Tag) error {
	// 更新时若 slug 为空，回退为按名称生成，避免唯一索引写入空值。
	if tag.Slug == "" {
		tag.Slug = slug.Generate("tag", tag.Name)
	}

	if err := r.EnsureUniqueSlug(tag); err != nil {
		return err
	}

	return r.db.Save(tag).Error
}

// Delete 删除标签，采用软删除。
func (r *TagRepository) Delete(id uint) error {
	return r.db.Delete(&model.Tag{}, id).Error
}

// List 分页查询标签列表
func (r *TagRepository) List(params *TagListParams) ([]*model.Tag, int64, error) {
	query := r.db.Model(&model.Tag{})

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.IsHot != nil {
		query = query.Where("is_hot = ?", *params.IsHot)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询标签总数失败: %w", err)
	}

	// 设置分页默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var tags []*model.Tag
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("usage_count DESC, id ASC").Offset(offset).Limit(params.PageSize).Find(&tags).Error; err != nil {
		return nil, 0, fmt.Errorf("查询标签列表失败: %w", err)
	}

	return tags, total, nil
}

// GetPopular 获取热门标签，按使用次数倒序。
func (r *TagRepository) GetPopular(limit int) ([]*model.Tag, error) {
	var tags []*model.Tag
	if err := r.db.Where("status = ?", model.TagStatusEnabled).
		Order("usage_count DESC, id ASC").
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("查询热门标签失败: %w", err)
	}
	return tags, nil
}

// EnsureUniqueSlug 确保标签 slug 唯一，冲突时追加序号。
func (r *TagRepository) EnsureUniqueSlug(tag *model.Tag) error {
	originalSlug := tag.Slug
	counter := 1

	for {
		var count int64
		query := r.db.Model(&model.Tag{}).Where("slug = ?", tag.Slug)

		// 更新操作时排除当前标签自身。
		if tag.ID != 0 {
			query = query.Where("id != ?", tag.ID)
		}

		if err := query.Count(&count).Error; err != nil {
			return fmt.Errorf("检查标签 slug 冲突失败: %w", err)
		}

		if count == 0 {
			break
		}

		tag.Slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	return nil
}
