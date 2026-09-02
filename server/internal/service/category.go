// Package service 业务逻辑层
package service

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// CategoryServiceInterface 分类服务接口
type CategoryServiceInterface interface {
	// 基础CRUD操作
	CreateCategory(req *CreateCategoryRequest, operatorID uint) (*model.Category, error)
	UpdateCategory(req *UpdateCategoryRequest, operatorID uint) (*model.Category, error)
	DeleteCategory(id uint, operatorID uint) error
	GetCategory(id uint) (*model.Category, error)

	// 查询操作
	ListCategories(req *ListCategoriesRequest) (*CategoryListResponse, error)
	GetCategoryTree() ([]*CategoryTreeNode, error)
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Slug        string `json:"slug" binding:"omitempty,max=50"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	CoverImage  string `json:"coverImage" binding:"omitempty,max=255"`
	ParentID    *uint  `json:"parentId"`
	SortOrder   *int   `json:"sortOrder"`
	Status      *int   `json:"status" binding:"omitempty,oneof=0 1"`
	IsFeatured  *bool  `json:"isFeatured"`
	SEOTitle       string `json:"seoTitle" binding:"omitempty,max=100"`
	SEODescription string `json:"seoDescription" binding:"omitempty,max=255"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Name        *string `json:"name" binding:"omitempty,min=1,max=50"`
	Slug        *string `json:"slug" binding:"omitempty,max=50"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	CoverImage  *string `json:"coverImage" binding:"omitempty,max=255"`
	SortOrder   *int   `json:"sortOrder"`
	Status      *int   `json:"status" binding:"omitempty,oneof=0 1"`
	IsFeatured  *bool  `json:"isFeatured"`
	SEOTitle       *string `json:"seoTitle" binding:"omitempty,max=100"`
	SEODescription *string `json:"seoDescription" binding:"omitempty,max=255"`
}

// ListCategoriesRequest 分类列表请求
type ListCategoriesRequest struct {
	Page     int    `json:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"pageSize" binding:"omitempty,min=1,max=100"`
	Status   *int   `json:"status"`
	Search   string `json:"search"`
}

// CategoryListResponse 分类列表响应
type CategoryListResponse struct {
	Categories []*model.Category `json:"categories"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
}

// CategoryTreeNode 分类树节点
type CategoryTreeNode struct {
	*model.Category
	Children []*CategoryTreeNode `json:"children,omitempty"`
}

// CategoryService 分类服务实现
type CategoryService struct {
	categoryRepo repository.CategoryRepositoryInterface
}

