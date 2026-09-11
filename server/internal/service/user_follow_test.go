package service

import (
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeFollowRepo 用户关注仓储的测试替身。
type fakeFollowRepo struct {
	repository.UserFollowRepositoryInterface
	follows        []*model.UserFollow
	followFunc     func(followerID, followingID uint) (bool, error)
	unfollowFunc   func(followerID, followingID uint) (bool, error)
	followers      []*model.UserFollow
	following      []*model.UserFollow
	isFollowing    bool
	followerCount  int64
	followingCount int64
}

func (f *fakeFollowRepo) ListFollowers(userID uint, params *repository.FollowListParams) ([]*model.UserFollow, int64, error) {
	return f.followers, int64(len(f.followers)), nil
}

func (f *fakeFollowRepo) ListFollowing(userID uint, params *repository.FollowListParams) ([]*model.UserFollow, int64, error) {
	return f.following, int64(len(f.following)), nil
}

func (f *fakeFollowRepo) IsFollowing(followerID, followingID uint) (bool, error) {
	return f.isFollowing, nil
}

func (f *fakeFollowRepo) CountFollowers(userID uint) (int64, error) {
	return f.followerCount, nil
}

func (f *fakeFollowRepo) CountFollowing(userID uint) (int64, error) {
	return f.followingCount, nil
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
	userRepo := &fakeUserRepo{user: &domain.User{ID: 2, Role: "user", Status: 1}}
	return newTestFollowServiceWithNotification(followRepo, userRepo, &fakeNotificationRepo{})
}

// newTestFollowServiceWithNotification 创建可注入通知替身的关注服务实例。
func newTestFollowServiceWithNotification(
	followRepo repository.UserFollowRepositoryInterface,
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepositoryInterface,
) *UserFollowService {
	return NewUserFollowService(followRepo, userRepo, notificationRepo, &fakeArticleRepo{}).(*UserFollowService)
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

// TestFollowNotifiesTargetUser 验证关注成功后通知被关注用户。
func TestFollowNotifiesTargetUser(t *testing.T) {
	followRepo := &fakeFollowRepo{}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 2, Role: "user", Status: 1}}
	notificationRepo := &fakeNotificationRepo{}
	svc := newTestFollowServiceWithNotification(followRepo, userRepo, notificationRepo)

	if err := svc.Follow(1, 2); err != nil {
		t.Fatalf("关注失败: %v", err)
	}
	if len(notificationRepo.created) != 1 {
		t.Fatalf("通知数 = %d, 期望 1", len(notificationRepo.created))
	}
	notification := notificationRepo.created[0]
	if notification.UserID != 2 {
		t.Errorf("通知接收者 = %d, 期望 2", notification.UserID)
	}
	if notification.Type != model.NotificationTypeFollow {
		t.Errorf("通知类型 = %s, 期望 follow", notification.Type)
	}
	if notification.SenderID == nil || *notification.SenderID != 1 {
		t.Errorf("通知触发者 = %v, 期望 1", notification.SenderID)
	}
}

// TestFollowSelfSkipsNotification 验证关注被拒时不产生通知。
func TestFollowSelfSkipsNotification(t *testing.T) {
	followRepo := &fakeFollowRepo{}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "user", Status: 1}}
	notificationRepo := &fakeNotificationRepo{}
	svc := newTestFollowServiceWithNotification(followRepo, userRepo, notificationRepo)

	if err := svc.Follow(1, 1); err == nil {
		t.Fatal("关注自己应返回错误")
	}
	if len(notificationRepo.created) != 0 {
		t.Errorf("通知数 = %d, 期望 0", len(notificationRepo.created))
	}
}

// TestListFollowersIncludesUserSummary 验证粉丝列表附带关注者用户摘要。
func TestListFollowersIncludesUserSummary(t *testing.T) {
	followerUser := model.User{ID: 1, Username: "user1", Nickname: "昵称一", Avatar: "a.png"}
	followRepo := &fakeFollowRepo{
		followers: []*model.UserFollow{
			{ID: 9, FollowerID: 1, FollowingID: 2, Follower: followerUser},
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 2, Role: "user", Status: 1}}
	svc := newTestFollowServiceWithNotification(followRepo, userRepo, &fakeNotificationRepo{})

	result, err := svc.ListFollowers(2, &ListFollowsRequest{})
	if err != nil {
		t.Fatalf("查询粉丝列表失败: %v", err)
	}
	if len(result.Follows) != 1 {
		t.Fatalf("粉丝条目数 = %d, 期望 1", len(result.Follows))
	}
	summary := result.Follows[0].User
	if summary == nil || summary.Nickname != "昵称一" || summary.ID != 1 {
		t.Errorf("用户摘要 = %+v, 期望昵称一的摘要", summary)
	}
}

// TestIsFollowingPassThrough 验证关注状态查询透传仓储结果。
func TestIsFollowingPassThrough(t *testing.T) {
	followRepo := &fakeFollowRepo{isFollowing: true}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 2, Role: "user", Status: 1}}
	svc := newTestFollowServiceWithNotification(followRepo, userRepo, &fakeNotificationRepo{})

	isFollowing, err := svc.IsFollowing(1, 2)
	if err != nil {
		t.Fatalf("查询关注状态失败: %v", err)
	}
	if !isFollowing {
		t.Error("关注状态应为 true")
	}
}

// TestGetPublicProfileAggregates 验证公开资料聚合用户信息与公开统计。
func TestGetPublicProfileAggregates(t *testing.T) {
	followRepo := &fakeFollowRepo{followerCount: 3, followingCount: 1}
	userRepo := &fakeUserRepo{user: &domain.User{
		ID: 2, Username: "user2", Nickname: "昵称二",
		Avatar: "avatar.png", Bio: "简介内容", Website: "https://example.com", Role: "user", Status: 1,
	}}
	articleRepo := &fakeArticleRepo{
		getByAuthor: func(authorID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			return []*model.Article{publishedArticle(1)}, 4, nil
		},
	}
	notificationRepo := &fakeNotificationRepo{}
	svc := NewUserFollowService(followRepo, userRepo, notificationRepo, articleRepo).(*UserFollowService)

	profile, err := svc.GetPublicProfile(2)
	if err != nil {
		t.Fatalf("获取公开资料失败: %v", err)
	}
	if profile.Nickname != "昵称二" || profile.FollowerCount != 3 || profile.ArticleCount != 4 {
		t.Errorf("公开资料 = %+v, 期望昵称二、3 粉丝、4 篇文章", profile)
	}
}
