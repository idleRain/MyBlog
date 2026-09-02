// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"
	"MyBlog/pkg/slug"

	"gorm.io/gorm"
)

// ErrCategoryNotFound 分类不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrCategoryNotFound = errors.New("分类不存在")

// CategoryRepositoryInterface 分类仓储接口
type CategoryRepositoryInterface interface {
	// 基础CRUD操作
	Create(category *model.Category) error
	GetByID(id uint) (*model.Category, error)
	GetBySlug(slug string) (*model.Category, error)
	Update(category *model.Category) error
	Delete(id uint) error

	// 查询操作
	List(params *CategoryListParams) ([]*model.Category, int64, error)
	ListAll() ([]*model.Category, error)
	GetByParentID(parentID uint) ([]*model.Category, error)
	CountByParentID(parentID uint) (int64, error)

	// 工具方法
	EnsureUniqueSlug(category *model.Category) error
	UpdatePath(id uint, path string) error
}

// CategoryListParams 分类列表查询参数
type CategoryListParams struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	Status   *int `json:"status"`
	Search   string `json:"search"`
}

// CategoryRepository 分类仓储实现
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓储实例
func NewCategoryRepository(db *gorm.DB) CategoryRepositoryInterface {
	return &CategoryRepository{db: db}
}

// Create 创建分类
func (r *CategoryRepository) Create(category *model.Category) error {
	// 分类 slug 为空时按名称生成，确保唯一索引不写入空值。
	if category.Slug == "" {
		category.Slug = slug.Generate("category", category.Name)
	}

	if err := r.EnsureUniqueSlug(category); err != nil {
		return err
	}

	return r.db.Create(category).Error
}

// GetByID 根据ID获取分类
func (r *CategoryRepository) GetByID(id uint) (*model.Category, error) {
	var category model.Category
	if err := r.db.First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}
	return &category, nil
}

// GetBySlug 根据Slug获取分类
func (r *CategoryRepository) GetBySlug(slug string) (*model.Category, error) {
	var category model.Category
	if err := r.db.Where("slug = ?", slug).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}
	return &category, nil
}

// Update 更新分类
func (r *CategoryRepository) Update(category *model.Category) error {
	// 更新时若 slug 为空，回退为按名称生成，避免唯一索引写入空值。
	if category.Slug == "" {
		category.Slug = slug.Generate("category", category.Name)
	}

	if err := r.EnsureUniqueSlug(category); err != nil {
		return err
	}

	return r.db.Save(category).Error
}

// Delete 删除分类，采用软删除。
func (r *CategoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

// List 分页查询分类列表
func (r *CategoryRepository) List(params *CategoryListParams) ([]*model.Category, int64, error) {
	query := r.db.Model(&model.Category{})

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询分类总数失败: %w", err)
	}

	// 设置分页默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var categories []*model.Category
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("sort_order ASC, id ASC").Offset(offset).Limit(params.PageSize).Find(&categories).Error; err != nil {
		return nil, 0, fmt.Errorf("查询分类列表失败: %w", err)
	}

	return categories, total, nil
}

// ListAll 查询全部分类，用于构建分类树。
func (r *CategoryRepository) ListAll() ([]*model.Category, error) {
	var categories []*model.Category
	if err := r.db.Order("sort_order ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询分类列表失败: %w", err)
	}
	return categories, nil
}

// GetByParentID 查询指定父分类下的直接子分类。
func (r *CategoryRepository) GetByParentID(parentID uint) ([]*model.Category, error) {
	var categories []*model.Category
	if err := r.db.Where("parent_id = ?", parentID).Order("sort_order ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询子分类失败: %w", err)
	}
	return categories, nil
}

// CountByParentID 统计指定父分类下的子分类数量。
func (r *CategoryRepository) CountByParentID(parentID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Category{}).Where("parent_id = ?", parentID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计子分类失败: %w", err)
	}
	return count, nil
}

// EnsureUniqueSlug 确保分类 slug 唯一，冲突时追加序号。
func (r *CategoryRepository) EnsureUniqueSlug(category *model.Category) error {
	originalSlug := category.Slug
	counter := 1

	for {
		var count int64
		query := r.db.Model(&model.Category{}).Where("slug = ?", category.Slug)

		// 更新操作时排除当前分类自身。
		if category.ID != 0 {
			query = query.Where("id != ?", category.ID)
		}

		if err := query.Count(&count).Error; err != nil {
			return fmt.Errorf("检查分类 slug 冲突失败: %w", err)
		}

		if count == 0 {
			break
		}

		category.Slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	return nil
}

// UpdatePath 更新分类物化路径，用于创建后回填子树路径。
func (r *CategoryRepository) UpdatePath(id uint, path string) error {
	if err := r.db.Model(&model.Category{}).Where("id = ?", id).Update("path", path).Error; err != nil {
		return fmt.Errorf("更新分类路径失败: %w", err)
	}
	return nil
}
