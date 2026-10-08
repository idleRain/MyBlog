package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"MyBlog/internal/domain"
)

// lockoutStateLog 记录登录计数相关列在一次落库后的状态快照。
type lockoutStateLog struct {
	failedLoginCount uint
	lockedUntil      *time.Time
}

// copyTimePtr 复制时间指针，避免后续修改改写已记录快照的内容。
func copyTimePtr(source *time.Time) *time.Time {
	if source == nil {
		return nil
	}
	cloned := *source
	return &cloned
}

// lockoutUserRepo 登录锁定测试的用户仓储替身。
// 模拟原子更新语句的列级效果：失败自增仅演进计数，锁定与清零按条件更新，
// 其余字段不参与写回，用于验证服务编排与状态演进语义。
type lockoutUserRepo struct {
	fakeUserRepo
	failures    uint
	lockedUntil *time.Time
	// eventLog 收集每次落库后的计数与锁定快照。
	eventLog     *[]lockoutStateLog
	incrementErr error
}

func (f *lockoutUserRepo) GetByUsername(string) (*domain.User, error) {
	return f.user, nil
}

func (f *lockoutUserRepo) GetByEmail(string) (*domain.User, error) {
	return f.user, nil
}

// record 复制当前列状态到事件日志，取代旧版全字段 Update 快照。
func (f *lockoutUserRepo) record() {
	*f.eventLog = append(*f.eventLog, lockoutStateLog{
		failedLoginCount: f.failures,
		lockedUntil:      copyTimePtr(f.lockedUntil),
	})
}

// IncrementLoginFailures 模拟数据库侧自增，仅演进失败计数列。
func (f *lockoutUserRepo) IncrementLoginFailures(uint) error {
	if f.incrementErr != nil {
		return f.incrementErr
	}
	f.failures++
	f.record()
	return nil
}

// LockUserAfterFailures 模拟阈值条件更新，计数未达阈值时不写锁定。
func (f *lockoutUserRepo) LockUserAfterFailures(_ uint, maxAttempts uint, lockedUntil time.Time) (bool, error) {
	if f.failures < maxAttempts {
		return false, nil
	}
	until := lockedUntil
	f.lockedUntil = &until
	f.record()
	return true, nil
}

// ResetLoginFailures 模拟清零与清除锁定的原子列更新。
func (f *lockoutUserRepo) ResetLoginFailures(uint) error {
	f.failures = 0
	f.lockedUntil = nil
	f.record()
	return nil
}

// newLockoutTestService 创建注入锁定策略的用户服务与状态快照通道。
func newLockoutTestService(user *domain.User, policy LoginLockoutPolicy) (*userService, *[]lockoutStateLog) {
	events := &[]lockoutStateLog{}
	repo := &lockoutUserRepo{
		fakeUserRepo: fakeUserRepo{user: user},
		eventLog:     events,
		incrementErr: errors.New("未使用"),
	}
	repo.incrementErr = nil
	svc := NewUserService(repo, &recordedTokenService{}, NewRBACService(),
		WithLoginLockoutPolicy(policy)).(*userService)
	return svc, events
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
	svc, events := newLockoutTestService(lockedLoginUser(t, 0, nil), policy)

	for attempt := 1; attempt <= 3; attempt++ {
		_, err := svc.Login("admin", "wrong-pass1")
		if err == nil {
			t.Fatalf("第 %d 次错误密码登录应被拒绝", attempt)
		}
		if !strings.Contains(err.Error(), "密码错误") {
			t.Errorf("第 %d 次失败应返回密码错误，实际为 %v", attempt, err)
		}
	}

	if len(*events) != 4 {
		t.Fatalf("应产生 3 次计数更新加 1 次阈值命中锁定写入，实际 %v", *events)
	}
	last := (*events)[len(*events)-1]
	if last.failedLoginCount != 3 {
		t.Errorf("第 3 次失败后计数 = %d, 期望 3", last.failedLoginCount)
	}
	if last.lockedUntil == nil || !last.lockedUntil.After(time.Now()) {
		t.Error("达到阈值后应设置未来的锁定截止时间")
	}
}

