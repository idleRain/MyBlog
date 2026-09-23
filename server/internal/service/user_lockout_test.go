package service

import (
	"strings"
	"testing"
	"time"

	"MyBlog/internal/domain"
)

// loginUpdateSnapshot 捕获一次用户更新时的失败计数与锁定截止时间。
type loginUpdateSnapshot struct {
	failedLoginCount uint
	lockedUntil      *time.Time
}

// lockoutUserRepo 登录锁定测试的用户仓储替身，在 fakeUserRepo 基础上补充用户名查找路径。
type lockoutUserRepo struct {
	fakeUserRepo
}

func (f *lockoutUserRepo) GetByUsername(string) (*domain.User, error) {
	return f.user, nil
}

func (f *lockoutUserRepo) GetByEmail(string) (*domain.User, error) {
	return f.user, nil
}

// newLockoutTestService 创建注入锁定策略的用户服务与更新快照通道。
func newLockoutTestService(user *domain.User, policy LoginLockoutPolicy) (*userService, *[]loginUpdateSnapshot) {
	snapshots := &[]loginUpdateSnapshot{}
	repo := &lockoutUserRepo{fakeUserRepo{user: user}}
	repo.updateFunc = func(updated *domain.User) error {
		*snapshots = append(*snapshots, loginUpdateSnapshot{
			failedLoginCount: updated.FailedLoginCount,
			lockedUntil:      updated.LockedUntil,
		})
		return nil
	}
	svc := NewUserService(repo, &recordedTokenService{}, NewRBACService(),
		WithLoginLockoutPolicy(policy)).(*userService)
	return svc, snapshots
}

// lockedLoginUser 构造携带失败计数与锁定时间的登录测试用户。
func lockedLoginUser(t *testing.T, failedCount uint, lockedUntil *time.Time) *domain.User {
	t.Helper()
	return &domain.User{
		ID:               1,
		Username:         "admin",
		Password:         hashPassword(t, "correct-pass1"),
		Role:             "user",
		Status:           1,
		FailedLoginCount: failedCount,
		LockedUntil:      lockedUntil,
	}
}

// TestLoginFailureAccumulatesUntilLockout 连续失败逐次累计，达到阈值时锁定账户。
func TestLoginFailureAccumulatesUntilLockout(t *testing.T) {
	policy := LoginLockoutPolicy{Enabled: true, MaxFailedLogins: 3, LockDuration: 15 * time.Minute}
	svc, snapshots := newLockoutTestService(lockedLoginUser(t, 0, nil), policy)

	for attempt := 1; attempt <= 3; attempt++ {
		_, err := svc.Login("admin", "wrong-pass1")
		if err == nil {
			t.Fatalf("第 %d 次错误密码登录应被拒绝", attempt)
		}
		if !strings.Contains(err.Error(), "密码错误") {
			t.Errorf("第 %d 次失败应返回密码错误，实际为 %v", attempt, err)
		}
	}

	if len(*snapshots) != 3 {
		t.Fatalf("应产生 3 次计数更新，实际 %d 次", len(*snapshots))
	}
	last := (*snapshots)[2]
	if last.failedLoginCount != 3 {
		t.Errorf("第 3 次失败后计数 = %d, 期望 3", last.failedLoginCount)
	}
	if last.lockedUntil == nil || !last.lockedUntil.After(time.Now()) {
		t.Error("达到阈值后应设置未来的锁定截止时间")
	}
}

// TestLoginLockedAccountRejectsCorrectPassword 锁定期间即使密码正确也拒绝登录。
func TestLoginLockedAccountRejectsCorrectPassword(t *testing.T) {
	policy := LoginLockoutPolicy{Enabled: true, MaxFailedLogins: 3, LockDuration: 15 * time.Minute}
	futureLock := time.Now().Add(10 * time.Minute)
	svc, snapshots := newLockoutTestService(lockedLoginUser(t, 3, &futureLock), policy)

	_, err := svc.Login("admin", "correct-pass1")
	if err == nil || !strings.Contains(err.Error(), "锁定") {
		t.Errorf("锁定账户应被拒绝，实际为 %v", err)
	}
	if len(*snapshots) != 0 {
		t.Error("锁定拒绝路径不应写回用户数据")
	}
}

// TestLoginExpiredLockoutAllowsRetry 锁定到期后自动解除，重新登录成功且计数清零。
func TestLoginExpiredLockoutAllowsRetry(t *testing.T) {
	policy := LoginLockoutPolicy{Enabled: true, MaxFailedLogins: 3, LockDuration: 15 * time.Minute}
	expiredLock := time.Now().Add(-10 * time.Minute)
	svc, snapshots := newLockoutTestService(lockedLoginUser(t, 3, &expiredLock), policy)

	response, err := svc.Login("admin", "correct-pass1")
	if err != nil {
		t.Fatalf("锁定到期后的登录应成功: %v", err)
	}
	if response == nil {
		t.Fatal("登录响应不应为空")
	}

	foundReset := false
	for _, snapshot := range *snapshots {
		if snapshot.failedLoginCount == 0 && snapshot.lockedUntil == nil {
			foundReset = true
		}
	}
	if !foundReset {
		t.Error("登录成功后应写回清零的计数与空锁定时间")
	}
}

// TestLoginSuccessResetsFailureCount 未触发锁定的成功登录应清零历史失败计数。
func TestLoginSuccessResetsFailureCount(t *testing.T) {
	policy := LoginLockoutPolicy{Enabled: true, MaxFailedLogins: 5, LockDuration: 15 * time.Minute}
	svc, snapshots := newLockoutTestService(lockedLoginUser(t, 2, nil), policy)

	if _, err := svc.Login("admin", "correct-pass1"); err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	if len(*snapshots) != 1 || (*snapshots)[0].failedLoginCount != 0 {
		t.Errorf("成功登录应清零失败计数，实际快照 %v", *snapshots)
	}
}

// TestLoginLockoutDisabledKeepsCountingOnly 锁定策略关闭时仅累计计数，不设置锁定时间。
func TestLoginLockoutDisabledKeepsCountingOnly(t *testing.T) {
	policy := LoginLockoutPolicy{Enabled: false, MaxFailedLogins: 2, LockDuration: 15 * time.Minute}
	svc, snapshots := newLockoutTestService(lockedLoginUser(t, 0, nil), policy)

	for attempt := 0; attempt < 5; attempt++ {
		_, err := svc.Login("admin", "wrong-pass1")
		if err == nil || !strings.Contains(err.Error(), "密码错误") {
			t.Fatalf("第 %d 次失败应返回密码错误，实际为 %v", attempt+1, err)
		}
	}

	if len(*snapshots) != 5 {
		t.Fatalf("应产生 5 次计数更新，实际 %d 次", len(*snapshots))
	}
	for _, snapshot := range *snapshots {
		if snapshot.lockedUntil != nil {
			t.Fatal("策略关闭时不应设置锁定时间")
		}
	}
}
