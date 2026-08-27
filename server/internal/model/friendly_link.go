package model

import (
	"time"

	"gorm.io/gorm"
)

// FriendlyLink 友情链接模型，管理互链申请、审核与展示的完整生命周期。
type FriendlyLink struct {
	ID           uint           `json:"id" gorm:"primaryKey;comment:链接ID"`
	Name         string         `json:"name" gorm:"not null;size:50;comment:站点名称"`
	URL          string         `json:"url" gorm:"uniqueIndex;not null;size:255;comment:站点URL，全局唯一防止重复收录"`
	Logo         string         `json:"logo" gorm:"size:500;comment:站点图标或头像URL"`
	Description  string         `json:"description" gorm:"size:255;comment:站点简介"`
	ContactEmail string         `json:"contactEmail" gorm:"size:100;comment:站长联系邮箱"`
	SortOrder    int            `json:"sortOrder" gorm:"default:0;index;comment:展示排序权重，数值小的靠前"`
	Status       LinkStatus     `json:"status" gorm:"size:20;default:pending;index;comment:链接状态：pending-待审核 active-展示中 hidden-已隐藏 rejected-已拒绝"`
	IsReciprocal bool           `json:"isReciprocal" gorm:"default:false;comment:是否已确认对方回链"`
	Note         string         `json:"note" gorm:"size:255;comment:管理员备注，如收录时间与沟通记录"`
	CreatedAt    time.Time      `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt    time.Time      `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名
func (FriendlyLink) TableName() string {
	return "friendly_links"
}

// 定义友链状态枚举
type LinkStatus string

const (
	LinkStatusPending  LinkStatus = "pending"  // 待审核
	LinkStatusActive   LinkStatus = "active"   // 展示中
	LinkStatusHidden   LinkStatus = "hidden"   // 已隐藏
	LinkStatusRejected LinkStatus = "rejected" // 已拒绝
)

// IsVisible 检查友链是否对外展示。
func (f *FriendlyLink) IsVisible() bool {
	return f.Status == LinkStatusActive
}

// Approve 审核通过并进入展示状态。
func (f *FriendlyLink) Approve() {
	f.Status = LinkStatusActive
}

// Hide 将友链下架但保留数据。
func (f *FriendlyLink) Hide() {
	f.Status = LinkStatusHidden
}

// Reject 拒绝该互链申请。
func (f *FriendlyLink) Reject() {
	f.Status = LinkStatusRejected
}
