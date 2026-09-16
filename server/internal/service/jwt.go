// Package service JWT令牌服务
package service

import (
	"MyBlog/internal/config"
	"MyBlog/internal/domain"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jtiRandomBytes 令牌唯一标识的随机字节数，Base64URL 编码后为 16 字符，
// 足以在令牌生命周期内避免实例碰撞，同时控制令牌体积。
const jtiRandomBytes = 12

// tokenFullSegmentCount 完整三段 JWT 的段数，用于区分完整令牌与 payload-only 输入。
const tokenFullSegmentCount = 3

// tokenPayloadSegmentIndex payload 段在完整三段 JWT 中的下标位置。
const tokenPayloadSegmentIndex = 1

// TokenType 令牌类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTClaims JWT声明 - 极简版，只保留绝对必需的字段
type JWTClaims struct {
	UserID    uint   `json:"u"`   // 进一步缩短字段名：uid -> u
	JTI       string `json:"jti"` // 令牌实例唯一标识，撤销键以此区分同一秒内签发的不同令牌
	ExpiresAt int64  `json:"exp"` // 直接使用Unix时间戳，不用jwt.NewNumericDate包装
}

// Valid 实现jwt.Claims interface
func (c JWTClaims) Valid() error {
	now := time.Now().Unix()
	if c.ExpiresAt < now {
		return fmt.Errorf("token已过期")
	}
	return nil
}

// GetExpirationTime 实现jwt.Claims接口
func (c JWTClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.ExpiresAt, 0)), nil
}

