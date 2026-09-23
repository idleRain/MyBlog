package service

import (
	"testing"
	"time"

	"MyBlog/internal/domain"
)

// TestLogoutRevokesAccessToken 登出撤销 access 后，同串令牌必须校验失败。
func TestLogoutRevokesAccessToken(t *testing.T) {
	svc := newTestTokenService(t)
	user := &domain.User{ID: 7}

	pair, err := svc.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	if err := svc.RevokeToken(pair.AccessToken); err != nil {
		t.Fatalf("撤销访问令牌失败: %v", err)
	}

	if _, err := svc.ValidateAccessToken(pair.AccessToken); err == nil {
		t.Fatal("登出后访问令牌仍可通过校验，撤销检查未命中")
	}
}

// TestRefreshRotationInvalidatesOldRefreshToken 刷新旋转后旧 refresh 必须失效，
// 新令牌对必须仍然可用，防止旧刷新令牌在有效期内被无限再刷。
func TestRefreshRotationInvalidatesOldRefreshToken(t *testing.T) {
	svc := newTestTokenService(t)
	user := &domain.User{ID: 8}

	pair, err := svc.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	newPair, err := svc.RefreshAccessToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("刷新令牌失败: %v", err)
	}

	if _, err := svc.ValidateRefreshToken(pair.RefreshToken); err == nil {
		t.Fatal("旋转后旧刷新令牌仍可通过校验")
	}
	if _, err := svc.ValidateRefreshToken(newPair.RefreshToken); err != nil {
		t.Fatalf("旋转后的新刷新令牌应可正常校验: %v", err)
	}
	if _, err := svc.ValidateAccessToken(newPair.AccessToken); err != nil {
		t.Fatalf("旋转后的新访问令牌应可正常校验: %v", err)
	}
}

// TestExpiredTokenRejectedAndPruned 过期令牌必须校验失败，
// 并随下一次签发被惰性清理，令牌表不随时间无界增长。
func TestExpiredTokenRejectedAndPruned(t *testing.T) {
	svc := newTestTokenService(t).(*tokenService)
	user := &domain.User{ID: 10}

	// 构造一个已经过期两小时的历史令牌，模拟积压的过期记录。
	expiredToken, err := svc.issueToken(user.ID, AccessToken, time.Now().Add(-2*time.Hour), time.Hour)
	if err != nil {
		t.Fatalf("构造过期令牌失败: %v", err)
	}

	if _, err := svc.ValidateAccessToken(expiredToken); err == nil {
		t.Fatal("已过期令牌不应通过校验")
	}

	// 触发一次签发，过期记录应被惰性清理。
	if _, err := svc.GenerateTokenPair(user); err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	if _, exists := svc.tokens[expiredToken]; exists {
		t.Fatal("过期令牌记录未被清理")
	}
}

// TestGeneratedTokensAreUniquePerIssue 同一用户连续生成的令牌必须互不相同，
// 否则撤销单个令牌会误伤同批签发的其他令牌。
func TestGeneratedTokensAreUniquePerIssue(t *testing.T) {
	svc := newTestTokenService(t)
	user := &domain.User{ID: 11}

	first, err := svc.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("生成第一组令牌失败: %v", err)
	}
	second, err := svc.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("生成第二组令牌失败: %v", err)
	}

	if first.AccessToken == second.AccessToken {
		t.Fatal("两次签发的访问令牌完全相同，无法区分令牌实例")
	}
	if first.RefreshToken == second.RefreshToken {
		t.Fatal("两次签发的刷新令牌完全相同，无法区分令牌实例")
	}
}
