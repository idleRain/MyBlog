package model

import (
	"time"

	"MyBlog/internal/domain"
)

// User 用户模型，已合并至 domain.User，此处保留类型别名以兼容既有引用与 GORM 关联。
type User = domain.User

// 用户角色、状态与性别常量统一引用 domain 定义，避免双份维护。
type UserRole = domain.UserRole

const (
	RoleUser       = domain.RoleUser
	RoleEditor     = domain.RoleEditor
	RoleAdmin      = domain.RoleAdmin
	RoleSuperAdmin = domain.RoleSuperAdmin
)

const (
	UserStatusInactive = domain.UserStatusInactive
	UserStatusActive   = domain.UserStatusActive
	UserStatusLocked   = domain.UserStatusLocked
)

const (
	UserGenderUnknown = domain.UserGenderUnknown
	UserGenderMale    = domain.UserGenderMale
	UserGenderFemale  = domain.UserGenderFemale
)

// UserSession 用户会话模型
type UserSession struct {
	ID              uint       `json:"id" gorm:"primaryKey;comment:会话ID"`
	UserID          uint       `json:"userId" gorm:"not null;index;index:idx_user_active,priority:1;comment:用户ID"`
	RefreshToken    string     `json:"refreshToken" gorm:"uniqueIndex;not null;size:255;comment:刷新令牌"`
	AccessTokenHash string     `json:"accessTokenHash" gorm:"size:64;comment:访问令牌哈希值"`
	DeviceInfo      string     `json:"deviceInfo" gorm:"type:json;comment:设备信息，包含浏览器与操作系统等"`
	DeviceType      string     `json:"deviceType" gorm:"size:20;index;default:web;comment:设备类型：web/mobile/tablet/desktop"`
	IPAddress       string     `json:"ipAddress" gorm:"size:45;index;comment:登录IP地址"`
	UserAgent       string     `json:"userAgent" gorm:"type:text;comment:用户代理字符串"`
	ExpiresAt       time.Time  `json:"expiresAt" gorm:"type:datetime(3);index;comment:令牌过期时间"`
	LastUsedAt      *time.Time `json:"lastUsedAt" gorm:"type:datetime(3);comment:最后使用时间"`
	LastRefreshAt   *time.Time `json:"lastRefreshAt" gorm:"type:datetime(3);comment:刷新令牌最近轮换时间"`
	LogoutAt        *time.Time `json:"logoutAt" gorm:"type:datetime(3);comment:会话注销时间"`
	IsActive        bool       `json:"isActive" gorm:"default:true;index;index:idx_user_active,priority:2;comment:会话状态：1-活跃，0-已注销"`
	CreatedAt       time.Time  `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt       time.Time  `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (UserSession) TableName() string {
	return "user_sessions"
}

// UserActivity 用户活动日志模型
type UserActivity struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:活动ID"`
	UserID       *uint     `json:"userId" gorm:"index;index:idx_user_created,priority:1;comment:用户ID"`
	Action       string    `json:"action" gorm:"not null;size:50;index;comment:操作类型"`
	ResourceType string    `json:"resourceType" gorm:"size:50;comment:资源类型：article/comment/user 等"`
	ResourceID   *uint     `json:"resourceId" gorm:"comment:资源ID"`
	Description  string    `json:"description" gorm:"size:255;comment:操作描述"`
	Status       string    `json:"status" gorm:"size:20;default:success;comment:执行结果：success-成功 failed-失败"`
	ErrorMessage string    `json:"errorMessage" gorm:"size:500;comment:失败原因，成功时为空"`
	DurationMs   uint      `json:"durationMs" gorm:"default:0;comment:操作耗时，单位毫秒"`
	Metadata     string    `json:"metadata" gorm:"type:json;comment:额外元数据"`
	IPAddress    string    `json:"ipAddress" gorm:"size:45;index;comment:IP地址"`
	UserAgent    string    `json:"userAgent" gorm:"type:text;comment:用户代理"`
	CreatedAt    time.Time `json:"createdAt" gorm:"type:datetime(3);index;index:idx_user_created,priority:2;comment:创建时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// TableName 指定表名
func (UserActivity) TableName() string {
	return "user_activities"
}
