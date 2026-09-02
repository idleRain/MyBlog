// Package service 业务逻辑层
package service

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// FriendlyLinkServiceInterface 友情链接服务接口
type FriendlyLinkServiceInterface interface {
	// 基础CRUD操作
	CreateLink(req *CreateFriendlyLinkRequest) (*model.FriendlyLink, error)
	UpdateLink(req *UpdateFriendlyLinkRequest) (*model.FriendlyLink, error)
	DeleteLink(id uint, operatorID uint) error

	// 状态流转
	ApproveLink(id uint, operatorID uint) error
	HideLink(id uint, operatorID uint) error
	RejectLink(id uint, operatorID uint) error

	// 查询操作
	ListLinks(req *ListFriendlyLinksRequest) (*FriendlyLinkListResponse, error)
	ListVisibleLinks() ([]*model.FriendlyLink, error)
}

// CreateFriendlyLinkRequest 创建友情链接请求
type CreateFriendlyLinkRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=50"`
	URL          string `json:"url" binding:"required,max=255"`
	Logo         string `json:"logo" binding:"omitempty,max=500"`
	Description  string `json:"description" binding:"omitempty,max=255"`
	ContactEmail string `json:"contactEmail" binding:"omitempty,email,max=100"`
	SortOrder    *int   `json:"sortOrder"`
	IsReciprocal *bool  `json:"isReciprocal"`
}

// UpdateFriendlyLinkRequest 更新友情链接请求
type UpdateFriendlyLinkRequest struct {
	ID           uint    `json:"id" binding:"required"`
	Name         *string `json:"name" binding:"omitempty,min=1,max=50"`
	URL          *string `json:"url" binding:"omitempty,max=255"`
	Logo         *string `json:"logo" binding:"omitempty,max=500"`
	Description  *string `json:"description" binding:"omitempty,max=255"`
	ContactEmail *string `json:"contactEmail" binding:"omitempty,email,max=100"`
	SortOrder    *int    `json:"sortOrder"`
	IsReciprocal *bool   `json:"isReciprocal"`
}

// ListFriendlyLinksRequest 友情链接列表请求
type ListFriendlyLinksRequest struct {
	Page     int    `json:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"pageSize" binding:"omitempty,min=1,max=100"`
	Status   string `json:"status" binding:"omitempty,oneof=pending active hidden rejected"`
}

// FriendlyLinkListResponse 友情链接列表响应
type FriendlyLinkListResponse struct {
	Links    []*model.FriendlyLink `json:"links"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}

// FriendlyLinkService 友情链接服务实现
type FriendlyLinkService struct {
	linkRepo repository.FriendlyLinkRepositoryInterface
}

// NewFriendlyLinkService 创建友情链接服务实例
func NewFriendlyLinkService(linkRepo repository.FriendlyLinkRepositoryInterface) FriendlyLinkServiceInterface {
	return &FriendlyLinkService{
		linkRepo: linkRepo,
	}
}

// CreateLink 创建友情链接，新链接默认待审核。
func (s *FriendlyLinkService) CreateLink(req *CreateFriendlyLinkRequest) (*model.FriendlyLink, error) {
	// 校验站点 URL 唯一。
	if existing, err := s.linkRepo.GetByURL(req.URL); err == nil && existing != nil {
		return nil, errors.New("该站点 URL 已存在")
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	isReciprocal := false
	if req.IsReciprocal != nil {
		isReciprocal = *req.IsReciprocal
	}

	link := &model.FriendlyLink{
		Name:         req.Name,
		URL:          req.URL,
		Logo:         req.Logo,
		Description:  req.Description,
		ContactEmail: req.ContactEmail,
		SortOrder:    sortOrder,
		Status:       model.LinkStatusPending,
		IsReciprocal: isReciprocal,
	}

	if err := s.linkRepo.Create(link); err != nil {
		return nil, fmt.Errorf("创建友情链接失败: %w", err)
	}

	return s.linkRepo.GetByID(link.ID)
}

// UpdateLink 更新友情链接。
func (s *FriendlyLinkService) UpdateLink(req *UpdateFriendlyLinkRequest) (*model.FriendlyLink, error) {
	link, err := s.linkRepo.GetByID(req.ID)
	if err != nil {
		return nil, err
	}

	// 可选字段仅在显式传入时更新，省略时保留原值。
	if req.Name != nil {
		link.Name = *req.Name
	}
	if req.URL != nil {
		link.URL = *req.URL
	}
	if req.Logo != nil {
		link.Logo = *req.Logo
	}
	if req.Description != nil {
		link.Description = *req.Description
	}
	if req.ContactEmail != nil {
		link.ContactEmail = *req.ContactEmail
	}
	if req.SortOrder != nil {
		link.SortOrder = *req.SortOrder
	}
	if req.IsReciprocal != nil {
		link.IsReciprocal = *req.IsReciprocal
	}

	if err := s.linkRepo.Update(link); err != nil {
		return nil, fmt.Errorf("更新友情链接失败: %w", err)
	}

	return s.linkRepo.GetByID(link.ID)
}

// DeleteLink 删除友情链接。
func (s *FriendlyLinkService) DeleteLink(id uint, operatorID uint) error {
	if _, err := s.linkRepo.GetByID(id); err != nil {
		return err
	}
	return s.linkRepo.Delete(id)
}

// ApproveLink 审核通过友情链接。
func (s *FriendlyLinkService) ApproveLink(id uint, operatorID uint) error {
	return s.transitionStatus(id, model.LinkStatusActive)
}

// HideLink 下架友情链接。
func (s *FriendlyLinkService) HideLink(id uint, operatorID uint) error {
	return s.transitionStatus(id, model.LinkStatusHidden)
}

// RejectLink 拒绝友情链接。
func (s *FriendlyLinkService) RejectLink(id uint, operatorID uint) error {
	return s.transitionStatus(id, model.LinkStatusRejected)
}

// ListLinks 分页查询友情链接列表。
func (s *FriendlyLinkService) ListLinks(req *ListFriendlyLinksRequest) (*FriendlyLinkListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	params := &repository.FriendlyLinkListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   model.LinkStatus(req.Status),
	}

	links, total, err := s.linkRepo.List(params)
	if err != nil {
		return nil, err
	}

	return &FriendlyLinkListResponse{
		Links:    links,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// ListVisibleLinks 获取对外展示的友情链接。
func (s *FriendlyLinkService) ListVisibleLinks() ([]*model.FriendlyLink, error) {
	return s.linkRepo.ListVisible()
}

// transitionStatus 执行友情链接状态流转。
func (s *FriendlyLinkService) transitionStatus(id uint, status model.LinkStatus) error {
	if _, err := s.linkRepo.GetByID(id); err != nil {
		return err
	}
	return s.linkRepo.UpdateStatus(id, status)
}
