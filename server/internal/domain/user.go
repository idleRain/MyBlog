// Package domain 全系统唯一类型语言层。
// 本包承载领域实体与行为方法，是后端各层的公共类型来源，
// 避免此前 repository 与 model 双 User 各自演进导致的类型倒挂。
package domain

import (
	"time"

	"MyBlog/pkg/datetime"

	"gorm.io/gorm"
)

// User 用户领域实体，同时承担 GORM 持久化实体与领域对象职责。
// 合并自原 model.User 与 repository.User，字段集以 model.User 为准，
// 保证登录与鉴权路径能够访问 IsLocked 等安全字段。
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
	LockedUntil       *time.Time        `json:"-" gorm:"type:datetime(3);comment:账户锁定截止时间，到期后可重新登录"`
	PasswordChangedAt *time.Time        `json:"-" gorm:"type:datetime(3);comment:密码最后修改时间，用于安全审计"`
	LastLoginAt       *time.Time        `json:"-" gorm:"type:datetime(3);index;comment:最后登录时间"`
	LastLoginIP       string            `json:"-" gorm:"size:45;comment:最后登录IP，IPv6 最长 45 字符"`
	LoginCount        uint              `json:"-" gorm:"default:0;comment:累计登录成功次数"`
	EmailVerifiedAt   *time.Time        `json:"-" gorm:"type:datetime(3);comment:邮箱验证完成时间，为空表示未验证"`
	Remark            string            `json:"-" gorm:"size:500;comment:管理员备注，仅管理端可见"`
	CreatedAt         time.Time         `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt         time.Time         `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
	DeletedAt         gorm.DeletedAt    `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定 users 表名，供 GORM 持久化映射使用。
func (User) TableName() string {
	return "users"
}

// IsActive 检查用户是否为活跃状态。
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

// HasRole 检查用户是否具有指定角色。
func (u *User) HasRole(role UserRole) bool {
	return UserRole(u.Role) == role
}

// IsAdmin 检查用户是否为管理员级别，包括 admin 与 superadmin。
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin) || u.HasRole(RoleSuperAdmin)
}

// IsSuperAdmin 检查用户是否为超级管理员。
func (u *User) IsSuperAdmin() bool {
	return u.HasRole(RoleSuperAdmin)
}

// CanEdit 检查用户是否具备编辑权限，即 editor 及以上角色。
func (u *User) CanEdit() bool {
	role := UserRole(u.Role)
	return role == RoleEditor || role == RoleAdmin || role == RoleSuperAdmin
}

// GetRoleLevel 获取角色权限级别，数值越大权限越高。
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
