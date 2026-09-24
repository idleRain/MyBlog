package service

import (
	"errors"
	"strings"
	"testing"

	"MyBlog/internal/domain"
)

// errInvalidRefreshTokenForTest 模拟刷新令牌校验失败的替身错误。
var errInvalidRefreshTokenForTest = errors.New("刷新令牌无效")

// refreshTokenService 刷新链路的 令牌服务替身，返回可配置的身份与令牌对，
// 并记录 RefreshAccessToken 是否被调用，用于断言状态校验先于旋转发生。
type refreshTokenService struct {
	tokenService
	identity      *TokenIdentity
	identityErr   error
	pair          *TokenPair
	pairErr       error
	refreshCalled bool
}

func (f *refreshTokenService) ValidateRefreshToken(string) (*TokenIdentity, error) {
	if f.identityErr != nil {
		return nil, f.identityErr
	}
	return f.identity, nil
}

func (f *refreshTokenService) RefreshAccessToken(string) (*TokenPair, error) {
	f.refreshCalled = true
	if f.pairErr != nil {
		return nil, f.pairErr
	}
	return f.pair, nil
}

// TestRefreshTokenRejectsDisabledUser 刷新链路必须查库校验用户状态，
// 被禁用用户的 refresh 不得继续换取新令牌。
func TestRefreshTokenRejectsDisabledUser(t *testing.T) {
	repo := &fakeUserRepo{user: &domain.User{ID: 5, Username: "user5", Role: "user", Status: 0}}
	tokenFake := &refreshTokenService{
		identity: &TokenIdentity{UserID: 5},
		pair:     &TokenPair{AccessToken: "a", RefreshToken: "r"},
	}
	svc := NewUserService(repo, tokenFake, NewRBACService())

	_, err := svc.RefreshToken("refresh-token")
	if err == nil || !strings.Contains(err.Error(), "禁用") {
		t.Errorf("禁用用户的刷新请求应被拒绝，实际为 %v", err)
	}
	if tokenFake.refreshCalled {
		t.Error("状态校验失败时不应执行令牌旋转")
	}
}

// TestRefreshTokenRejectsMissingUser 令牌指向的用户不存在时刷新必须失败。
func TestRefreshTokenRejectsMissingUser(t *testing.T) {
	// 仓储中的用户 ID 与令牌身份不一致，模拟用户已被删除。
	repo := &fakeUserRepo{user: &domain.User{ID: 6, Username: "user6", Role: "user", Status: 1}}
	tokenFake := &refreshTokenService{identity: &TokenIdentity{UserID: 99}}
	svc := NewUserService(repo, tokenFake, NewRBACService())

	_, err := svc.RefreshToken("refresh-token")
	if err == nil {
		t.Fatal("用户不存在时刷新应被拒绝")
	}
	if tokenFake.refreshCalled {
		t.Error("用户不存在时不应执行令牌旋转")
	}
}

// TestRefreshTokenRejectsInvalidToken 刷新令牌本身无效时直接拒绝，不触发查库与旋转。
func TestRefreshTokenRejectsInvalidToken(t *testing.T) {
	repo := &fakeUserRepo{user: &domain.User{ID: 7, Username: "user7", Role: "user", Status: 1}}
	tokenFake := &refreshTokenService{identityErr: errInvalidRefreshTokenForTest}
	svc := NewUserService(repo, tokenFake, NewRBACService())

	_, err := svc.RefreshToken("broken-token")
	if err == nil {
		t.Fatal("无效刷新令牌应被拒绝")
	}
	if tokenFake.refreshCalled {
		t.Error("无效令牌不应执行令牌旋转")
	}
}

// TestRefreshTokenAllowsActiveUser 正常用户的刷新链路应完成令牌旋转并返回新令牌对。
func TestRefreshTokenAllowsActiveUser(t *testing.T) {
	repo := &fakeUserRepo{user: &domain.User{ID: 8, Username: "user8", Role: "user", Status: 1}}
	tokenFake := &refreshTokenService{
		identity: &TokenIdentity{UserID: 8},
		pair:     &TokenPair{AccessToken: "new-a", RefreshToken: "new-r", ExpiresIn: 900},
	}
	svc := NewUserService(repo, tokenFake, NewRBACService())

	pair, err := svc.RefreshToken("refresh-token")
	if err != nil {
		t.Fatalf("正常用户刷新失败: %v", err)
	}
	if !tokenFake.refreshCalled {
		t.Error("正常用户应执行令牌旋转")
	}
	if pair.AccessToken != "new-a" || pair.RefreshToken != "new-r" {
		t.Errorf("令牌对未正确透传: %s / %s", pair.AccessToken, pair.RefreshToken)
	}
}
