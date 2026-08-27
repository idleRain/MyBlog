package model

import (
	"time"

	"MyBlog/pkg/datetime"

	"gorm.io/gorm"
)

// User 用户模型，承载账号身份、个人资料与安全状态三类信息。
type User struct {
	ID                uint              `json:"id" gorm:"primaryKey;comment:用户ID"`
	Username          string            `json:"username" gorm:"uniqueIndex;not null;size:50;comment:用户名，全局唯一"`
	Email             string            `json:"email" gorm:"uniqueIndex;not null;size:100;comment:邮箱地址，全局唯一"`
	Phone             *string           `json:"phone,omitempty" gorm:"uniqueIndex;size:20;comment:手机号，全局唯一，未绑定时为空"`
	Password          string            `json:"-" gorm:"not null;size:255;comment:密码，存储 bcrypt 哈希值"`
	Nickname          string            `json:"nickname" gorm:"size:50;comment:用户昵称，为空时展示用户名"`
	Avatar            string            `json:"avatar" gorm:"size:255;comment:头像URL"`
	CoverImage        string            `json:"coverImage" gorm:"size:500;comment:个人主页封面图URL"`
	Bio               string            `json:"bio" gorm:"type:text;comment:个人简介"`
	Website           string            `json:"website" gorm:"size:255;comment:个人网站URL"`
	Location          string            `json:"location" gorm:"size:100;comment:常居地描述"`
	Gender            *int              `json:"gender,omitempty" gorm:"type:tinyint;comment:性别：0-未知 1-男 2-女"`
	Birthday          datetime.JSONDate `json:"birthday" gorm:"type:date;comment:生日"`
	Timezone          string            `json:"timezone" gorm:"size:50;default:Asia/Shanghai;comment:用户时区标识，IANA 命名格式"`
	Locale            string            `json:"locale" gorm:"size:10;default:zh-CN;comment:用户界面语言标识，BCP 47 格式"`
	Role              string            `json:"role" gorm:"size:20;index;default:user;comment:用户角色：superadmin/admin/editor/user"`
	Status            int               `json:"status" gorm:"type:tinyint;index;default:1;comment:用户状态：1-正常 0-禁用 2-锁定"`
	FailedLoginCount  uint              `json:"-" gorm:"default:0;comment:连续登录失败次数，登录成功后清零"`
	LockedUntil       *time.Time        `json:"lockedUntil" gorm:"type:datetime(3);comment:账户锁定截止时间，到期后可重新登录"`
	PasswordChangedAt *time.Time        `json:"passwordChangedAt" gorm:"type:datetime(3);comment:密码最后修改时间，用于安全审计"`
	LastLoginAt       *time.Time        `json:"lastLoginAt" gorm:"type:datetime(3);index;comment:最后登录时间"`
	LastLoginIP       string            `json:"lastLoginIP" gorm:"size:45;comment:最后登录IP，IPv6 最长 45 字符"`
	LoginCount        uint              `json:"loginCount" gorm:"default:0;comment:累计登录成功次数"`
	EmailVerifiedAt   *time.Time        `json:"emailVerifiedAt" gorm:"type:datetime(3);comment:邮箱验证完成时间，为空表示未验证"`
	Remark            string            `json:"-" gorm:"size:500;comment:管理员备注，仅管理端可见"`
	CreatedAt         time.Time         `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt         time.Time         `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
	DeletedAt         gorm.DeletedAt    `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Articles         []Article         `json:"-" gorm:"foreignKey:AuthorID"`
	Comments         []Comment         `json:"-" gorm:"foreignKey:UserID"`
	MediaFiles       []MediaFile       `json:"-" gorm:"foreignKey:UploaderID"`
	Sessions         []UserSession     `json:"-" gorm:"foreignKey:UserID"`
	Activities       []UserActivity    `json:"-" gorm:"foreignKey:UserID"`
	ArticleLikes     []ArticleLike     `json:"-" gorm:"foreignKey:UserID"`
	CommentLikes     []CommentLike     `json:"-" gorm:"foreignKey:UserID"`
	ArticleBookmarks []ArticleBookmark `json:"-" gorm:"foreignKey:UserID"`
	Notifications    []Notification    `json:"-" gorm:"foreignKey:UserID"`
	SearchLogs       []SearchLog       `json:"-" gorm:"foreignKey:UserID"`
	Followers        []UserFollow      `json:"-" gorm:"foreignKey:FollowingID"`
	Following        []UserFollow      `json:"-" gorm:"foreignKey:FollowerID"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

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
	UserStatusLocked   = 2 // 锁定
)

// 定义用户性别常量
const (
	UserGenderUnknown = 0 // 未知
	UserGenderMale    = 1 // 男
	UserGenderFemale  = 2 // 女
)

// IsActive 检查用户是否为活跃状态
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsLocked 检查用户是否处于锁定状态，锁定状态在到期后自动解除。
func (u *User) IsLocked() bool {
	if u.Status == UserStatusLocked {
		return u.LockedUntil == nil || u.LockedUntil.After(time.Now())
	}
	return u.LockedUntil != nil && u.LockedUntil.After(time.Now())
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
