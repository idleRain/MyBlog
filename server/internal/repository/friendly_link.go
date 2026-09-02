// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"

	"gorm.io/gorm"
)

// ErrFriendlyLinkNotFound 友情链接不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrFriendlyLinkNotFound = errors.New("友情链接不存在")

// FriendlyLinkRepositoryInterface 友情链接仓储接口
type FriendlyLinkRepositoryInterface interface {
	// 基础CRUD操作
	Create(link *model.FriendlyLink) error
	GetByID(id uint) (*model.FriendlyLink, error)
	GetByURL(url string) (*model.FriendlyLink, error)
	Update(link *model.FriendlyLink) error
	Delete(id uint) error
	UpdateStatus(id uint, status model.LinkStatus) error

	// 查询操作
	List(params *FriendlyLinkListParams) ([]*model.FriendlyLink, int64, error)
	ListVisible() ([]*model.FriendlyLink, error)
}

// FriendlyLinkListParams 友情链接列表查询参数
type FriendlyLinkListParams struct {
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	Status   model.LinkStatus `json:"status"`
}

// FriendlyLinkRepository 友情链接仓储实现
type FriendlyLinkRepository struct {
	db *gorm.DB
}

// NewFriendlyLinkRepository 创建友情链接仓储实例
func NewFriendlyLinkRepository(db *gorm.DB) FriendlyLinkRepositoryInterface {
	return &FriendlyLinkRepository{db: db}
}

// Create 创建友情链接
func (r *FriendlyLinkRepository) Create(link *model.FriendlyLink) error {
	return r.db.Create(link).Error
}

// GetByID 根据ID获取友情链接
func (r *FriendlyLinkRepository) GetByID(id uint) (*model.FriendlyLink, error) {
	var link model.FriendlyLink
	if err := r.db.First(&link, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFriendlyLinkNotFound
		}
		return nil, fmt.Errorf("查询友情链接失败: %w", err)
	}
	return &link, nil
}

// GetByURL 根据站点URL获取友情链接
func (r *FriendlyLinkRepository) GetByURL(url string) (*model.FriendlyLink, error) {
	var link model.FriendlyLink
	if err := r.db.Where("url = ?", url).First(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFriendlyLinkNotFound
		}
		return nil, fmt.Errorf("查询友情链接失败: %w", err)
	}
	return &link, nil
}

// Update 更新友情链接
func (r *FriendlyLinkRepository) Update(link *model.FriendlyLink) error {
	return r.db.Save(link).Error
}

// Delete 删除友情链接，采用软删除。
func (r *FriendlyLinkRepository) Delete(id uint) error {
	return r.db.Delete(&model.FriendlyLink{}, id).Error
}

// UpdateStatus 更新友情链接状态。
func (r *FriendlyLinkRepository) UpdateStatus(id uint, status model.LinkStatus) error {
	if err := r.db.Model(&model.FriendlyLink{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("更新友情链接状态失败: %w", err)
	}
	return nil
}

// List 分页查询友情链接列表。
func (r *FriendlyLinkRepository) List(params *FriendlyLinkListParams) ([]*model.FriendlyLink, int64, error) {
	query := r.db.Model(&model.FriendlyLink{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询友情链接总数失败: %w", err)
	}

	// 设置分页默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var links []*model.FriendlyLink
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("sort_order ASC, id ASC").
		Offset(offset).Limit(params.PageSize).
		Find(&links).Error; err != nil {
		return nil, 0, fmt.Errorf("查询友情链接列表失败: %w", err)
	}

	return links, total, nil
}

// ListVisible 查询全部对外展示的友情链接。
func (r *FriendlyLinkRepository) ListVisible() ([]*model.FriendlyLink, error) {
	var links []*model.FriendlyLink
	if err := r.db.Where("status = ?", model.LinkStatusActive).
		Order("sort_order ASC, id ASC").
		Find(&links).Error; err != nil {
		return nil, fmt.Errorf("查询友情链接列表失败: %w", err)
	}
	return links, nil
}
