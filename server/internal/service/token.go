// Package service 令牌服务
package service

import (
	"MyBlog/internal/config"
	"MyBlog/internal/domain"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// tokenRandomBytes 令牌随机部分的字节数。
// 16 字节即 128 位随机量，十六进制编码后为 32 个字符，
// 在令牌生命周期内碰撞概率可忽略，同时保持令牌紧凑易读。
const tokenRandomBytes = 16

// secondsPerMinute 分钟到秒的换算基数，用于把配置中的分钟有效期转换为秒。
const secondsPerMinute = 60

// TokenType 令牌类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// TokenIdentity 通过校验的令牌所承载的身份。
// 不透明令牌自身不携带任何数据，身份与生命周期完全由服务端令牌表给出，
// 本结构仅用于把校验结果传递给调用方。
type TokenIdentity struct {
	UserID    uint
	ExpiresAt time.Time
}

// TokenPair 令牌对，字段名即对外 API 契约。
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// TokenServiceInterface 令牌服务接口。
type TokenServiceInterface interface {
	GenerateTokenPair(user *domain.User) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*TokenIdentity, error)
	ValidateRefreshToken(tokenString string) (*TokenIdentity, error)
	RefreshAccessToken(refreshTokenString string) (*TokenPair, error)
	RevokeToken(tokenString string) error
	// RevokeUserTokens 撤销指定用户当前存活的全部令牌，用于改密等全局失效场景。
	RevokeUserTokens(userID uint) error
}

// tokenRecord 令牌表中单条存活令牌的记录。
type tokenRecord struct {
	userID    uint
	tokenType TokenType
	expiresAt time.Time
}

// tokenService 令牌服务实现，采用不透明令牌方案。
// 签发时生成密码学随机串，并把身份与生命周期登记在服务端令牌表；
// 校验与撤销均以该表为唯一权威，令牌本身不承载任何可被篡改的身份信息。
type tokenService struct {
	config *config.Config
	// tokens 以令牌串为键登记存活令牌，是校验与撤销的唯一依据。
	tokens map[string]tokenRecord
	// tokensByUser 按用户索引其存活令牌串，供按用户整体撤销使用。
	tokensByUser map[uint]map[string]struct{}
	// mu 保护令牌表与用户索引的并发读写。
	mu sync.RWMutex
}

// NewTokenService 创建令牌服务实例。
func NewTokenService(cfg *config.Config) TokenServiceInterface {
	return &tokenService{
		config:       cfg,
		tokens:       make(map[string]tokenRecord),
		tokensByUser: make(map[uint]map[string]struct{}),
	}
}

// GenerateTokenPair 生成访问令牌与刷新令牌对。
func (s *tokenService) GenerateTokenPair(user *domain.User) (*TokenPair, error) {
	now := time.Now()

	accessToken, err := s.issueToken(user.ID, AccessToken, now,
		time.Duration(s.config.JWT.AccessExpire)*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	refreshToken, err := s.issueToken(user.ID, RefreshToken, now,
		time.Duration(s.config.JWT.RefreshExpire)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWT.AccessExpire) * secondsPerMinute,
	}, nil
}

// issueToken 生成随机令牌串并登记到令牌表。
func (s *tokenService) issueToken(userID uint, tokenType TokenType, issuedAt time.Time,
	duration time.Duration) (string, error) {

	tokenString, err := newTokenString()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[tokenString] = tokenRecord{
		userID:    userID,
		tokenType: tokenType,
		expiresAt: issuedAt.Add(duration),
	}
	if s.tokensByUser[userID] == nil {
		s.tokensByUser[userID] = make(map[string]struct{})
	}
	s.tokensByUser[userID][tokenString] = struct{}{}

	s.pruneExpiredLocked(issuedAt)

	return tokenString, nil
}