// NewCategoryService 创建分类服务实例
func NewCategoryService(categoryRepo repository.CategoryRepositoryInterface) CategoryServiceInterface {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(req *CreateCategoryRequest, operatorID uint) (*model.Category, error) {
	// 校验父分类存在性并推导层级字段。
	parent, err := s.resolveParent(req.ParentID)
	if err != nil {
		return nil, err
	}

	// 构建分类对象，默认状态为显示。
	status := model.CategoryStatusShown
	if req.Status != nil {
		status = *req.Status
	}
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	isFeatured := false
	if req.IsFeatured != nil {
		isFeatured = *req.IsFeatured
	}

	category := &model.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		ParentID:    req.ParentID,
		SortOrder:   sortOrder,
		Status:      status,
		IsFeatured:  isFeatured,
		SEOTitle:    req.SEOTitle,
		SEODescription: req.SEODescription,
	}

	// 有父分类时继承父分类的 root 与 level。
	if parent != nil {
		rootID := parent.ID
		if parent.RootID != nil {
			rootID = *parent.RootID
		}
		category.RootID = &rootID
		category.Level = parent.Level + 1
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, fmt.Errorf("创建分类失败: %w", err)
	}

	// 首次创建后回填物化路径，便于一次查询整棵子树。
	if err := s.updateCategoryPath(category); err != nil {
		return nil, err
	}

	return s.categoryRepo.GetByID(category.ID)
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(req *UpdateCategoryRequest, operatorID uint) (*model.Category, error) {
	category, err := s.categoryRepo.GetByID(req.ID)
	if err != nil {
		return nil, err
	}

	// 名称与描述类字段仅在显式传入时更新，省略时保留原值。
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Slug != nil {
		category.Slug = *req.Slug
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.CoverImage != nil {
		category.CoverImage = *req.CoverImage
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		category.Status = *req.Status
	}
	if req.IsFeatured != nil {
		category.IsFeatured = *req.IsFeatured
	}
	if req.SEOTitle != nil {
		category.SEOTitle = *req.SEOTitle
	}
	if req.SEODescription != nil {
		category.SEODescription = *req.SEODescription
	}

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, fmt.Errorf("更新分类失败: %w", err)
	}

	return s.categoryRepo.GetByID(category.ID)
}

// DeleteCategory 删除分类
func (s *CategoryService) DeleteCategory(id uint, operatorID uint) error {
	// 校验分类存在。
	if _, err := s.categoryRepo.GetByID(id); err != nil {
		return err
	}

	// 存在子分类时禁止删除，需先处理子分类。
	childCount, err := s.categoryRepo.CountByParentID(id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("存在子分类，无法删除")
	}

	return s.categoryRepo.Delete(id)
}

// GetCategory 根据ID获取分类
func (s *CategoryService) GetCategory(id uint) (*model.Category, error) {
	return s.categoryRepo.GetByID(id)
}

// ListCategories 分页查询分类列表
func (s *CategoryService) ListCategories(req *ListCategoriesRequest) (*CategoryListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	params := &repository.CategoryListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
		Search:   req.Search,
	}

	categories, total, err := s.categoryRepo.List(params)
	if err != nil {
		return nil, err
	}

	return &CategoryListResponse{
		Categories: categories,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// GetCategoryTree 获取分类树
func (s *CategoryService) GetCategoryTree() ([]*CategoryTreeNode, error) {
	categories, err := s.categoryRepo.ListAll()
	if err != nil {
		return nil, err
	}

	return buildCategoryTree(categories), nil
}

// resolveParent 校验并返回父分类，父分类为空时返回 nil。
func (s *CategoryService) resolveParent(parentID *uint) (*model.Category, error) {
	if parentID == nil {
		return nil, nil
	}

	parent, err := s.categoryRepo.GetByID(*parentID)
	if err != nil {
		return nil, errors.New("父分类不存在")
	}
	return parent, nil
}

// updateCategoryPath 回填分类物化路径，顶级为 /ID，子级为父路径 + /ID。
func (s *CategoryService) updateCategoryPath(category *model.Category) error {
	path := fmt.Sprintf("/%d", category.ID)
	if category.ParentID != nil {
		parent, err := s.categoryRepo.GetByID(*category.ParentID)
		if err != nil {
			return errors.New("父分类不存在")
		}
		path = parent.Path + fmt.Sprintf("/%d", category.ID)
	}

	// 直接更新物化路径字段，避免触发完整 Save 覆盖其他字段。
	return s.categoryRepo.UpdatePath(category.ID, path)
}

// buildCategoryTree 将扁平分类列表构建为树形结构，仅保留叶子节点可见状态由调用方决定。
func buildCategoryTree(categories []*model.Category) []*CategoryTreeNode {
	// 按 ID 建立索引，方便快速查找节点。
	nodeMap := make(map[uint]*CategoryTreeNode, len(categories))
	for _, category := range categories {
		nodeMap[category.ID] = &CategoryTreeNode{Category: category}
	}

	var roots []*CategoryTreeNode
	for _, category := range categories {
		node := nodeMap[category.ID]
		if category.ParentID != nil {
			// 父节点存在时挂载为子节点，父节点缺失时按根节点处理。
			if parentNode, exists := nodeMap[*category.ParentID]; exists {
				parentNode.Children = append(parentNode.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	return roots
}