// GetIssuedAt 实现jwt.Claims接口
func (c JWTClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetNotBefore 实现jwt.Claims接口
func (c JWTClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetIssuer 实现jwt.Claims接口
func (c JWTClaims) GetIssuer() (string, error) {
	return "", nil
}

// GetSubject 实现jwt.Claims接口
func (c JWTClaims) GetSubject() (string, error) {
	return "", nil
}

// GetAudience 实现jwt.Claims接口
func (c JWTClaims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// TokenPair 令牌对 - 优化版本，只传输payload
type TokenPair struct {
	AccessToken  string `json:"access_token"`  // 只包含payload部分
	RefreshToken string `json:"refresh_token"` // 只包含payload部分
	ExpiresIn    int64  `json:"expires_in"`    // access token 的有效期，单位秒。
}

// 固定的JWT Header（Base64编码）
const fixedJWTHeader = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"

// JWTService JWT服务接口
type JWTService interface {
	GenerateTokenPair(user *domain.User) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*JWTClaims, error)
	ValidateRefreshToken(tokenString string) (*JWTClaims, error)
	RefreshAccessToken(refreshTokenString string) (*TokenPair, error)
	RevokeToken(tokenString string) error
	IsTokenRevoked(tokenString string) bool
	// 新增：从payload重构完整JWT
	ReconstructFullToken(payloadOnly string, tokenType TokenType) (string, error)
}

// jwtService JWT服务实现
type jwtService struct {
	config *config.Config
	// revokedTokens 以归一化撤销键记录已撤销令牌，值为令牌的过期时间。
	// 过期令牌本身无法通过签名校验，对应撤销键随之失去存在意义，
	// 撤销写入时惰性清理，防止撤销表随历史登出无界增长。
	revokedTokens map[string]time.Time
	// mu 保护 revokedTokens 的并发读写，避免多请求同时撤销与校验时产生数据竞争。
	mu sync.RWMutex
}

// NewJWTService 创建JWT服务实例
func NewJWTService(cfg *config.Config) JWTService {
	return &jwtService{
		config:        cfg,
		revokedTokens: make(map[string]time.Time),
	}
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (j *jwtService) GenerateTokenPair(user *domain.User) (*TokenPair, error) {
	now := time.Now()

	// 生成访问令牌
	accessToken, err := j.generateToken(user, AccessToken, now,
		time.Duration(j.config.JWT.AccessExpire)*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err := j.generateToken(user, RefreshToken, now,
		time.Duration(j.config.JWT.RefreshExpire)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.config.JWT.AccessExpire * 60), // 转换为秒
	}, nil
}

// generateToken 生成指定类型的令牌 - 优化版：只返回payload部分
func (j *jwtService) generateToken(user *domain.User, tokenType TokenType,
	issuedAt time.Time, duration time.Duration) (string, error) {

	tokenID, err := newTokenJTI()
	if err != nil {
		return "", fmt.Errorf("生成令牌唯一标识失败: %w", err)
	}

	claims := JWTClaims{
		UserID:    user.ID,
		JTI:       tokenID,
		ExpiresAt: issuedAt.Add(duration).Unix(),
	}

	// 直接序列化payload为JSON，然后Base64编码
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("序列化claims失败: %w", err)
	}

	// 返回Base64编码的payload（去掉padding）
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return payloadBase64, nil
}

// newTokenJTI 生成令牌实例唯一标识，随机源不可用时返回错误，
// 唯一性是撤销键正确性的前提，失败时令牌签发必须整体失败。
func newTokenJTI() (string, error) {
	randomBytes := make([]byte, jtiRandomBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("读取随机源失败: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

// ReconstructFullToken 从payload重构完整的JWT token
func (j *jwtService) ReconstructFullToken(payloadOnly string, tokenType TokenType) (string, error) {
	var secretKey []byte
	switch tokenType {
	case AccessToken:
		secretKey = []byte(j.config.JWT.AccessSecret)
	case RefreshToken:
		secretKey = []byte(j.config.JWT.RefreshSecret)
	default:
		return "", fmt.Errorf("不支持的令牌类型: %s", tokenType)
	}

	// 重构完整的JWT：header.payload.signature
	headerAndPayload := fixedJWTHeader + "." + payloadOnly

	// 计算签名
	signature, err := j.calculateSignature(headerAndPayload, secretKey)
	if err != nil {
		return "", fmt.Errorf("计算签名失败: %w", err)
	}

	return headerAndPayload + "." + signature, nil
}

// calculateSignature 计算JWT签名
func (j *jwtService) calculateSignature(data string, secretKey []byte) (string, error) {
	// 使用HMAC-SHA256计算签名
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	signature := h.Sum(nil)

	// 返回Base64 URL编码的签名（无padding）
	return base64.RawURLEncoding.EncodeToString(signature), nil
}

// ValidateAccessToken 验证访问令牌 - 支持payload-only格式
func (j *jwtService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	// 如果不包含点号，说明是payload-only格式，需要重构完整JWT
	if !strings.Contains(tokenString, ".") {
		fullToken, err := j.ReconstructFullToken(tokenString, AccessToken)
		if err != nil {
			return nil, fmt.Errorf("重构完整token失败: %w", err)
		}
		tokenString = fullToken
	}
	return j.validateToken(tokenString, []byte(j.config.JWT.AccessSecret))
}

// ValidateRefreshToken 验证刷新令牌 - 支持payload-only格式
func (j *jwtService) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	// 如果不包含点号，说明是payload-only格式，需要重构完整JWT
	if !strings.Contains(tokenString, ".") {
		fullToken, err := j.ReconstructFullToken(tokenString, RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("重构完整token失败: %w", err)
		}
		tokenString = fullToken
	}
	return j.validateToken(tokenString, []byte(j.config.JWT.RefreshSecret))
}

// validateToken 验证令牌的通用方法
func (j *jwtService) validateToken(tokenString string, secretKey []byte) (*JWTClaims, error) {
	// 检查令牌是否已被撤销
	if j.IsTokenRevoked(tokenString) {
		return nil, fmt.Errorf("令牌已被撤销")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("令牌解析失败: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}

// RefreshAccessToken 使用刷新令牌生成新的访问令牌
func (j *jwtService) RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	// 验证刷新令牌
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌验证失败: %w", err)
	}

	// 只需要UserID来生成新token，不需要其他用户信息
	user := &domain.User{
		ID: claims.UserID,
	}

	// 生成新的令牌对
	tokenPair, err := j.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("生成新令牌对失败: %w", err)
	}

	// 撤销旧的刷新令牌，内存实现无失败路径，显式忽略错误返回值
	_ = j.RevokeToken(refreshTokenString)

	return tokenPair, nil
}

// tokenRevocationKey 归一化令牌的撤销键，撤销侧与校验侧共用本函数保证键一致。
// 本系统签发的是 payload-only 串，但校验侧拿到的是重构后的完整三段 JWT，
// 键统一取 payload 段，完整 JWT 取中段，其余输入原样返回。
func tokenRevocationKey(tokenString string) string {
	parts := strings.Split(tokenString, ".")
	if len(parts) == tokenFullSegmentCount {
		return parts[tokenPayloadSegmentIndex]
	}
	return tokenString
}

// tokenExpiry 解析令牌 payload 的过期时间作为撤销记录的生命周期终点。
// payload 无法解析时回退为当前时间加刷新令牌有效期，保证撤销键至少
// 覆盖最长令牌生命周期后才被清理，不产生撤销保护空窗。
func (j *jwtService) tokenExpiry(payload string) time.Time {
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err == nil {
		var claims JWTClaims
		if jsonErr := json.Unmarshal(decoded, &claims); jsonErr == nil && claims.ExpiresAt > 0 {
			return time.Unix(claims.ExpiresAt, 0)
		}
	}
	return time.Now().Add(time.Duration(j.config.JWT.RefreshExpire) * time.Hour)
}

// RevokeToken 撤销令牌，撤销键经归一化后落表，同时惰性清理已过期的撤销记录。
func (j *jwtService) RevokeToken(tokenString string) error {
	key := tokenRevocationKey(tokenString)
	if key == "" {
		return nil
	}

	expiry := j.tokenExpiry(key)

	j.mu.Lock()
	defer j.mu.Unlock()

	j.revokedTokens[key] = expiry
	j.pruneExpiredRevocationsLocked(time.Now())
	return nil
}

// IsTokenRevoked 检查令牌是否已被撤销，过期撤销记录对应的令牌自身也已过期，
// 签名校验会拒绝该令牌，撤销状态随之自然失效。
func (j *jwtService) IsTokenRevoked(tokenString string) bool {
	key := tokenRevocationKey(tokenString)
	if key == "" {
		return false
	}

	j.mu.RLock()
	defer j.mu.RUnlock()

	expiry, revoked := j.revokedTokens[key]
	return revoked && expiry.After(time.Now())
}

// pruneExpiredRevocationsLocked 移除生命周期已终结的撤销记录，约束撤销表规模。
// 调用方必须已持有写锁，随撤销写入触发，无需独立清扫协程。
func (j *jwtService) pruneExpiredRevocationsLocked(now time.Time) {
	for key, expiry := range j.revokedTokens {
		if !expiry.After(now) {
			delete(j.revokedTokens, key)
		}
	}
}
