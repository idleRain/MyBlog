package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"MyBlog/internal/domain"
	"MyBlog/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// loginUserRepo 登录场景的用户仓储替身，按用户名与邮箱分别返回可配置结果。
type loginUserRepo struct {
	repository.UserRepository
	user        *domain.User
	usernameErr error
	emailErr    error
	// failureCalls / lockCalls / resetCalls 记录登录计数相关仓储方法的触达次数，
	// 并保存最近一次锁定阈值传入值，用于断言原子更新路径被真实走通。
	failureCalls int
	lockCalls    int
	lockAttempts uint
	resetCalls   int
}

func (f *loginUserRepo) GetByUsername(username string) (*domain.User, error) {
	if f.usernameErr != nil {
		return nil, f.usernameErr
	}
	return f.user, nil
}

func (f *loginUserRepo) GetByEmail(email string) (*domain.User, error) {
	if f.emailErr != nil {
		return nil, f.emailErr
	}
	return f.user, nil
}

// IncrementLoginFailures 登录失败累计的替身实现，记录触达次数。
func (f *loginUserRepo) IncrementLoginFailures(uint) error {
	f.failureCalls++
	return nil
}

// LockUserAfterFailures 锁定写入的替身实现，记录阈值传入值并模拟条件命中结果。
func (f *loginUserRepo) LockUserAfterFailures(_ uint, maxAttempts uint, _ time.Time) (bool, error) {
	f.lockCalls++
	f.lockAttempts = maxAttempts
	// 阈值为零值时视作策略未注入，模拟条件不命中。
	if f.lockAttempts == 0 {
		return false, nil
	}
	return uint(f.failureCalls) >= f.lockAttempts, nil
}

// ResetLoginFailures 成功登录清零的替身实现，记录触达次数。
func (f *loginUserRepo) ResetLoginFailures(uint) error {
	f.resetCalls++
	return nil
}

// loginTokenService 登录场景的 令牌服务替身，仅覆盖令牌对生成方法。
type loginTokenService struct {
	tokenService
	tokenPair *TokenPair
	err       error
}

func (f *loginTokenService) GenerateTokenPair(user *domain.User) (*TokenPair, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.tokenPair, nil
}

// loginRBACService 登录场景的 RBAC 服务替身，仅覆盖权限查询方法。
type loginRBACService struct {
	RBACService
	permissions []Permission
}

func (f *loginRBACService) GetUserPermissions(userRole string) []Permission {
	return f.permissions
}

// newLoginUserService 创建注入登录替身的用户服务实例。
func newLoginUserService(repo *loginUserRepo, tokenSvc *loginTokenService, rbac *loginRBACService) *userService {
	return NewUserService(repo, tokenSvc, rbac).(*userService)
}

// hashPassword 以最低成本生成测试密码哈希，避免默认成本拖慢测试。
func hashPassword(t *testing.T, plain string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("生成测试密码哈希失败: %v", err)
	}
	return string(hash)
}

// TestLoginSuccess 验证用户名与密码均正确时返回完整登录响应。
func TestLoginSuccess(t *testing.T) {
	user := &domain.User{
		ID: 1, Username: "admin", Email: "admin@myblog.local",
		Password: hashPassword(t, "correct-pass1"), Role: "superadmin", Status: 1,
	}
	repo := &loginUserRepo{user: user}
	tokenSvc := &loginTokenService{tokenPair: &TokenPair{AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresIn: 1800}}
	rbac := &loginRBACService{permissions: []Permission{PermissionArticleRead}}
	svc := newLoginUserService(repo, tokenSvc, rbac)

	response, err := svc.Login("admin", "correct-pass1")
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	if response.User != user {
		t.Error("登录响应应携带与仓储一致的完整用户")
	}
	if response.AccessToken != "access-token" || response.RefreshToken != "refresh-token" {
		t.Errorf("令牌对未正确透传: %s / %s", response.AccessToken, response.RefreshToken)
	}
	if response.ExpiresIn != 1800 {
		t.Errorf("过期时间 = %d，期望 1800", response.ExpiresIn)
	}
	if len(response.Permissions) != 1 || response.Permissions[0] != string(PermissionArticleRead) {
		t.Errorf("权限列表 = %v，期望包含 article:read", response.Permissions)
	}
	// 计数本就为零的成功登录不应触发清零写库。
	if repo.resetCalls != 0 {
		t.Errorf("无失败记录的成功登录不应清零写库，实际触达 %d 次", repo.resetCalls)
	}
}

