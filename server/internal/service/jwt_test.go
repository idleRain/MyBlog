package service

import (
	"sync"
	"testing"

	"MyBlog/internal/config"
	"MyBlog/internal/domain"
)

// newTestJWTService 构造携带最小合法配置的 JWT 服务实例。
func newTestJWTService(t *testing.T) JWTService {
	t.Helper()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:  "test-access-secret",
			RefreshSecret: "test-refresh-secret",
			AccessExpire:  30,
			RefreshExpire: 24,
		},
	}
	return NewJWTService(cfg)
}

// TestRevokeToken 验证撤销后的令牌被识别为已撤销。
func TestRevokeToken(t *testing.T) {
	svc := newTestJWTService(t)

	token := "revoked-sample-token"
	if svc.IsTokenRevoked(token) {
		t.Fatal("未撤销的令牌不应被判定为已撤销")
	}

	if err := svc.RevokeToken(token); err != nil {
		t.Fatalf("撤销令牌失败: %v", err)
	}

	if !svc.IsTokenRevoked(token) {
		t.Fatal("撤销后的令牌应被判定为已撤销")
	}
}

// TestRevokeTokenConcurrent 验证并发撤销与查询不会产生数据竞争，配合 -race 运行。
func TestRevokeTokenConcurrent(t *testing.T) {
	svc := newTestJWTService(t)

	const (
		workerCount = 8
		tokenCount  = 200
	)

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(base int) {
			defer wg.Done()
			for j := 0; j < tokenCount; j++ {
				token := string(rune('a'+base)) + string(rune('a'+j))
				_ = svc.RevokeToken(token)
				_ = svc.IsTokenRevoked(token)
			}
		}(i)
	}
	wg.Wait()
}

// TestGenerateTokenPairRoundTrip 验证生成的访问令牌可被校验且携带正确的用户 ID。
func TestGenerateTokenPairRoundTrip(t *testing.T) {
	svc := newTestJWTService(t)

	user := &domain.User{ID: 42}
	pair, err := svc.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("校验访问令牌失败: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("访问令牌用户 ID = %d, 期望 42", claims.UserID)
	}
}
