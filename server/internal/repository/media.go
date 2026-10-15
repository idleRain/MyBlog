// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"

	"gorm.io/gorm"
)

// MediaRepositoryInterface 媒体仓储接口
type MediaRepositoryInterface interface {
	// 基础CRUD操作
	Create(media *model.MediaFile) error
	GetByID(id uint) (*model.MediaFile, error)
	GetByFileHashAndUploader(hash string, uploaderID uint) (*model.MediaFile, error)
	Update(media *model.MediaFile) error
	Delete(id uint) error

	// 查询操作
	List(params *MediaListParams) ([]*model.MediaFile, int64, error)
}

// MediaListParams 媒体列表查询参数
type MediaListParams struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	Folder     string `json:"folder"`
	MimeType   string `json:"mimeType"`
	UploaderID *uint  `json:"uploaderId"`
}

// MediaRepository 媒体仓储实现
type MediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository 创建媒体仓储实例
func NewMediaRepository(db *gorm.DB) MediaRepositoryInterface {
	return &MediaRepository{db: db}
}

// Create 创建媒体文件记录
func (r *MediaRepository) Create(media *model.MediaFile) error {
	return r.db.Create(media).Error
}

// GetByID 根据ID获取媒体文件
func (r *MediaRepository) GetByID(id uint) (*model.MediaFile, error) {
	var media model.MediaFile
	if err := r.db.Preload("Uploader").First(&media, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMediaNotFound
		}
		return nil, fmt.Errorf("查询媒体文件失败: %w", err)
	}
	return &media, nil
}

// GetByFileHashAndUploader 按文件哈希与上传者查询媒体文件，用于同一上传者范围内的秒传去重。
// 同一内容可能被多个上传者各持一份记录，限定上传者使命中落在调用方自己的记录上，
// 避免非首位上传者每次重复上传都新建副本。
func (r *MediaRepository) GetByFileHashAndUploader(hash string, uploaderID uint) (*model.MediaFile, error) {
	var media model.MediaFile
	if err := r.db.Where("file_hash = ?", hash).Where("uploader_id = ?", uploaderID).First(&media).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMediaNotFound
		}
		return nil, fmt.Errorf("查询媒体文件失败: %w", err)
	}
	return &media, nil
}

// Update 更新媒体文件记录
func (r *MediaRepository) Update(media *model.MediaFile) error {
	return r.db.Save(media).Error
}

// Delete 删除媒体文件，采用软删除。
func (r *MediaRepository) Delete(id uint) error {
	return r.db.Delete(&model.MediaFile{}, id).Error
}

// List 分页查询媒体文件列表
func (r *MediaRepository) List(params *MediaListParams) ([]*model.MediaFile, int64, error) {
	query := r.db.Model(&model.MediaFile{})

	if params.Folder != "" {
		query = query.Where("folder = ?", params.Folder)
	}
	if params.MimeType != "" {
		query = query.Where("mime_type LIKE ?", params.MimeType+"%")
	}
	if params.UploaderID != nil {
		query = query.Where("uploader_id = ?", *params.UploaderID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询媒体文件总数失败: %w", err)
	}

	// 设置分页默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var mediaFiles []*model.MediaFile
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(params.PageSize).
		Find(&mediaFiles).Error; err != nil {
		return nil, 0, fmt.Errorf("查询媒体文件列表失败: %w", err)
	}

	return mediaFiles, total, nil
}
