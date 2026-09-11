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
	CreateComment(req *CreateCommentRequest, userID *uint) (*model.Comment, error)
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
	commentRepo      repository.CommentRepositoryInterface
	articleRepo      repository.ArticleRepositoryInterface
	settingRepo      repository.SettingRepositoryInterface
	notificationRepo repository.NotificationRepositoryInterface
}

// NewCommentService 创建评论服务实例
func NewCommentService(
	commentRepo repository.CommentRepositoryInterface,
	articleRepo repository.ArticleRepositoryInterface,
	settingRepo repository.SettingRepositoryInterface,
	notificationRepo repository.NotificationRepositoryInterface,
) CommentServiceInterface {
	return &CommentService{
		commentRepo:      commentRepo,
		articleRepo:      articleRepo,
		settingRepo:      settingRepo,
		notificationRepo: notificationRepo,
	}
}

// settingEnabled 读取布尔型站点设置，设置缺失时回退调用方给定的默认值。
func (s *CommentService) settingEnabled(key string, defaultValue bool) bool {
	setting, err := s.settingRepo.GetByKey(key)
	if err != nil {
		return defaultValue
	}
	return setting.GetBoolValue()
}

// CreateComment 创建评论，支持注册用户与游客双通道，评论默认待审核。
func (s *CommentService) CreateComment(req *CreateCommentRequest, userID *uint) (*model.Comment, error) {
	// 校验文章存在且允许评论。
	article, err := s.articleRepo.GetByID(req.ArticleID)
	if err != nil {
		return nil, err
	}
	if !article.CanComment() {
		return nil, errors.New("该文章不允许评论")
	}

	// 游客通道受站点开关控制，登录通道不受该开关限制。
	if userID == nil && !s.settingEnabled(model.SettingAllowGuestComment, true) {
		return nil, errors.New("站点已关闭游客评论，请登录后发言")
	}

	comment := &model.Comment{
		ArticleID:     req.ArticleID,
		Content:       req.Content,
		Status:        model.CommentStatusPending,
		AuthorName:    req.AuthorName,
		AuthorEmail:   req.AuthorEmail,
		AuthorWebsite: req.AuthorWebsite,
	}

	// 站点开启评论自动通过时，评论跳过待审核直接进入已审核状态。
	if s.settingEnabled(model.SettingCommentAutoApprove, false) {
		comment.Status = model.CommentStatusApproved
	}

	// 登录评论绑定账号身份，展示名经关联用户解析，游客字段仅游客通道生效。
	if userID != nil {
		comment.UserID = userID
		comment.AuthorName = ""
		comment.AuthorEmail = ""
		comment.AuthorWebsite = ""
	}

	// 游客提交时必须提供姓名。
	if comment.UserID == nil && comment.AuthorName == "" {
		return nil, errors.New("游客评论必须填写姓名")
	}

	// 处理回复关系，回复时继承父评论的文章归属。
	var parent *model.Comment
	if req.ParentID != nil {
		var err error
		parent, err = s.commentRepo.GetByID(*req.ParentID)
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

	s.notifyParentAuthor(parent, comment, article.Slug, userID)

	return s.commentRepo.GetByID(comment.ID)
}

// notifyParentAuthor 在回复目标为注册用户时写入评论回复通知。
// 通知为副产物，写入失败不阻断评论创建；游客作者与自回复不产生通知。
func (s *CommentService) notifyParentAuthor(parent *model.Comment, reply *model.Comment, articleSlug string, replierID *uint) {
	if parent == nil || parent.UserID == nil {
		return
	}
	if replierID != nil && *parent.UserID == *replierID {
		return
	}

	relatedType := model.RelatedTypeComment
	notification := &model.Notification{
		UserID:      *parent.UserID,
		SenderID:    replierID,
		Type:        model.NotificationTypeCommentReply,
		Title:       "你的评论收到了新回复",
		ActionURL:   fmt.Sprintf("/blog/%s#comment-%d", articleSlug, reply.ID),
		RelatedType: &relatedType,
		RelatedID:   &reply.ID,
	}
	_ = s.notificationRepo.Create(notification)
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

// LikeComment 点赞评论，首次点赞时通知评论作者。
func (s *CommentService) LikeComment(commentID, userID uint) error {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return err
	}

	created, err := s.commentRepo.AddLike(commentID, userID)
	if err != nil {
		return err
	}

	// 首次点赞才通知，重复点赞保持幂等。
	if created {
		s.notifyCommentLike(comment, userID)
	}
	return nil
}

// notifyCommentLike 在评论作者为注册用户且非点赞者本人时写入点赞通知。
// 通知为副产物，写入失败不阻断点赞。
func (s *CommentService) notifyCommentLike(comment *model.Comment, likerID uint) {
	if comment.UserID == nil || *comment.UserID == likerID {
		return
	}

	article, err := s.articleRepo.GetByID(comment.ArticleID)
	if err != nil {
		return
	}

	relatedType := model.RelatedTypeComment
	notification := &model.Notification{
		UserID:      *comment.UserID,
		SenderID:    &likerID,
		Type:        model.NotificationTypeCommentLike,
		Title:       "你的评论收到了新的点赞",
		ActionURL:   fmt.Sprintf("/blog/%s#comment-%d", article.Slug, comment.ID),
		RelatedType: &relatedType,
		RelatedID:   &comment.ID,
	}
	_ = s.notificationRepo.Create(notification)
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