// newTokenString 生成密码学安全的随机令牌串。
// 随机源不可用时必须让签发整体失败，否则会退化为可预测的令牌。
func newTokenString() (string, error) {
	randomBytes := make([]byte, tokenRandomBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("读取随机源失败: %w", err)
	}
	return hex.EncodeToString(randomBytes), nil
}

// ValidateAccessToken 校验访问令牌，令牌不存在、已过期或类型不符时返回错误。
func (s *tokenService) ValidateAccessToken(tokenString string) (*TokenIdentity, error) {
	return s.validateToken(tokenString, AccessToken)
}

// ValidateRefreshToken 校验刷新令牌，令牌不存在、已过期或类型不符时返回错误。
func (s *tokenService) ValidateRefreshToken(tokenString string) (*TokenIdentity, error) {
	return s.validateToken(tokenString, RefreshToken)
}

// validateToken 查表校验令牌并核对类型，访问令牌与刷新令牌不可互换使用。
func (s *tokenService) validateToken(tokenString string, tokenType TokenType) (*TokenIdentity, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("令牌为空")
	}

	s.mu.RLock()
	record, exists := s.tokens[tokenString]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("令牌无效或已失效")
	}
	if record.tokenType != tokenType {
		return nil, fmt.Errorf("令牌类型不符")
	}
	if !record.expiresAt.After(time.Now()) {
		return nil, fmt.Errorf("令牌已过期")
	}

	return &TokenIdentity{UserID: record.userID, ExpiresAt: record.expiresAt}, nil
}

// RefreshAccessToken 使用刷新令牌换取新令牌对，旧刷新令牌随即撤销。
func (s *tokenService) RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	identity, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌验证失败: %w", err)
	}

	tokenPair, err := s.GenerateTokenPair(&domain.User{ID: identity.UserID})
	if err != nil {
		return nil, fmt.Errorf("生成新令牌对失败: %w", err)
	}

	// 刷新即旋转：旧刷新令牌立即失效，防止同一刷新令牌被重复兑换。
	if err := s.RevokeToken(refreshTokenString); err != nil {
		return nil, fmt.Errorf("撤销旧刷新令牌失败: %w", err)
	}

	return tokenPair, nil
}

// RevokeToken 撤销单个令牌，令牌不在表中时视为已失效，不产生错误。
func (s *tokenService) RevokeToken(tokenString string) error {
	if tokenString == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.removeTokenLocked(tokenString)
	return nil
}

// RevokeUserTokens 撤销指定用户当前存活的全部令牌。
func (s *tokenService) RevokeUserTokens(userID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for tokenString := range s.tokensByUser[userID] {
		delete(s.tokens, tokenString)
	}
	delete(s.tokensByUser, userID)

	return nil
}

// removeTokenLocked 从令牌表与用户索引中移除令牌，调用方必须已持有写锁。
func (s *tokenService) removeTokenLocked(tokenString string) {
	record, exists := s.tokens[tokenString]
	if !exists {
		return
	}

	delete(s.tokens, tokenString)
	s.detachFromUserLocked(record.userID, tokenString)
}

// detachFromUserLocked 从用户名下摘除令牌串，该用户无存活令牌时一并删除索引项。
// 调用方必须已持有写锁。
func (s *tokenService) detachFromUserLocked(userID uint, tokenString string) {
	userTokens, exists := s.tokensByUser[userID]
	if !exists {
		return
	}

	delete(userTokens, tokenString)
	if len(userTokens) == 0 {
		delete(s.tokensByUser, userID)
	}
}

// pruneExpiredLocked 移除生命周期已终结的令牌记录，约束令牌表规模。
// 调用方必须已持有写锁，随签发触发，无需独立清扫协程。
func (s *tokenService) pruneExpiredLocked(now time.Time) {
	for tokenString, record := range s.tokens {
		if record.expiresAt.After(now) {
			continue
		}
		delete(s.tokens, tokenString)
		s.detachFromUserLocked(record.userID, tokenString)
	}
}
