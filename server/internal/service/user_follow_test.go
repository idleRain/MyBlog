package service

import (
	"testing"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeFollowRepo 用户关注仓储的测试替身。
type fakeFollowRepo struct {
	repository.UserFollowRepositoryInterface
	follows    []*model.UserFollow
	followFunc func(followerID, followingID uint) (bool, error)
	unfollowFunc func(followerID, followingID uint) (bool, error)
}

func (f *fakeFollowRepo) Follow(followerID, followingID uint) (bool, error) {
	if f.followFunc != nil {
		return f.followFunc(followerID, followingID)
	}
	f.follows = append(f.follows, &model.UserFollow{FollowerID: followerID, FollowingID: followingID})
	return true, nil
}

func (f *fakeFollowRepo) Unfollow(followerID, followingID uint) (bool, error) {
	if f.unfollowFunc != nil {
		return f.unfollowFunc(followerID, followingID)
	}
	for i, follow := range f.follows {
		if follow.FollowerID == followerID && follow.FollowingID == followingID {
			f.follows = append(f.follows[:i], f.follows[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// newTestFollowService 创建注入测试替身的关注服务实例。
func newTestFollowService(followRepo repository.UserFollowRepositoryInterface) *UserFollowService {
	userRepo := &fakeUserRepo{user: &repository.User{ID: 2, Role: "user", Status: 1}}
	return NewUserFollowService(followRepo, userRepo).(*UserFollowService)
}

// TestFollowSelfRejected 验证禁止关注自己。
func TestFollowSelfRejected(t *testing.T) {
	svc := newTestFollowService(&fakeFollowRepo{})

	err := svc.Follow(1, 1)
	if err == nil {
		t.Fatal("关注自己应返回错误")
	}
}

// TestFollowTargetNotFound 验证目标用户不存在时报错。
func TestFollowTargetNotFound(t *testing.T) {
	svc := newTestFollowService(&fakeFollowRepo{})

	err := svc.Follow(1, 999)
	if err == nil {
		t.Fatal("关注不存在的用户应返回错误")
	}
}

// TestFollowSuccess 验证正常关注调用仓储。
func TestFollowSuccess(t *testing.T) {
	calledFollow := false
	repo := &fakeFollowRepo{
		followFunc: func(followerID, followingID uint) (bool, error) {
			calledFollow = true
			return true, nil
		},
	}
	svc := newTestFollowService(repo)

	if err := svc.Follow(1, 2); err != nil {
		t.Fatalf("关注失败: %v", err)
	}
	if !calledFollow {
		t.Error("关注应调用仓储 Follow")
	}
}

// TestUnfollowSuccess 验证取消关注调用仓储。
func TestUnfollowSuccess(t *testing.T) {
	calledUnfollow := false
	repo := &fakeFollowRepo{
		unfollowFunc: func(followerID, followingID uint) (bool, error) {
			calledUnfollow = true
			return true, nil
		},
	}
	svc := newTestFollowService(repo)

	if err := svc.Unfollow(1, 2); err != nil {
		t.Fatalf("取消关注失败: %v", err)
	}
	if !calledUnfollow {
		t.Error("取消关注应调用仓储 Unfollow")
	}
}
