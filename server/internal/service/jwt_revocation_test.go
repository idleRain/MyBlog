package service

import (
	"testing"
	"time"

	"MyBlog/internal/domain"
)

// TestLogoutRevokesAccessToken 登出撤销 access 后，同串令牌必须校验失败。
// 该用例锚定 BE-01 键失配回归：撤销键与校验键必须经由同一归一化路径。
func TestLogoutRevokesAccessToken(t *testing.T) {
	svc := newTestJWTService(t)
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
	svc := newTestJWTService(t)
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

// TestRevokeKeyAcceptsFullTokenForm 撤销 payload-only 串后，完整三段 JWT 形式
// 的同一令牌同样必须命中撤销状态，两种形态的撤销键必须归一化到同一值。
func TestRevokeKeyAcceptsFullTokenForm(t *testing.T) {
	svc := newTestJWTService(t)
	user := &domain.User{ID: 9}

	pair, err := svc.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	if err := svc.RevokeToken(pair.AccessToken); err != nil {
		t.Fatalf("撤销访问令牌失败: %v", err)
	}

	fullToken, err := svc.ReconstructFullToken(pair.AccessToken, AccessToken)
	if err != nil {
		t.Fatalf("重构完整令牌失败: %v", err)
	}

	if !svc.IsTokenRevoked(fullToken) {
		t.Fatal("完整三段形式的已撤销令牌未命中撤销状态")
	}
	if _, err := svc.ValidateAccessToken(fullToken); err == nil {
		t.Fatal("完整三段形式的已撤销令牌仍可通过校验")
	}
}

// TestRevokedTokenClearedAfterExpiry 撤销记录以令牌过期时间为生命周期，
// 已过期的撤销记录应被清理，撤销表不随时间无界增长。
func TestRevokedTokenClearedAfterExpiry(t *testing.T) {
	svc := newTestJWTService(t).(*jwtService)
	user := &domain.User{ID: 10}

	// 构造一个已经过期一小时的历史令牌，模拟积压的过期撤销记录。
	expiredToken, err := svc.generateToken(user, AccessToken, time.Now().Add(-2*time.Hour), time.Hour)
	if err != nil {
		t.Fatalf("构造过期令牌失败: %v", err)
	}

	if err := svc.RevokeToken(expiredToken); err != nil {
		t.Fatalf("撤销过期令牌失败: %v", err)
	}

	if len(svc.revokedTokens) != 0 {
		t.Fatalf("过期撤销记录应被清理，剩余 %d 条", len(svc.revokedTokens))
	}
	if svc.IsTokenRevoked(expiredToken) {
		t.Fatal("已过期令牌不应维持撤销状态")
	}
}

// TestGeneratedTokensAreUniquePerIssue 同一用户连续生成的令牌必须互不相同。
// 撤销键以 payload 为准，若同一秒内两次签发产生同串令牌，旋转撤销会误杀新令牌。
func TestGeneratedTokensAreUniquePerIssue(t *testing.T) {
	svc := newTestJWTService(t)
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
		t.Fatal("两次签发的访问令牌完全相同，撤销键无法区分令牌实例")
	}
	if first.RefreshToken == second.RefreshToken {
		t.Fatal("两次签发的刷新令牌完全相同，撤销键无法区分令牌实例")
	}
}
