// Package service 业务逻辑层
package service

import (
	"errors"

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
}

// ListFollowsRequest 关注列表请求
type ListFollowsRequest struct {
	Page     int `json:"page" binding:"omitempty,min=1"`
	PageSize int `json:"pageSize" binding:"omitempty,min=1,max=100"`
}

// FollowListResponse 关注列表响应
type FollowListResponse struct {
	Follows  []*model.UserFollow `json:"follows"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

// UserFollowService 用户关注服务实现
type UserFollowService struct {
	followRepo repository.UserFollowRepositoryInterface
	userRepo   repository.UserRepository
}

// NewUserFollowService 创建用户关注服务实例
func NewUserFollowService(
	followRepo repository.UserFollowRepositoryInterface,
	userRepo repository.UserRepository,
) UserFollowServiceInterface {
	return &UserFollowService{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

// Follow 关注目标用户，禁止自我关注。
func (s *UserFollowService) Follow(followerID, followingID uint) error {
	// 防止用户关注自己。
	if followerID == followingID {
		return errors.New("不能关注自己")
	}

	// 校验目标用户存在。
	if _, err := s.userRepo.GetByID(followingID); err != nil {
		return errors.New("目标用户不存在")
	}

	_, err := s.followRepo.Follow(followerID, followingID)
	return err
}

// Unfollow 取消关注目标用户。
func (s *UserFollowService) Unfollow(followerID, followingID uint) error {
	_, err := s.followRepo.Unfollow(followerID, followingID)
	return err
}

// ListFollowers 分页查询粉丝列表。
func (s *UserFollowService) ListFollowers(userID uint, req *ListFollowsRequest) (*FollowListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	params := &repository.FollowListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	follows, total, err := s.followRepo.ListFollowers(userID, params)
	if err != nil {
		return nil, err
	}

	return &FollowListResponse{
		Follows:  follows,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// ListFollowing 分页查询关注列表。
func (s *UserFollowService) ListFollowing(userID uint, req *ListFollowsRequest) (*FollowListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	params := &repository.FollowListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	follows, total, err := s.followRepo.ListFollowing(userID, params)
	if err != nil {
		return nil, err
	}

	return &FollowListResponse{
		Follows:  follows,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
