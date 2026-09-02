// Package service 业务逻辑层
package service

import (
	"errors"
	"fmt"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// CommentServiceInterface 评论服务接口
type CommentServiceInterface interface {
	// 评论操作
	CreateComment(req *CreateCommentRequest) (*model.Comment, error)
	GetCommentsByArticle(articleID uint, req *ListCommentsRequest) (*CommentListResponse, error)
	LikeComment(commentID, userID uint) error
	UnlikeComment(commentID, userID uint) error

	// 审核操作
	ApproveComment(id uint, operatorID uint) error
	RejectComment(id uint, operatorID uint) error
	MarkCommentSpam(id uint, operatorID uint) error
	TrashComment(id uint, operatorID uint) error
	DeleteComment(id uint, operatorID uint) error
	ListComments(req *AdminListCommentsRequest) (*CommentListResponse, error)
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	ArticleID     uint   `json:"articleId" binding:"required"`
	ParentID      *uint  `json:"parentId"`
	Content       string `json:"content" binding:"required,min=1,max=2000"`
	AuthorName    string `json:"authorName" binding:"omitempty,max=50"`
	AuthorEmail   string `json:"authorEmail" binding:"omitempty,email,max=100"`
	AuthorWebsite string `json:"authorWebsite" binding:"omitempty,max=255"`
}

// ListCommentsRequest 文章评论列表请求
type ListCommentsRequest struct {
	Page     int `json:"page" binding:"omitempty,min=1"`
	PageSize int `json:"pageSize" binding:"omitempty,min=1,max=100"`
}

// AdminListCommentsRequest 管理端评论列表请求
type AdminListCommentsRequest struct {
	Page     int    `json:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"pageSize" binding:"omitempty,min=1,max=100"`
	Status   string `json:"status" binding:"omitempty,oneof=pending approved rejected spam trash"`
	Keyword  string `json:"keyword"`
}

// CommentListResponse 评论列表响应
type CommentListResponse struct {
	Comments []*model.Comment `json:"comments"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// CommentService 评论服务实现
type CommentService struct {
	commentRepo repository.CommentRepositoryInterface
	articleRepo repository.ArticleRepositoryInterface
}

// NewCommentService 创建评论服务实例
func NewCommentService(
	commentRepo repository.CommentRepositoryInterface,
	articleRepo repository.ArticleRepositoryInterface,
) CommentServiceInterface {
	return &CommentService{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
	}
}

// CreateComment 创建评论，支持注册用户与游客双通道，评论默认待审核。
func (s *CommentService) CreateComment(req *CreateCommentRequest) (*model.Comment, error) {
	// 校验文章存在且允许评论。
	article, err := s.articleRepo.GetByID(req.ArticleID)
	if err != nil {
		return nil, err
	}
	if !article.CanComment() {
		return nil, errors.New("该文章不允许评论")
	}

	comment := &model.Comment{
		ArticleID:  req.ArticleID,
		Content:    req.Content,
		Status:     model.CommentStatusPending,
		AuthorName: req.AuthorName,
	}

	// 游客提交时必须提供姓名。
	if comment.AuthorName == "" {
		return nil, errors.New("游客评论必须填写姓名")
	}

	// 处理回复关系，回复时继承父评论的文章归属。
	if req.ParentID != nil {
		parent, err := s.commentRepo.GetByID(*req.ParentID)
		if err != nil {
			return nil, errors.New("父评论不存在")
		}
		if parent.ArticleID != req.ArticleID {
			return nil, errors.New("父评论不属于该文章")
		}

		comment.ParentID = req.ParentID
		// 两级树：根评论回复为二级，二级回复统一挂到根评论下。
		if parent.Level == 1 {
			comment.RootID = &parent.ID
			comment.Level = 2
		} else {
			comment.RootID = parent.RootID
			comment.Level = 2
		}
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}

	// 回复评论时递增父评论的回复计数。
	if comment.ParentID != nil {
		if err := s.commentRepo.IncrementReplyCount(*comment.ParentID); err != nil {
			return nil, err
		}
	}

	return s.commentRepo.GetByID(comment.ID)
}

// GetCommentsByArticle 获取文章评论列表。
func (s *CommentService) GetCommentsByArticle(articleID uint, req *ListCommentsRequest) (*CommentListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	params := &repository.CommentListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	comments, total, err := s.commentRepo.ListByArticle(articleID, params)
	if err != nil {
		return nil, err
	}

	return &CommentListResponse{
		Comments: comments,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// LikeComment 点赞评论。
func (s *CommentService) LikeComment(commentID, userID uint) error {
	if _, err := s.commentRepo.GetByID(commentID); err != nil {
		return err
	}
	_, err := s.commentRepo.AddLike(commentID, userID)
	return err
}

// UnlikeComment 取消点赞评论。
func (s *CommentService) UnlikeComment(commentID, userID uint) error {
	_, err := s.commentRepo.RemoveLike(commentID, userID)
	return err
}

// ApproveComment 审核通过评论。
func (s *CommentService) ApproveComment(id uint, operatorID uint) error {
	return s.moderateComment(id, model.CommentStatusApproved)
}

// RejectComment 拒绝评论。
func (s *CommentService) RejectComment(id uint, operatorID uint) error {
	return s.moderateComment(id, model.CommentStatusRejected)
}

// MarkCommentSpam 标记为垃圾评论。
func (s *CommentService) MarkCommentSpam(id uint, operatorID uint) error {
	return s.moderateComment(id, model.CommentStatusSpam)
}

// TrashComment 移入回收站。
func (s *CommentService) TrashComment(id uint, operatorID uint) error {
	return s.moderateComment(id, model.CommentStatusTrash)
}

// DeleteComment 删除评论，采用软删除。
func (s *CommentService) DeleteComment(id uint, operatorID uint) error {
	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.commentRepo.Delete(id); err != nil {
		return err
	}

	// 删除回复评论时回退父评论的回复计数。
	if comment.ParentID != nil {
		return s.commentRepo.DecrementReplyCount(*comment.ParentID)
	}
	return nil
}

// ListComments 管理端分页查询评论。
func (s *CommentService) ListComments(req *AdminListCommentsRequest) (*CommentListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	params := &repository.CommentListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   model.CommentStatus(req.Status),
		Keyword:  req.Keyword,
	}

	comments, total, err := s.commentRepo.ListAdmin(params)
	if err != nil {
		return nil, err
	}

	return &CommentListResponse{
		Comments: comments,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// moderateComment 执行评论状态流转，复用统一的审核入口。
func (s *CommentService) moderateComment(id uint, status model.CommentStatus) error {
	if _, err := s.commentRepo.GetByID(id); err != nil {
		return err
	}
	return s.commentRepo.UpdateStatus(id, status)
}
