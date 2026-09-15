package service

import (
	"errors"
	"strings"
	"testing"

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

// loginJWTService 登录场景的 JWT 服务替身，仅覆盖令牌对生成方法。
type loginJWTService struct {
	JWTService
	tokenPair *TokenPair
	err       error
}

func (f *loginJWTService) GenerateTokenPair(user *domain.User) (*TokenPair, error) {
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
func newLoginUserService(repo *loginUserRepo, jwt *loginJWTService, rbac *loginRBACService) *userService {
	return NewUserService(repo, jwt, rbac).(*userService)
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
	jwt := &loginJWTService{tokenPair: &TokenPair{AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresIn: 1800}}
	rbac := &loginRBACService{permissions: []Permission{PermissionArticleRead}}
	svc := newLoginUserService(repo, jwt, rbac)

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
}

// TestLoginUserNotFound 验证用户名与邮箱均查不到用户时返回用户不存在。
func TestLoginUserNotFound(t *testing.T) {
	repo := &loginUserRepo{usernameErr: errors.New("用户不存在"), emailErr: errors.New("用户不存在")}
	svc := newLoginUserService(repo, &loginJWTService{}, &loginRBACService{})

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
	svc := newLoginUserService(repo, &loginJWTService{}, &loginRBACService{})

	_, err := svc.Login("admin", "wrong-pass1")
	if err == nil || !strings.Contains(err.Error(), "密码错误") {
		t.Errorf("应返回密码错误，实际为 %v", err)
	}
}

// TestLoginDisabledUser 验证用户状态非正常时拒绝登录。
func TestLoginDisabledUser(t *testing.T) {
	user := &domain.User{
		ID: 1, Username: "admin", Password: hashPassword(t, "correct-pass1"), Role: "user", Status: 0,
	}
	repo := &loginUserRepo{user: user}
	svc := newLoginUserService(repo, &loginJWTService{}, &loginRBACService{})

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
	jwt := &loginJWTService{tokenPair: &TokenPair{AccessToken: "a", RefreshToken: "r", ExpiresIn: 1}}
	svc := newLoginUserService(repo, jwt, &loginRBACService{})

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
	jwt := &loginJWTService{err: errors.New("signing failed")}
	svc := newLoginUserService(repo, jwt, &loginRBACService{})

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
