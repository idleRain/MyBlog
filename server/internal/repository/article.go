package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"MyBlog/internal/model"

	"gorm.io/gorm"
)

// ArticleRepositoryInterface 文章仓储接口
type ArticleRepositoryInterface interface {
	// 基础CRUD操作
	Create(article *model.Article) error
	GetByID(id uint) (*model.Article, error)
	GetBySlug(slug string) (*model.Article, error)
	Update(article *model.Article) error
	Delete(id uint) error

	// 查询操作
	List(params *ArticleListParams) ([]*model.Article, int64, error)
	GetByAuthor(authorID uint, params *ArticleListParams) ([]*model.Article, int64, error)
	GetByCategory(categoryID uint, params *ArticleListParams) ([]*model.Article, int64, error)
	GetByTag(tagID uint, params *ArticleListParams) ([]*model.Article, int64, error)
	Search(keyword string, params *ArticleListParams) ([]*model.Article, int64, error)

	// 统计操作
	GetPopular(limit int) ([]*model.Article, error)
	GetRecent(limit int) ([]*model.Article, error)
	IncrementViewCount(id uint) error
	UpdateLikeCount(id uint) error
	UpdateCommentCount(id uint) error

	// 分类和标签关联
	AddCategory(articleID, categoryID uint) error
	RemoveCategory(articleID, categoryID uint) error
	AddTag(articleID, tagID uint) error
	RemoveTag(articleID, tagID uint) error
	SyncTags(articleID uint, tagIDs []uint) error
	SyncCategories(articleID uint, categoryIDs []uint) error

	// 状态管理
	Publish(id uint) error
	Unpublish(id uint) error
	Archive(id uint) error
	SetPrivate(id uint) error
}

// ArticleListParams 文章列表查询参数
type ArticleListParams struct {
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Status   model.ArticleStatus `json:"status"`
	AuthorID uint                `json:"authorId"`
	SortBy   string              `json:"sortBy"` // created_at, updated_at, published_at, view_count
	Order    string              `json:"order"`  // asc, desc
	Search   string              `json:"search"`
}

// ArticleRepository 文章仓储实现
type ArticleRepository struct {
	db *gorm.DB
}

// NewArticleRepository 创建文章仓储实例
func NewArticleRepository(db *gorm.DB) ArticleRepositoryInterface {
	return &ArticleRepository{db: db}
}

// Create 创建文章
func (r *ArticleRepository) Create(article *model.Article) error {
	if article.Slug == "" {
		article.Slug = generateSlug(article.Title)
	}

	// 确保slug唯一
	if err := r.ensureUniqueSlug(article); err != nil {
		return err
	}

	// 设置发布时间
	if article.Status == model.ArticleStatusPublished && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	return r.db.Create(article).Error
}

// GetByID 根据ID获取文章
func (r *ArticleRepository) GetByID(id uint) (*model.Article, error) {
	var article model.Article
	err := r.db.Preload("Author").
		Preload("Category").
		Preload("Categories").
		Preload("Tags").
		First(&article, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, err
	}

	return &article, nil
}

// GetBySlug 根据Slug获取文章
func (r *ArticleRepository) GetBySlug(slug string) (*model.Article, error) {
	var article model.Article
	err := r.db.Preload("Author").
		Preload("Category").
		Preload("Categories").
		Preload("Tags").
		Where("slug = ?", slug).
		First(&article).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, err
	}

	return &article, nil
}

// Update 更新文章
func (r *ArticleRepository) Update(article *model.Article) error {
	// 检查slug唯一性
	if err := r.ensureUniqueSlug(article); err != nil {
		return err
	}

	// 如果状态变更为发布，设置发布时间
	if article.Status == model.ArticleStatusPublished && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	return r.db.Save(article).Error
}

// Delete 删除文章（软删除）
func (r *ArticleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Article{}, id).Error
}

