// Package service 业务逻辑层
package service

import (
	"fmt"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// TagServiceInterface 标签服务接口
type TagServiceInterface interface {
	// 基础CRUD操作
	CreateTag(req *CreateTagRequest, operatorID uint) (*model.Tag, error)
	UpdateTag(req *UpdateTagRequest, operatorID uint) (*model.Tag, error)
	DeleteTag(id uint, operatorID uint) error
	GetTag(id uint) (*model.Tag, error)

	// 查询操作
	ListTags(req *ListTagsRequest) (*TagListResponse, error)
	GetPopularTags(limit int) ([]*model.Tag, error)
	ListAllTags() ([]*model.Tag, error)
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=30"`
	Slug        string `json:"slug" binding:"omitempty,max=30"`
	Color       string `json:"color" binding:"omitempty,min=4,max=7"`
	Description string `json:"description" binding:"omitempty,max=200"`
	Status      *int   `json:"status" binding:"omitempty,oneof=0 1"`
	IsHot       *bool  `json:"isHot"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	ID          uint    `json:"id" binding:"required"`
	Name        *string `json:"name" binding:"omitempty,min=1,max=30"`
	Slug        *string `json:"slug" binding:"omitempty,max=30"`
	Color       *string `json:"color" binding:"omitempty,min=4,max=7"`
	Description *string `json:"description" binding:"omitempty,max=200"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
	IsHot       *bool   `json:"isHot"`
}

// ListTagsRequest 标签列表请求
type ListTagsRequest struct {
	Page     int    `json:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"pageSize" binding:"omitempty,min=1,max=100"`
	Status   *int   `json:"status"`
	IsHot    *bool  `json:"isHot"`
	Search   string `json:"search"`
}

// TagListResponse 标签列表响应
type TagListResponse struct {
	Tags     []*model.Tag `json:"tags"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

// TagService 标签服务实现
type TagService struct {
	tagRepo repository.TagRepositoryInterface
}

// NewTagService 创建标签服务实例
func NewTagService(tagRepo repository.TagRepositoryInterface) TagServiceInterface {
	return &TagService{
		tagRepo: tagRepo,
	}
}

// CreateTag 创建标签
func (s *TagService) CreateTag(req *CreateTagRequest, operatorID uint) (*model.Tag, error) {
	// 校验标签名称唯一。
	if _, err := s.tagRepo.GetByName(req.Name); err == nil {
		return nil, fmt.Errorf("标签名称已存在")
	}

	// 默认状态为启用。
	status := model.TagStatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	isHot := false
	if req.IsHot != nil {
		isHot = *req.IsHot
	}

	tag := &model.Tag{
		Name:        req.Name,
		Slug:        req.Slug,
		Color:       req.Color,
		Description: req.Description,
		Status:      status,
		IsHot:       isHot,
	}

	// 颜色未指定时使用默认灰色。
	if tag.Color == "" {
		tag.Color = "#808080"
	}

	if err := s.tagRepo.Create(tag); err != nil {
		return nil, fmt.Errorf("创建标签失败: %w", err)
	}

	return s.tagRepo.GetByID(tag.ID)
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(req *UpdateTagRequest, operatorID uint) (*model.Tag, error) {
	tag, err := s.tagRepo.GetByID(req.ID)
	if err != nil {
		return nil, err
	}

	// 名称变更时校验唯一性。
	if req.Name != nil && *req.Name != tag.Name {
		if existing, err := s.tagRepo.GetByName(*req.Name); err == nil && existing.ID != req.ID {
			return nil, fmt.Errorf("标签名称已存在")
		}
	}

	// 可选字段仅在显式传入时更新，省略时保留原值。
	if req.Name != nil {
		tag.Name = *req.Name
	}
	if req.Slug != nil {
		tag.Slug = *req.Slug
	}
	if req.Color != nil {
		tag.Color = *req.Color
	}
	if req.Description != nil {
		tag.Description = *req.Description
	}
	if req.Status != nil {
		tag.Status = *req.Status
	}
	if req.IsHot != nil {
		tag.IsHot = *req.IsHot
	}

	if err := s.tagRepo.Update(tag); err != nil {
		return nil, fmt.Errorf("更新标签失败: %w", err)
	}

	return s.tagRepo.GetByID(tag.ID)
}

// DeleteTag 删除标签
func (s *TagService) DeleteTag(id uint, operatorID uint) error {
	if _, err := s.tagRepo.GetByID(id); err != nil {
		return err
	}
	return s.tagRepo.Delete(id)
}

// GetTag 根据ID获取标签
func (s *TagService) GetTag(id uint) (*model.Tag, error) {
	return s.tagRepo.GetByID(id)
}

// ListTags 分页查询标签列表
func (s *TagService) ListTags(req *ListTagsRequest) (*TagListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	params := &repository.TagListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
		IsHot:    req.IsHot,
		Search:   req.Search,
	}

	tags, total, err := s.tagRepo.List(params)
	if err != nil {
		return nil, err
	}

	return &TagListResponse{
		Tags:     tags,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetPopularTags 获取热门标签。
func (s *TagService) GetPopularTags(limit int) ([]*model.Tag, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.tagRepo.GetPopular(limit)
}

// ListAllTags 获取全部标签，供文章编辑选择使用，不含分页。
func (s *TagService) ListAllTags() ([]*model.Tag, error) {
	return s.tagRepo.ListAll()
}
