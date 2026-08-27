package model

import (
	"time"
)

// AuthToken 认证令牌模型，支撑密码找回与邮箱验证等带时效的一次性流程。
// 令牌原文不下库，仅存储 SHA256 哈希值，防止拖库后伪造请求。
type AuthToken struct {
	ID        uint       `json:"id" gorm:"primaryKey;comment:令牌ID"`
	UserID    uint       `json:"userId" gorm:"not null;index;comment:所属用户ID"`
	TokenHash string     `json:"-" gorm:"uniqueIndex;not null;size:64;comment:令牌哈希值，SHA256 十六进制"`
	TokenType TokenType  `json:"tokenType" gorm:"size:30;default:password_reset;index;comment:令牌用途：password_reset-密码重置 email_verify-邮箱验证"`
	ExpiresAt time.Time  `json:"expiresAt" gorm:"type:datetime(3);index;comment:令牌过期时间"`
	UsedAt    *time.Time `json:"usedAt" gorm:"type:datetime(3);comment:令牌核销时间，为空表示未使用"`
	RequestIP string     `json:"requestIP" gorm:"size:45;comment:申请令牌时的来源IP，用于安全审计"`
	CreatedAt time.Time  `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`

	// 关联关系
	User User `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (AuthToken) TableName() string {
	return "auth_tokens"
}

// 定义令牌用途枚举
type TokenType string

const (
	TokenTypePasswordReset TokenType = "password_reset" // 密码重置
	TokenTypeEmailVerify   TokenType = "email_verify"   // 邮箱验证
)

// IsUsable 检查令牌是否可用，需同时满足未核销且未过期。
func (t *AuthToken) IsUsable() bool {
	return t.UsedAt == nil && t.ExpiresAt.After(time.Now())
}

// MarkUsed 核销令牌，一次性令牌使用后立即失效。
func (t *AuthToken) MarkUsed() {
	now := time.Now()
	t.UsedAt = &now
}
