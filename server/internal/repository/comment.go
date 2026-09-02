// Package repository 数据访问层
package repository

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"

	"gorm.io/gorm"
)

// ErrCommentNotFound 评论不存在的哨兵错误，供 service 与 handler 层识别业务错误。
var ErrCommentNotFound = errors.New("评论不存在")

// CommentRepositoryInterface 评论仓储接口
type CommentRepositoryInterface interface {
	// 基础CRUD操作
	Create(comment *model.Comment) error
	GetByID(id uint) (*model.Comment, error)
	Update(comment *model.Comment) error
	Delete(id uint) error
	UpdateStatus(id uint, status model.CommentStatus) error

	// 查询操作
	ListByArticle(articleID uint, params *CommentListParams) ([]*model.Comment, int64, error)
	ListByUser(userID uint, params *CommentListParams) ([]*model.Comment, int64, error)
	ListAdmin(params *CommentListParams) ([]*model.Comment, int64, error)
	GetByParentID(parentID uint) ([]*model.Comment, error)

	// 点赞操作
	AddLike(commentID, userID uint) (bool, error)
	RemoveLike(commentID, userID uint) (bool, error)

	// 计数维护
	IncrementReplyCount(id uint) error
	DecrementReplyCount(id uint) error
}

// CommentListParams 评论列表查询参数
type CommentListParams struct {
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Status   model.CommentStatus `json:"status"`
	Keyword  string              `json:"keyword"`
}

// CommentRepository 评论仓储实现
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 创建评论仓储实例
func NewCommentRepository(db *gorm.DB) CommentRepositoryInterface {
	return &CommentRepository{db: db}
}

// Create 创建评论
func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

// GetByID 根据ID获取评论
func (r *CommentRepository) GetByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	if err := r.db.Preload("User").First(&comment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}
	return &comment, nil
}

// Update 更新评论
func (r *CommentRepository) Update(comment *model.Comment) error {
	return r.db.Save(comment).Error
}

// Delete 删除评论，采用软删除。
func (r *CommentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Comment{}, id).Error
}

// UpdateStatus 更新评论审核状态。
func (r *CommentRepository) UpdateStatus(id uint, status model.CommentStatus) error {
	if err := r.db.Model(&model.Comment{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("更新评论状态失败: %w", err)
	}
	return nil
}

// ListByArticle 分页查询文章评论，仅返回已审核通过的评论。
func (r *CommentRepository) ListByArticle(articleID uint, params *CommentListParams) ([]*model.Comment, int64, error) {
	query := r.db.Model(&model.Comment{}).
		Where("article_id = ?", articleID).
		Where("status = ?", model.CommentStatusApproved)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询评论总数失败: %w", err)
	}

	// 设置分页默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var comments []*model.Comment
	offset := (params.Page - 1) * params.PageSize
	if err := query.Preload("User").
		Order("is_pinned DESC, created_at ASC").
		Offset(offset).Limit(params.PageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, fmt.Errorf("查询评论列表失败: %w", err)
	}

	return comments, total, nil
}

// ListByUser 分页查询用户发表的评论。
func (r *CommentRepository) ListByUser(userID uint, params *CommentListParams) ([]*model.Comment, int64, error) {
	query := r.db.Model(&model.Comment{}).Where("user_id = ?", userID)

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询评论总数失败: %w", err)
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var comments []*model.Comment
	offset := (params.Page - 1) * params.PageSize
	if err := query.Preload("Article").
		Order("created_at DESC").
		Offset(offset).Limit(params.PageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, fmt.Errorf("查询评论列表失败: %w", err)
	}

	return comments, total, nil
}

// ListAdmin 分页查询全量评论，用于管理端审核。
func (r *CommentRepository) ListAdmin(params *CommentListParams) ([]*model.Comment, int64, error) {
	query := r.db.Model(&model.Comment{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Keyword != "" {
		searchTerm := "%" + params.Keyword + "%"
		query = query.Where("content LIKE ? OR author_name LIKE ?", searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询评论总数失败: %w", err)
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	var comments []*model.Comment
	offset := (params.Page - 1) * params.PageSize
	if err := query.Preload("User").Preload("Article").
		Order("created_at DESC").
		Offset(offset).Limit(params.PageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, fmt.Errorf("查询评论列表失败: %w", err)
	}

	return comments, total, nil
}

// GetByParentID 查询指定父评论下的直接回复。
func (r *CommentRepository) GetByParentID(parentID uint) ([]*model.Comment, error) {
	var comments []*model.Comment
	if err := r.db.Where("parent_id = ? AND status = ?", parentID, model.CommentStatusApproved).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("查询评论回复失败: %w", err)
	}
	return comments, nil
}

// AddLike 添加评论点赞记录，依赖唯一索引防重复，返回是否新增。
func (r *CommentRepository) AddLike(commentID, userID uint) (bool, error) {
	var added bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec("INSERT IGNORE INTO comment_likes (comment_id, user_id, created_at) VALUES (?, ?, NOW())", commentID, userID)
		if result.Error != nil {
			return result.Error
		}
		added = result.RowsAffected > 0
		if added {
			return tx.Model(&model.Comment{}).
				Where("id = ?", commentID).
				UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
		}
		return nil
	})
	return added, err
}

// RemoveLike 移除评论点赞记录，返回是否实际删除。
func (r *CommentRepository) RemoveLike(commentID, userID uint) (bool, error) {
	var removed bool
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec("DELETE FROM comment_likes WHERE comment_id = ? AND user_id = ?", commentID, userID)
		if result.Error != nil {
			return result.Error
		}
		removed = result.RowsAffected > 0
		if removed {
			return tx.Model(&model.Comment{}).
				Where("id = ?", commentID).
				UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).Error
		}
		return nil
	})
	return removed, err
}

// IncrementReplyCount 递增评论回复数。
func (r *CommentRepository) IncrementReplyCount(id uint) error {
	return r.db.Model(&model.Comment{}).
		Where("id = ?", id).
		UpdateColumn("reply_count", gorm.Expr("reply_count + 1")).Error
}

// DecrementReplyCount 递减评论回复数，回复数不小于零。
func (r *CommentRepository) DecrementReplyCount(id uint) error {
	return r.db.Model(&model.Comment{}).
		Where("id = ?", id).
		UpdateColumn("reply_count", gorm.Expr("CASE WHEN reply_count > 0 THEN reply_count - 1 ELSE 0 END")).Error
}