// TestLoginUserNotFound 验证用户名与邮箱均查不到用户时返回用户不存在。
func TestLoginUserNotFound(t *testing.T) {
	repo := &loginUserRepo{usernameErr: errors.New("用户不存在"), emailErr: errors.New("用户不存在")}
	svc := newLoginUserService(repo, &loginTokenService{}, &loginRBACService{})

	_, err := svc.Login("ghost", "whatever1")
	if err == nil || !strings.Contains(err.Error(), "用户不存在") {
		t.Errorf("应返回用户不存在，实际为 %v", err)
	}
}

// TestLoginWrongPassword 验证密码错误时返回密码错误。
func TestLoginWrongPassword(t *testing.T) {
	user := &domain.User{
		ID: 1, Username: "admin", Password: hashPassword(t, "correct-pass1"), Role: "user", Status: 1,
	}
	repo := &loginUserRepo{user: user}
	svc := newLoginUserService(repo, &loginTokenService{}, &loginRBACService{})

	_, err := svc.Login("admin", "wrong-pass1")
	if err == nil || !strings.Contains(err.Error(), "密码错误") {
		t.Errorf("应返回密码错误，实际为 %v", err)
	}

	// 失败计数必须经原子自增路径落库，未达到阈值时不应写锁定。
	if repo.failureCalls != 1 {
		t.Errorf("失败登录应触达一次原子计数，实际 %d 次", repo.failureCalls)
	}
	if repo.lockCalls != 1 {
		t.Errorf("开启锁定策略时应触达一次条件锁定更新，实际 %d 次", repo.lockCalls)
	}
}

// TestLoginLockoutMessageOnThreshold 验证失败次数达到阈值时返回锁定文案并写入锁定。
func TestLoginLockoutMessageOnThreshold(t *testing.T) {
	user := &domain.User{
		ID: 1, Username: "admin", Password: hashPassword(t, "correct-pass1"), Role: "user", Status: 1,
	}
	repo := &loginUserRepo{user: user}
	svc := newLoginUserService(repo, &loginTokenService{}, &loginRBACService{})
	setLockoutPolicy(svc, repo, 2, time.Minute)

	// 第一次失败未达阈值，使用普通文案。
	if _, err := svc.Login("admin", "wrong-pass1"); err == nil || strings.Contains(err.Error(), "锁定") {
		t.Errorf("未达阈值不应提示锁定，实际 %v", err)
	}
	// 第二次失败达到阈值，锁定更新应命中并返回锁定文案。
	_, err := svc.Login("admin", "wrong-pass1")
	if err == nil || !strings.Contains(err.Error(), "账户已被锁定") {
		t.Errorf("达到阈值应提示锁定，实际 %v", err)
	}
	if repo.failureCalls != 2 {
		t.Errorf("两次失败应累计计数两次，实际 %d 次", repo.failureCalls)
	}
}

// setLockoutPolicy 为测试服务实例与仓储替身注入一致的锁定策略，
// 替身凭最大失败次数模拟条件更新的阈值命中结果。
func setLockoutPolicy(svc *userService, repo *loginUserRepo, maxFailedLogins uint, lockDuration time.Duration) {
	svc.lockoutPolicy = LoginLockoutPolicy{
		Enabled:         true,
		MaxFailedLogins: maxFailedLogins,
		LockDuration:    lockDuration,
	}
	repo.lockAttempts = maxFailedLogins
}

