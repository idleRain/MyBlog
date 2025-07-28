package model

import (
	"time"

	"MyBlog/pkg/datetime"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID              uint              `json:"id" gorm:"primaryKey;comment:用户ID"`
	Username        string            `json:"username" gorm:"uniqueIndex;not null;size:50;comment:用户名，全局唯一"`
	Email           string            `json:"email" gorm:"uniqueIndex;not null;size:100;comment:邮箱地址，全局唯一"`
	Password        string            `json:"-" gorm:"not null;size:255;comment:密码（bcrypt加密）"`
	Nickname        string            `json:"nickname" gorm:"size:50;comment:用户昵称"`
	Avatar          string            `json:"avatar" gorm:"size:255;comment:头像URL"`
	Bio             string            `json:"bio" gorm:"type:text;comment:个人简介"`
	Birthday        datetime.JSONDate `json:"birthday" gorm:"type:date;comment:生日"`
	Role            string            `json:"role" gorm:"default:user;size:20;comment:用户角色 superadmin/admin/editor/user"`
	Status          int               `json:"status" gorm:"default:1;comment:状态 1-正常 0-禁用"`
	LastLoginAt     *time.Time        `json:"lastLoginAt" gorm:"type:datetime(3);comment:最后登录时间"`
	LastLoginIP     string            `json:"lastLoginIP" gorm:"size:45;comment:最后登录IP"`
	LoginCount      uint              `json:"loginCount" gorm:"default:0;comment:登录次数"`
	EmailVerifiedAt *time.Time        `json:"emailVerifiedAt" gorm:"type:datetime(3);comment:邮箱验证时间"`
	CreatedAt       time.Time         `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt       time.Time         `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
	DeletedAt       gorm.DeletedAt    `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Articles   []Article      `json:"-" gorm:"foreignKey:AuthorID"`
	Comments   []Comment      `json:"-" gorm:"foreignKey:UserID"`
	MediaFiles []MediaFile    `json:"-" gorm:"foreignKey:UploaderID"`
	Sessions   []UserSession  `json:"-" gorm:"foreignKey:UserID"`
	Activities []UserActivity `json:"-" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserSession 用户会话模型
type UserSession struct {
	ID              uint       `json:"id" gorm:"primaryKey;comment:会话ID"`
	UserID          uint       `json:"userId" gorm:"not null;index;comment:用户ID"`
	RefreshToken    string     `json:"refreshToken" gorm:"uniqueIndex;not null;size:255;comment:刷新令牌"`
	AccessTokenHash string     `json:"accessTokenHash" gorm:"size:64;comment:访问令牌哈希值"`
	DeviceInfo      string     `json:"deviceInfo" gorm:"type:json;comment:设备信息（浏览器、操作系统等）"`
	IPAddress       string     `json:"ipAddress" gorm:"size:45;index;comment:登录IP地址"`
	UserAgent       string     `json:"userAgent" gorm:"type:text;comment:用户代理字符串"`
	ExpiresAt       time.Time  `json:"expiresAt" gorm:"type:datetime(3);index;comment:令牌过期时间"`
	LastUsedAt      *time.Time `json:"lastUsedAt" gorm:"type:datetime(3);comment:最后使用时间"`
	IsActive        bool       `json:"isActive" gorm:"default:true;index;comment:会话状态：1-活跃，0-已注销"`
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
	UserID       *uint     `json:"userId" gorm:"index;comment:用户ID"`
	Action       string    `json:"action" gorm:"not null;size:50;index;comment:操作类型"`
	ResourceType string    `json:"resourceType" gorm:"size:50;comment:资源类型（article、comment等）"`
	ResourceID   *uint     `json:"resourceId" gorm:"comment:资源ID"`
	Description  string    `json:"description" gorm:"size:255;comment:操作描述"`
	Metadata     string    `json:"metadata" gorm:"type:json;comment:额外元数据"`
	IPAddress    string    `json:"ipAddress" gorm:"size:45;index;comment:IP地址"`
	UserAgent    string    `json:"userAgent" gorm:"type:text;comment:用户代理"`
	CreatedAt    time.Time `json:"createdAt" gorm:"type:datetime(3);index;comment:创建时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// TableName 指定表名
func (UserActivity) TableName() string {
	return "user_activities"
}

// 定义用户角色常量
type UserRole string

const (
	RoleUser       UserRole = "user"       // 普通用户
	RoleEditor     UserRole = "editor"     // 编辑者
	RoleAdmin      UserRole = "admin"      // 管理员
	RoleSuperAdmin UserRole = "superadmin" // 超级管理员
)

// 定义用户状态常量
const (
	UserStatusInactive = 0 // 禁用
	UserStatusActive   = 1 // 正常
)

// IsActive 检查用户是否为活跃状态
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// HasRole 检查用户是否具有指定角色
func (u *User) HasRole(role UserRole) bool {
	return UserRole(u.Role) == role
}

// IsAdmin 检查用户是否为管理员级别（admin 或 superadmin）
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin) || u.HasRole(RoleSuperAdmin)
}

// IsSuperAdmin 检查用户是否为超级管理员
func (u *User) IsSuperAdmin() bool {
	return u.HasRole(RoleSuperAdmin)
}

// CanEdit 检查用户是否具有编辑权限（editor 及以上）
func (u *User) CanEdit() bool {
	role := UserRole(u.Role)
	return role == RoleEditor || role == RoleAdmin || role == RoleSuperAdmin
}

// GetRoleLevel 获取角色权限级别
func (u *User) GetRoleLevel() int {
	switch UserRole(u.Role) {
	case RoleSuperAdmin:
		return 4
	case RoleAdmin:
		return 3
	case RoleEditor:
		return 2
	case RoleUser:
		return 1
	default:
		return 0
	}
}