// List 获取文章列表
func (r *ArticleRepository) List(params *ArticleListParams) ([]*model.Article, int64, error) {
	query := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags")

	// 应用筛选条件
	query = r.applyFilters(query, params)

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序
	query = r.applyPagination(query, params)
	query = r.applySorting(query, params)

	var articles []*model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetByAuthor 获取指定作者的文章
func (r *ArticleRepository) GetByAuthor(authorID uint, params *ArticleListParams) ([]*model.Article, int64, error) {
	params.AuthorID = authorID
	return r.List(params)
}

// GetByCategory 获取指定分类的文章
func (r *ArticleRepository) GetByCategory(categoryID uint, params *ArticleListParams) ([]*model.Article, int64, error) {
	query := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Where("category_id = ? OR EXISTS (SELECT 1 FROM article_categories WHERE article_categories.article_id = articles.id AND article_categories.category_id = ?)", categoryID, categoryID)

	// 应用其他筛选条件
	query = r.applyFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = r.applyPagination(query, params)
	query = r.applySorting(query, params)

	var articles []*model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetByTag 获取指定标签的文章
func (r *ArticleRepository) GetByTag(tagID uint, params *ArticleListParams) ([]*model.Article, int64, error) {
	query := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Joins("JOIN article_tags ON article_tags.article_id = articles.id").
		Where("article_tags.tag_id = ?", tagID)

	// 应用其他筛选条件
	query = r.applyFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = r.applyPagination(query, params)
	query = r.applySorting(query, params)

	var articles []*model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// Search 全文搜索文章
func (r *ArticleRepository) Search(keyword string, params *ArticleListParams) ([]*model.Article, int64, error) {
	if keyword == "" {
		return r.List(params)
	}

	searchTerm := "%" + keyword + "%"
	query := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Where("title LIKE ? OR content LIKE ? OR summary LIKE ?", searchTerm, searchTerm, searchTerm)

	// 应用其他筛选条件
	query = r.applyFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = r.applyPagination(query, params)
	query = r.applySorting(query, params)

	var articles []*model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetPopular 获取热门文章
func (r *ArticleRepository) GetPopular(limit int) ([]*model.Article, error) {
	var articles []*model.Article
	err := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Where("status = ?", model.ArticleStatusPublished).
		Order("view_count DESC, like_count DESC").
		Limit(limit).
		Find(&articles).Error

	return articles, err
}

// GetRecent 获取最新文章
func (r *ArticleRepository) GetRecent(limit int) ([]*model.Article, error) {
	var articles []*model.Article
	err := r.db.Model(&model.Article{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Where("status = ?", model.ArticleStatusPublished).
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error

	return articles, err
}

// IncrementViewCount 增加浏览量
func (r *ArticleRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// UpdateLikeCount 更新点赞数
func (r *ArticleRepository) UpdateLikeCount(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("(SELECT COUNT(*) FROM article_likes WHERE article_id = ?)", id)).Error
}

// UpdateCommentCount 更新评论数
func (r *ArticleRepository) UpdateCommentCount(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("(SELECT COUNT(*) FROM comments WHERE article_id = ? AND deleted_at IS NULL)", id)).Error
}

// AddCategory 添加分类关联
func (r *ArticleRepository) AddCategory(articleID, categoryID uint) error {
	articleCategory := &model.ArticleCategory{
		ArticleID:  articleID,
		CategoryID: categoryID,
	}
	return r.db.Create(articleCategory).Error
}

// RemoveCategory 移除分类关联
func (r *ArticleRepository) RemoveCategory(articleID, categoryID uint) error {
	return r.db.Where("article_id = ? AND category_id = ?", articleID, categoryID).
		Delete(&model.ArticleCategory{}).Error
}

// AddTag 添加标签关联
func (r *ArticleRepository) AddTag(articleID, tagID uint) error {
	articleTag := &model.ArticleTag{
		ArticleID: articleID,
		TagID:     tagID,
	}
	return r.db.Create(articleTag).Error
}

// RemoveTag 移除标签关联
func (r *ArticleRepository) RemoveTag(articleID, tagID uint) error {
	return r.db.Where("article_id = ? AND tag_id = ?", articleID, tagID).
		Delete(&model.ArticleTag{}).Error
}

// SyncTags 同步标签关联（替换所有标签）
func (r *ArticleRepository) SyncTags(articleID uint, tagIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有关联
		if err := tx.Where("article_id = ?", articleID).Delete(&model.ArticleTag{}).Error; err != nil {
			return err
		}

		// 添加新关联
		for _, tagID := range tagIDs {
			articleTag := &model.ArticleTag{
				ArticleID: articleID,
				TagID:     tagID,
			}
			if err := tx.Create(articleTag).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// SyncCategories 同步分类关联（替换所有分类）
func (r *ArticleRepository) SyncCategories(articleID uint, categoryIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除现有关联
		if err := tx.Where("article_id = ?", articleID).Delete(&model.ArticleCategory{}).Error; err != nil {
			return err
		}

		// 添加新关联
		for _, categoryID := range categoryIDs {
			articleCategory := &model.ArticleCategory{
				ArticleID:  articleID,
				CategoryID: categoryID,
			}
			if err := tx.Create(articleCategory).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Publish 发布文章
func (r *ArticleRepository) Publish(id uint) error {
	now := time.Now()
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.ArticleStatusPublished,
			"published_at": &now,
		}).Error
}

// Unpublish 取消发布文章
func (r *ArticleRepository) Unpublish(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Update("status", model.ArticleStatusDraft).Error
}

// Archive 归档文章
func (r *ArticleRepository) Archive(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Update("status", model.ArticleStatusArchived).Error
}

// SetPrivate 设置为私有文章
func (r *ArticleRepository) SetPrivate(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Update("status", model.ArticleStatusPrivate).Error
}

// 私有辅助方法

// ensureUniqueSlug 确保slug唯一
func (r *ArticleRepository) ensureUniqueSlug(article *model.Article) error {
	originalSlug := article.Slug
	counter := 1

	for {
		var count int64
		query := r.db.Model(&model.Article{}).Where("slug = ?", article.Slug)

		// 如果是更新操作，排除当前文章
		if article.ID != 0 {
			query = query.Where("id != ?", article.ID)
		}

		if err := query.Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			break
		}

		article.Slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	return nil
}

// generateSlug 从标题生成slug
func generateSlug(title string) string {
	// 简单的slug生成逻辑，实际项目中可能需要更复杂的处理
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// 移除特殊字符（保留字母、数字、连字符）
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// applyFilters 应用筛选条件
func (r *ArticleRepository) applyFilters(query *gorm.DB, params *ArticleListParams) *gorm.DB {
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if params.AuthorID != 0 {
		query = query.Where("author_id = ?", params.AuthorID)
	}

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("title LIKE ? OR content LIKE ? OR summary LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	return query
}

// applyPagination 应用分页
func (r *ArticleRepository) applyPagination(query *gorm.DB, params *ArticleListParams) *gorm.DB {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	offset := (params.Page - 1) * params.PageSize
	return query.Offset(offset).Limit(params.PageSize)
}

// applySorting 应用排序
func (r *ArticleRepository) applySorting(query *gorm.DB, params *ArticleListParams) *gorm.DB {
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.Order == "" {
		params.Order = "desc"
	}

	orderStr := params.SortBy + " " + strings.ToUpper(params.Order)
	return query.Order(orderStr)
}