// TestLoginLockedAccountRejectsCorrectPassword 锁定期间即使密码正确也拒绝登录。
// 该路径在计数与锁定写入前被拦截，原子更新不应被触达。
func TestLoginLockedAccountRejectsCorrectPassword(t *testing.T) {
	policy := LoginLockoutPolicy{Enabled: true, MaxFailedLogins: 3, LockDuration: 15 * time.Minute}
	futureLock := time.Now().Add(10 * time.Minute)
	svc, events := newLockoutTestService(lockedLoginUser(t, 3, &futureLock), policy)

	_, err := svc.Login("admin", "correct-pass1")
	if err == nil || !strings.Contains(err.Error(), "锁定") {
		t.Errorf("锁定账户应被拒绝，实际为 %v", err)
	}
	if len(*events) != 0 {
		t.Error("锁定拒绝路径不应写回用户数据")
	}
}

// TestLoginExpiredLockoutAllowsRetry 锁定到期后自动解除，重新登录成功且计数清零。
func TestLoginExpiredLockoutAllowsRetry(t *testing.T) {
	policy := LoginLockoutPolicy{Enabled: true, MaxFailedLogins: 3, LockDuration: 15 * time.Minute}
	expiredLock := time.Now().Add(-10 * time.Minute)
	svc, events := newLockoutTestService(lockedLoginUser(t, 3, &expiredLock), policy)

	response, err := svc.Login("admin", "correct-pass1")
	if err != nil {
		t.Fatalf("锁定到期后的登录应成功: %v", err)
	}
	if response == nil {
		t.Fatal("登录响应不应为空")
	}

	if len(*events) != 1 {
		t.Fatalf("成功登录应产生一次清零落库，实际 %v", *events)
	}
	reset := (*events)[0]
	if reset.failedLoginCount != 0 || reset.lockedUntil != nil {
		t.Errorf("登录成功后应写回清零的计数与空锁定时间，实际 %v", reset)
	}
}

// TestLoginSuccessResetsFailureCount 未触发锁定的成功登录应清零历史失败计数。
func TestLoginSuccessResetsFailureCount(t *testing.T) {
	policy := LoginLockoutPolicy{Enabled: true, MaxFailedLogins: 5, LockDuration: 15 * time.Minute}
	svc, events := newLockoutTestService(lockedLoginUser(t, 2, nil), policy)

	if _, err := svc.Login("admin", "correct-pass1"); err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	if len(*events) != 1 || (*events)[0].failedLoginCount != 0 {
		t.Errorf("成功登录应清零失败计数，实际快照 %v", *events)
	}
}

// TestLoginLockoutDisabledKeepsCountingOnly 锁定策略关闭时仅累计计数，不设置锁定时间。
func TestLoginLockoutDisabledKeepsCountingOnly(t *testing.T) {
	policy := LoginLockoutPolicy{Enabled: false, MaxFailedLogins: 2, LockDuration: 15 * time.Minute}
	svc, events := newLockoutTestService(lockedLoginUser(t, 0, nil), policy)

	for attempt := 0; attempt < 5; attempt++ {
		_, err := svc.Login("admin", "wrong-pass1")
		if err == nil || !strings.Contains(err.Error(), "密码错误") {
			t.Fatalf("第 %d 次失败应返回密码错误，实际为 %v", attempt+1, err)
		}
	}

	if len(*events) != 5 {
		t.Fatalf("应产生 5 次计数更新，实际 %d 次", len(*events))
	}
	for _, snapshot := range *events {
		if snapshot.lockedUntil != nil {
			t.Fatal("策略关闭时不应设置锁定时间")
		}
	}
}
