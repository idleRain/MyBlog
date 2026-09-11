// Package service 业务逻辑层
package service

import (
	"errors"
	"time"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// UserFollowServiceInterface 用户关注服务接口
type UserFollowServiceInterface interface {
	// 关注操作
	Follow(followerID, followingID uint) error
	Unfollow(followerID, followingID uint) error

	// 查询操作
	ListFollowers(userID uint, req *ListFollowsRequest) (*FollowListResponse, error)
	ListFollowing(userID uint, req *ListFollowsRequest) (*FollowListResponse, error)
	IsFollowing(followerID, followingID uint) (bool, error)
	GetPublicProfile(userID uint) (*PublicUserProfile, error)
}

// ListFollowsRequest 关注列表请求
type ListFollowsRequest struct {
	Page     int `json:"page" binding:"omitempty,min=1"`
	PageSize int `json:"pageSize" binding:"omitempty,min=1,max=100"`
}

// FollowUserSummary 关注关系中的用户摘要。
type FollowUserSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// FollowItem 关注关系条目，user 为对方用户的摘要：粉丝列表中为关注者，关注列表中为被关注者。
type FollowItem struct {
	ID          uint               `json:"id"`
	FollowerID  uint               `json:"followerId"`
	FollowingID uint               `json:"followingId"`
	CreatedAt   time.Time          `json:"createdAt"`
	User        *FollowUserSummary `json:"user"`
}

// FollowListResponse 关注列表响应
type FollowListResponse struct {
	Follows  []*FollowItem `json:"follows"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

// PublicUserProfile 用户公开资料，仅包含可对外展示的字段与公开统计。
type PublicUserProfile struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	Nickname       string `json:"nickname"`
	Avatar         string `json:"avatar"`
	Bio            string `json:"bio"`
	Website        string `json:"website"`
	FollowerCount  int64  `json:"followerCount"`
	FollowingCount int64  `json:"followingCount"`
	ArticleCount   int64  `json:"articleCount"`
}

// UserFollowService 用户关注服务实现
type UserFollowService struct {
	followRepo       repository.UserFollowRepositoryInterface
	userRepo         repository.UserRepository
	notificationRepo repository.NotificationRepositoryInterface
	articleRepo      repository.ArticleRepositoryInterface
}

// NewUserFollowService 创建用户关注服务实例
func NewUserFollowService(
	followRepo repository.UserFollowRepositoryInterface,
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepositoryInterface,
	articleRepo repository.ArticleRepositoryInterface,
) UserFollowServiceInterface {
	return &UserFollowService{
		followRepo:       followRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		articleRepo:      articleRepo,
	}
}

// Follow 关注目标用户，禁止自我关注，成功后通知被关注用户。
func (s *UserFollowService) Follow(followerID, followingID uint) error {
	// 防止用户关注自己。
	if followerID == followingID {
		return errors.New("不能关注自己")
	}

	// 校验目标用户存在。
	if _, err := s.userRepo.GetByID(followingID); err != nil {
		return errors.New("目标用户不存在")
	}

	if _, err := s.followRepo.Follow(followerID, followingID); err != nil {
		return err
	}

	// 关注为副产通知，写入失败不阻断关注操作。
	relatedType := model.RelatedTypeUser
	notification := &model.Notification{
		UserID:      followingID,
		SenderID:    &followerID,
		Type:        model.NotificationTypeFollow,
		Title:       "有新用户关注了你",
		RelatedType: &relatedType,
		RelatedID:   &followerID,
	}
	_ = s.notificationRepo.Create(notification)
	return nil
}

// Unfollow 取消关注目标用户。
func (s *UserFollowService) Unfollow(followerID, followingID uint) error {
	_, err := s.followRepo.Unfollow(followerID, followingID)
	return err
}

// ListFollowers 分页查询粉丝列表，附带关注者用户摘要。
func (s *UserFollowService) ListFollowers(userID uint, req *ListFollowsRequest) (*FollowListResponse, error) {
	params := s.applyFollowPagination(req)

	follows, total, err := s.followRepo.ListFollowers(userID, params)
	if err != nil {
		return nil, err
	}

	items := make([]*FollowItem, 0, len(follows))
	for _, follow := range follows {
		// 粉丝列表中的对方用户为关注者。
		items = append(items, newFollowItem(follow, follow.Follower))
	}

	return &FollowListResponse{
		Follows:  items,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// ListFollowing 分页查询关注列表，附带被关注用户摘要。
func (s *UserFollowService) ListFollowing(userID uint, req *ListFollowsRequest) (*FollowListResponse, error) {
	params := s.applyFollowPagination(req)

	follows, total, err := s.followRepo.ListFollowing(userID, params)
	if err != nil {
		return nil, err
	}

	items := make([]*FollowItem, 0, len(follows))
	for _, follow := range follows {
		// 关注列表中的对方用户为被关注者。
		items = append(items, newFollowItem(follow, follow.Following))
	}

	return &FollowListResponse{
		Follows:  items,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// IsFollowing 查询当前用户是否已关注目标用户。
func (s *UserFollowService) IsFollowing(followerID, followingID uint) (bool, error) {
	return s.followRepo.IsFollowing(followerID, followingID)
}

// GetPublicProfile 获取用户公开资料，仅返回可对外展示的字段与公开统计。
func (s *UserFollowService) GetPublicProfile(userID uint) (*PublicUserProfile, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	followerCount, err := s.followRepo.CountFollowers(userID)
	if err != nil {
		return nil, err
	}
	followingCount, err := s.followRepo.CountFollowing(userID)
	if err != nil {
		return nil, err
	}

	// 文章数取已发布文章的分页总数，避免全量加载。
	_, articleCount, err := s.articleRepo.GetByAuthor(userID, &repository.ArticleListParams{
		Page:     1,
		PageSize: 1,
		Status:   model.ArticleStatusPublished,
	})
	if err != nil {
		return nil, err
	}

	return &PublicUserProfile{
		ID:             user.ID,
		Username:       user.Username,
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		Bio:            user.Bio,
		Website:        user.Website,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		ArticleCount:   articleCount,
	}, nil
}

// applyFollowPagination 归一关注列表的分页参数。
func (s *UserFollowService) applyFollowPagination(req *ListFollowsRequest) *repository.FollowListParams {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	return &repository.FollowListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	}
}

// newFollowItem 由关注关系与对方用户构建条目，用户缺失时摘要仅含零值标识。
func newFollowItem(follow *model.UserFollow, user model.User) *FollowItem {
	return &FollowItem{
		ID:          follow.ID,
		FollowerID:  follow.FollowerID,
		FollowingID: follow.FollowingID,
		CreatedAt:   follow.CreatedAt,
		User: &FollowUserSummary{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		},
	}
}
