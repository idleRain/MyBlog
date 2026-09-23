package service

import (
	"strings"
	"sync"
	"testing"

	"MyBlog/internal/config"
	"MyBlog/internal/domain"
)

// newTestTokenService 构造携带最小合法配置的令牌服务实例。
func newTestTokenService(t *testing.T) TokenServiceInterface {
	t.Helper()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessExpire:  30,
			RefreshExpire: 24,
		},
	}
	return NewTokenService(cfg)
}

// TestGenerateTokenPairRoundTrip 验证生成的访问令牌可被校验且携带正确的用户 ID。
func TestGenerateTokenPairRoundTrip(t *testing.T) {
	svc := newTestTokenService(t)

	user := &domain.User{ID: 42}
	pair, err := svc.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	identity, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("校验访问令牌失败: %v", err)
	}
	if identity.UserID != 42 {
		t.Errorf("访问令牌用户 ID = %d, 期望 42", identity.UserID)
	}
}

// TestIssuedTokenIsOpaque 验证签发结果为不透明随机串。
// 令牌不得携带可被解码的身份信息，也不得出现 JWT 的分段结构。
func TestIssuedTokenIsOpaque(t *testing.T) {
	svc := newTestTokenService(t)

	pair, err := svc.GenerateTokenPair(&domain.User{ID: 42})
	if err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	if strings.Contains(pair.AccessToken, ".") {
		t.Errorf("令牌不应包含分段分隔符: %s", pair.AccessToken)
	}

	expectedLength := tokenRandomBytes * 2
	if len(pair.AccessToken) != expectedLength {
		t.Errorf("访问令牌长度 = %d, 期望 %d", len(pair.AccessToken), expectedLength)
	}
	if len(pair.RefreshToken) != expectedLength {
		t.Errorf("刷新令牌长度 = %d, 期望 %d", len(pair.RefreshToken), expectedLength)
	}
}

// TestForgedTokenRejected 验证未经服务端签发的自造令牌无法通过校验。
// 该用例锚定认证绕过回归：令牌身份的唯一权威是服务端令牌表，
// 任何由调用方自行构造的串都不得被接受，无论其形态是否类似历史令牌。
func TestForgedTokenRejected(t *testing.T) {
	svc := newTestTokenService(t)

	// 覆盖三种典型伪造形态：历史 payload-only 串、随机十六进制串、空串。
	forgedTokens := []string{
		"eyJ1IjoxLCJqdGkiOiJ4IiwiZXhwIjo5OTk5OTk5OTk5fQ",
		"96a0f9cbe3174281a38529c30cda1d7a",
		"",
	}

	for _, forged := range forgedTokens {
		if _, err := svc.ValidateAccessToken(forged); err == nil {
			t.Errorf("自造令牌通过了访问令牌校验: %q", forged)
		}
		if _, err := svc.ValidateRefreshToken(forged); err == nil {
			t.Errorf("自造令牌通过了刷新令牌校验: %q", forged)
		}
	}
}

// TestTokenTypeNotInterchangeable 验证访问令牌与刷新令牌不可互换使用。
func TestTokenTypeNotInterchangeable(t *testing.T) {
	svc := newTestTokenService(t)

	pair, err := svc.GenerateTokenPair(&domain.User{ID: 43})
	if err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	if _, err := svc.ValidateRefreshToken(pair.AccessToken); err == nil {
		t.Error("访问令牌通过了刷新令牌校验，令牌类型未被区分")
	}
	if _, err := svc.ValidateAccessToken(pair.RefreshToken); err == nil {
		t.Error("刷新令牌通过了访问令牌校验，令牌类型未被区分")
	}
}

// TestRevokeToken 验证撤销后的令牌立即失效。
func TestRevokeToken(t *testing.T) {
	svc := newTestTokenService(t)

	pair, err := svc.GenerateTokenPair(&domain.User{ID: 44})
	if err != nil {
		t.Fatalf("生成令牌对失败: %v", err)
	}

	if _, err := svc.ValidateAccessToken(pair.AccessToken); err != nil {
		t.Fatalf("撤销前令牌应可通过校验: %v", err)
	}

	if err := svc.RevokeToken(pair.AccessToken); err != nil {
		t.Fatalf("撤销令牌失败: %v", err)
	}

	if _, err := svc.ValidateAccessToken(pair.AccessToken); err == nil {
		t.Fatal("撤销后令牌仍可通过校验")
	}
}

// TestRevokeTokenConcurrent 验证并发撤销与校验不产生数据竞争，配合 -race 运行。
func TestRevokeTokenConcurrent(t *testing.T) {
	svc := newTestTokenService(t)

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
				_, _ = svc.ValidateAccessToken(token)
			}
		}(i)
	}
	wg.Wait()
}
