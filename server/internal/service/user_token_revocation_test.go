package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"MyBlog/internal/domain"
)

// recordedTokenService 记录撤销调用的 令牌服务替身，供登出与改密撤销链路断言。
type recordedTokenService struct {
	tokenService
	revokedTokens  []string
	revokedUserIDs []uint
}

func (f *recordedTokenService) GenerateTokenPair(*domain.User) (*TokenPair, error) {
	// 成功登录路径需要真实令牌对，替身返回固定值避免依赖内嵌空接口。
	return &TokenPair{AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresIn: 900}, nil
}

func (f *recordedTokenService) RevokeToken(tokenString string) error {
	f.revokedTokens = append(f.revokedTokens, tokenString)
	return nil
}

func (f *recordedTokenService) RevokeUserTokens(userID uint) error {
	f.revokedUserIDs = append(f.revokedUserIDs, userID)
	return nil
}

// TestLogoutRevokesTokenPair 登出必须同时撤销访问令牌与刷新令牌，
// 仅撤销 access 会让 refresh 在有效期内继续可刷。
func TestLogoutRevokesTokenPair(t *testing.T) {
	jwtFake := &recordedTokenService{}
	svc := NewUserService(&fakeUserRepo{user: &domain.User{ID: 1}}, jwtFake, NewRBACService())

	if err := svc.Logout("access-token", "refresh-token"); err != nil {
		t.Fatalf("登出失败: %v", err)
	}

	if len(jwtFake.revokedTokens) != 2 {
		t.Fatalf("应撤销 2 个令牌，实际撤销 %d 个", len(jwtFake.revokedTokens))
	}
	if jwtFake.revokedTokens[0] != "access-token" || jwtFake.revokedTokens[1] != "refresh-token" {
		t.Errorf("撤销顺序 = %v, 期望先 access 后 refresh", jwtFake.revokedTokens)
	}
}

// TestLogoutSkipsEmptyRefreshToken 刷新令牌缺省时仅撤销访问令牌，跳过空撤销。
func TestLogoutSkipsEmptyRefreshToken(t *testing.T) {
	jwtFake := &recordedTokenService{}
	svc := NewUserService(&fakeUserRepo{user: &domain.User{ID: 1}}, jwtFake, NewRBACService())

	if err := svc.Logout("access-token", ""); err != nil {
		t.Fatalf("登出失败: %v", err)
	}

	if len(jwtFake.revokedTokens) != 1 || jwtFake.revokedTokens[0] != "access-token" {
		t.Errorf("应仅撤销访问令牌，实际撤销 %v", jwtFake.revokedTokens)
	}
}

// TestChangePasswordRevokesAllUserTokens 修改密码成功后必须撤销该用户全部既有令牌，
// 防止持有旧令牌的会话在密码已更换的情况下继续存活。
func TestChangePasswordRevokesAllUserTokens(t *testing.T) {
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("old12345678"), BcryptCost)
	userRepo := &fakeUserRepo{user: &domain.User{ID: 3, Username: "user3", Password: string(oldHash), Role: "user", Status: 1}}
	jwtFake := &recordedTokenService{}
	svc := NewUserService(userRepo, jwtFake, NewRBACService())

	if err := svc.ChangePassword(3, &ChangePasswordRequest{OldPassword: "old12345678", NewPassword: "new12345678"}); err != nil {
		t.Fatalf("修改密码失败: %v", err)
	}

	if len(jwtFake.revokedUserIDs) != 1 || jwtFake.revokedUserIDs[0] != 3 {
		t.Errorf("应撤销用户 3 的全部令牌，实际撤销 %v", jwtFake.revokedUserIDs)
	}
}

// TestChangePasswordKeepsTokensOnWrongOldPassword 旧密码校验失败时不得触发撤销。
func TestChangePasswordKeepsTokensOnWrongOldPassword(t *testing.T) {
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("old12345678"), BcryptCost)
	userRepo := &fakeUserRepo{user: &domain.User{ID: 4, Username: "user4", Password: string(oldHash), Role: "user", Status: 1}}
	jwtFake := &recordedTokenService{}
	svc := NewUserService(userRepo, jwtFake, NewRBACService())

	err := svc.ChangePassword(4, &ChangePasswordRequest{OldPassword: "wrong-pass1", NewPassword: "new12345678"})
	if err == nil {
		t.Fatal("旧密码错误应被拒绝")
	}

	if len(jwtFake.revokedUserIDs) != 0 {
		t.Errorf("改密失败不应触发撤销，实际撤销 %v", jwtFake.revokedUserIDs)
	}
}

// TestRevokeUserTokensInvalidatesIssuedTokens 按 userID 撤销后，
// 该用户已签发的访问与刷新令牌必须全部失效。
func TestRevokeUserTokensInvalidatesIssuedTokens(t *testing.T) {
	svc := newTestTokenService(t)
	user := &domain.User{ID: 21}

	pair, err := svc.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	if err := svc.RevokeUserTokens(user.ID); err != nil {
		t.Fatalf("按用户撤销失败: %v", err)
	}

	if _, err := svc.ValidateAccessToken(pair.AccessToken); err == nil {
		t.Fatal("按用户撤销后访问令牌仍可通过校验")
	}
	if _, err := svc.ValidateRefreshToken(pair.RefreshToken); err == nil {
		t.Fatal("按用户撤销后刷新令牌仍可通过校验")
	}
}

// TestRevokeUserTokensOnlyAffectsTargetUser 按 userID 撤销不得波及其他用户的令牌。
func TestRevokeUserTokensOnlyAffectsTargetUser(t *testing.T) {
	svc := newTestTokenService(t)
	target := &domain.User{ID: 22}
	bystander := &domain.User{ID: 23}

	targetPair, err := svc.GenerateTokenPair(target)
	if err != nil {
		t.Fatalf("生成目标用户令牌失败: %v", err)
	}
	bystanderPair, err := svc.GenerateTokenPair(bystander)
	if err != nil {
		t.Fatalf("生成旁观用户令牌失败: %v", err)
	}

	if err := svc.RevokeUserTokens(target.ID); err != nil {
		t.Fatalf("按用户撤销失败: %v", err)
	}

	if _, err := svc.ValidateAccessToken(bystanderPair.AccessToken); err != nil {
		t.Fatalf("其他用户的令牌不应被波及: %v", err)
	}
	if _, err := svc.ValidateAccessToken(targetPair.AccessToken); err == nil {
		t.Fatal("目标用户的令牌应被撤销")
	}
}
