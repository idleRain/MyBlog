package service

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"MyBlog/internal/domain"
)

// newTestUserService 创建注入测试替身的用户服务实例，jwtService 在自助资料链路中不被触碰。
func newTestUserService(userRepo *fakeUserRepo) *userService {
	return NewUserService(userRepo, nil, NewRBACService()).(*userService)
}

// TestUpdateProfileAppliesOptionalFields 验证自助资料仅更新显式传入的字段。
func TestUpdateProfileAppliesOptionalFields(t *testing.T) {
	userRepo := &fakeUserRepo{user: &domain.User{
		ID: 1, Username: "user1", Nickname: "旧昵称", Bio: "旧简介", Role: "user", Status: 1,
	}}
	svc := newTestUserService(userRepo)

	newNickname := "新昵称"
	newWebsite := "https://example.com"
	updated, err := svc.UpdateProfile(1, &UpdateProfileRequest{
		Nickname: &newNickname,
		Website:  &newWebsite,
	})
	if err != nil {
		t.Fatalf("更新资料失败: %v", err)
	}

	// 显式传入的字段被更新。
	if updated.Nickname != "新昵称" || updated.Website != newWebsite {
		t.Errorf("资料 = %q %q, 期望新昵称与新网站", updated.Nickname, updated.Website)
	}
	// 未传入的字段保留原值。
	if updated.Bio != "旧简介" {
		t.Errorf("未传入的简介应保留原值, 实际为 %q", updated.Bio)
	}
}

// TestUpdateProfileEmptyNicknameFallsBack 验证昵称被清空时回退为用户名。
func TestUpdateProfileEmptyNicknameFallsBack(t *testing.T) {
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Username: "user1", Nickname: "旧昵称", Role: "user", Status: 1}}
	svc := newTestUserService(userRepo)

	emptyNickname := ""
	updated, err := svc.UpdateProfile(1, &UpdateProfileRequest{Nickname: &emptyNickname})
	if err != nil {
		t.Fatalf("更新资料失败: %v", err)
	}
	if updated.Nickname != "user1" {
		t.Errorf("昵称 = %q, 期望回退为用户名 user1", updated.Nickname)
	}
}

// TestChangePasswordSuccess 验证旧密码校验通过后密码被替换为新的哈希。
func TestChangePasswordSuccess(t *testing.T) {
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("old12345678"), BcryptCost)
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Username: "user1", Password: string(oldHash), Role: "user", Status: 1}}
	svc := newTestUserService(userRepo)

	if err := svc.ChangePassword(1, &ChangePasswordRequest{OldPassword: "old12345678", NewPassword: "new12345678"}); err != nil {
		t.Fatalf("修改密码失败: %v", err)
	}
	// 新密码可被校验通过。
	if err := bcrypt.CompareHashAndPassword([]byte(userRepo.user.Password), []byte("new12345678")); err != nil {
		t.Errorf("密码未正确更新: %v", err)
	}
}

// TestChangePasswordRejectsWrongOldPassword 验证旧密码错误时拒绝修改。
func TestChangePasswordRejectsWrongOldPassword(t *testing.T) {
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("old12345678"), BcryptCost)
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Username: "user1", Password: string(oldHash), Role: "user", Status: 1}}
	svc := newTestUserService(userRepo)

	err := svc.ChangePassword(1, &ChangePasswordRequest{OldPassword: "wrong-pass", NewPassword: "new12345678"})
	if err == nil || !strings.Contains(err.Error(), "旧密码不正确") {
		t.Errorf("旧密码错误应被拒绝, 实际为 %v", err)
	}
}