// TestLoginSuccessResetsLockedState 验证带失败记录的成功登录会原子清零计数并清除锁定。
func TestLoginSuccessResetsLockedState(t *testing.T) {
	lockedUntil := time.Now().Add(-time.Minute)
	user := &domain.User{
		ID: 1, Username: "admin", Password: hashPassword(t, "correct-pass1"), Role: "user", Status: 1,
		FailedLoginCount: 2,
		LockedUntil:      &lockedUntil,
	}
	repo := &loginUserRepo{user: user}
	tokenSvc := &loginTokenService{tokenPair: &TokenPair{AccessToken: "a", RefreshToken: "r", ExpiresIn: 1}}
	svc := newLoginUserService(repo, tokenSvc, &loginRBACService{})

	// 锁定截止时间已过，登录允许成功。
	if _, err := svc.Login("admin", "correct-pass1"); err != nil {
		t.Fatalf("锁定过期后的成功登录不应失败: %v", err)
	}

	// 清零必须经原子语句走库且同步内存实体。
	if repo.resetCalls != 1 {
		t.Errorf("带失败记录的成功登录应触达一次清零，实际 %d 次", repo.resetCalls)
	}
	if user.FailedLoginCount != 0 || user.LockedUntil != nil {
		t.Errorf("成功登录后计数值与锁定应归零，实际 count=%d locked=%v", user.FailedLoginCount, user.LockedUntil != nil)
	}
}

// TestLoginDisabledUser 验证用户状态非正常时拒绝登录。
func TestLoginDisabledUser(t *testing.T) {
	user := &domain.User{
		ID: 1, Username: "admin", Password: hashPassword(t, "correct-pass1"), Role: "user", Status: 0,
	}
	repo := &loginUserRepo{user: user}
	svc := newLoginUserService(repo, &loginTokenService{}, &loginRBACService{})

	_, err := svc.Login("admin", "correct-pass1")
	if err == nil || !strings.Contains(err.Error(), "禁用") {
		t.Errorf("禁用用户应被拒绝，实际为 %v", err)
	}
}

// TestLoginFallsBackToEmail 验证用户名未命中时回退到邮箱查找。
func TestLoginFallsBackToEmail(t *testing.T) {
	user := &domain.User{
		ID: 1, Username: "admin", Email: "admin@myblog.local",
		Password: hashPassword(t, "correct-pass1"), Role: "user", Status: 1,
	}
	repo := &loginUserRepo{user: user, usernameErr: errors.New("用户不存在")}
	tokenSvc := &loginTokenService{tokenPair: &TokenPair{AccessToken: "a", RefreshToken: "r", ExpiresIn: 1}}
	svc := newLoginUserService(repo, tokenSvc, &loginRBACService{})

	response, err := svc.Login("admin@myblog.local", "correct-pass1")
	if err != nil {
		t.Fatalf("邮箱登录失败: %v", err)
	}
	if response.User != user {
		t.Error("邮箱登录应返回匹配用户")
	}
}

// TestLoginTokenGenerationFailure 验证令牌生成失败时登录报错。
func TestLoginTokenGenerationFailure(t *testing.T) {
	user := &domain.User{
		ID: 1, Username: "admin", Password: hashPassword(t, "correct-pass1"), Role: "user", Status: 1,
	}
	repo := &loginUserRepo{user: user}
	// 替身返回签发阶段的失败，令牌串未生成时登录必须整体失败。
	tokenSvc := &loginTokenService{err: errors.New("令牌签发失败")}
	svc := newLoginUserService(repo, tokenSvc, &loginRBACService{})

	_, err := svc.Login("admin", "correct-pass1")
	if err == nil {
		t.Error("令牌生成失败时登录应报错")
	}
}

// TestValidatePasswordStrength 表驱动验证密码强度边界。
func TestValidatePasswordStrength(t *testing.T) {
	svc := &userService{}

	testCases := []struct {
		name      string
		password  string
		expectErr bool
	}{
		{name: "少于 6 位被拒绝", password: "abc12", expectErr: true},
		{name: "超过 100 位被拒绝", password: strings.Repeat("a", 100) + "1", expectErr: true},
		{name: "纯数字被拒绝", password: "12345678", expectErr: true},
		{name: "纯字母被拒绝", password: "abcdefgh", expectErr: true},
		{name: "弱密码列表中的 admin 被拒绝", password: "admin", expectErr: true},
		{name: "弱密码列表中的 abc123 被拒绝", password: "abc123", expectErr: true},
		{name: "长度与构成合规的强密码通过", password: "k3Vb9xQ2mN", expectErr: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.validatePasswordStrength(tc.password)
			if (err != nil) != tc.expectErr {
				t.Errorf("校验结果 = %v，期望报错 = %v", err, tc.expectErr)
			}
		})
	}
}
