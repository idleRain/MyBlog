// Package service JWT令牌服务
package service

import (
	"MyBlog/internal/config"
	"MyBlog/internal/repository"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType 令牌类型
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTClaims JWT声明 - 极简版，只保留绝对必需的字段
type JWTClaims struct {
	UserID    uint  `json:"u"`   // 进一步缩短字段名：uid -> u
	ExpiresAt int64 `json:"exp"` // 直接使用Unix时间戳，不用jwt.NewNumericDate包装
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
	ExpiresIn    int64  `json:"expires_in"`    // access token过期时间（秒）
}

// 固定的JWT Header（Base64编码）
const fixedJWTHeader = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"

// JWTService JWT服务接口
type JWTService interface {
	GenerateTokenPair(user *repository.User) (*TokenPair, error)
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
	// TODO: 实现token撤销存储（可以用MongoDB或MySQL数据库）
	revokedTokens map[string]time.Time // 简单的内存存储，生产环境应使用MongoDB
}

// NewJWTService 创建JWT服务实例
func NewJWTService(cfg *config.Config) JWTService {
	return &jwtService{
		config:        cfg,
		revokedTokens: make(map[string]time.Time),
	}
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (j *jwtService) GenerateTokenPair(user *repository.User) (*TokenPair, error) {
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
func (j *jwtService) generateToken(user *repository.User, tokenType TokenType,
	issuedAt time.Time, duration time.Duration) (string, error) {

	claims := JWTClaims{
		UserID:    user.ID,
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
	user := &repository.User{
		ID: claims.UserID,
	}

	// 生成新的令牌对
	tokenPair, err := j.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("生成新令牌对失败: %w", err)
	}

	// 撤销旧的刷新令牌
	j.RevokeToken(refreshTokenString)

	return tokenPair, nil
}

// RevokeToken 撤销令牌
func (j *jwtService) RevokeToken(tokenString string) error {
	// 简单的内存存储实现，生产环境应使用MongoDB等持久化存储
	j.revokedTokens[tokenString] = time.Now()

	// TODO: 定期清理过期的撤销令牌，避免内存泄漏

	return nil
}

// IsTokenRevoked 检查令牌是否已被撤销
func (j *jwtService) IsTokenRevoked(tokenString string) bool {
	_, revoked := j.revokedTokens[tokenString]
	return revoked
}

// 向后兼容的函数，保持现有代码正常工作

// ValidateToken 验证JWT令牌（向后兼容）
// Deprecated: 使用 ValidateAccessToken 替代
func ValidateToken(tokenString string) (*JWTClaims, error) {
	// 为了向后兼容，这里需要创建一个临时的服务实例
	// 在实际使用中，应该使用依赖注入传递JWTService实例
	cfg := config.Get()
	service := NewJWTService(cfg)
	return service.ValidateAccessToken(tokenString)
}
