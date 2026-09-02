package service

import (
	"testing"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeNotificationRepo 通知仓储的测试替身。
type fakeNotificationRepo struct {
	repository.NotificationRepositoryInterface
	notifications []*model.Notification
	unread        int64
	markAsRead    func(id uint, userID uint) error
}

func (f *fakeNotificationRepo) ListByUser(userID uint, params *repository.NotificationListParams) ([]*model.Notification, int64, error) {
	return f.notifications, int64(len(f.notifications)), nil
}

func (f *fakeNotificationRepo) CountUnread(userID uint) (int64, error) {
	return f.unread, nil
}

func (f *fakeNotificationRepo) MarkAsRead(id uint, userID uint) error {
	if f.markAsRead != nil {
		return f.markAsRead(id, userID)
	}
	return nil
}

func (f *fakeNotificationRepo) MarkAllAsRead(userID uint) error {
	return nil
}

// TestListNotificationsIncludesUnread 验证通知列表响应附带未读数。
func TestListNotificationsIncludesUnread(t *testing.T) {
	repo := &fakeNotificationRepo{
		notifications: []*model.Notification{
			{ID: 1, UserID: 1, Type: model.NotificationTypeSystem, Title: "系统通知", IsRead: false},
		},
		unread: 1,
	}
	svc := NewNotificationService(repo)

	req := &ListNotificationsRequest{Page: 1, PageSize: 10}
	result, err := svc.ListNotifications(1, req)
	if err != nil {
		t.Fatalf("查询通知列表失败: %v", err)
	}

	if len(result.Notifications) != 1 {
		t.Errorf("通知条数 = %d, 期望 1", len(result.Notifications))
	}
	if result.UnreadCount != 1 {
		t.Errorf("未读数 = %d, 期望 1", result.UnreadCount)
	}
}

// TestGetUnreadCount 验证未读数统计。
func TestGetUnreadCount(t *testing.T) {
	repo := &fakeNotificationRepo{unread: 3}
	svc := NewNotificationService(repo)

	count, err := svc.GetUnreadCount(1)
	if err != nil {
		t.Fatalf("查询未读数失败: %v", err)
	}

	if count != 3 {
		t.Errorf("未读数 = %d, 期望 3", count)
	}
}
