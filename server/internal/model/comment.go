package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 评论模型
type Comment struct {
	ID            uint           `json:"id" gorm:"primaryKey;comment:评论ID"`
	ArticleID     uint           `json:"articleId" gorm:"not null;index;comment:文章ID"`
	UserID        *uint          `json:"userId" gorm:"index;comment:用户ID（注册用户）"`
	ParentID      *uint          `json:"parentId" gorm:"index;comment:父评论ID（回复功能）"`
	RootID        *uint          `json:"rootId" gorm:"index;comment:根评论ID（便于查询评论树）"`
	Level         uint8          `json:"level" gorm:"default:1;comment:评论层级"`
	AuthorName    string         `json:"authorName" gorm:"size:50;comment:游客姓名"`
	AuthorEmail   string         `json:"authorEmail" gorm:"size:100;comment:游客邮箱"`
	AuthorWebsite string         `json:"authorWebsite" gorm:"size:255;comment:游客网站"`
	AuthorIP      string         `json:"authorIP" gorm:"size:45;index;comment:评论者IP地址"`
	Content       string         `json:"content" gorm:"type:text;not null;comment:评论内容"`
	ContentHTML   string         `json:"contentHtml" gorm:"type:text;comment:评论内容（HTML格式，缓存用）"`
	Status        CommentStatus  `json:"status" gorm:"default:pending;index;comment:审核状态"`
	LikeCount     uint           `json:"likeCount" gorm:"default:0;comment:点赞数"`
	ReplyCount    uint           `json:"replyCount" gorm:"default:0;comment:回复数量"`
	UserAgent     string         `json:"userAgent" gorm:"type:text;comment:用户代理"`
	IsAuthor      bool           `json:"isAuthor" gorm:"default:false;comment:是否为文章作者回复"`
	IsPinned      bool           `json:"isPinned" gorm:"default:false;comment:是否置顶评论"`
	CreatedAt     time.Time      `json:"createdAt" gorm:"type:datetime(3);index;comment:创建时间"`
	UpdatedAt     time.Time      `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Article  Article   `json:"article" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE"`
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
	Parent   *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
	Root     *Comment  `json:"root,omitempty" gorm:"foreignKey:RootID;constraint:OnDelete:CASCADE"`
	Children []Comment `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}

// 定义评论状态枚举
type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"  // 待审核
	CommentStatusApproved CommentStatus = "approved" // 已通过
	CommentStatusRejected CommentStatus = "rejected" // 已拒绝
	CommentStatusSpam     CommentStatus = "spam"     // 垃圾评论
	CommentStatusTrash    CommentStatus = "trash"    // 回收站
)

// IsApproved 检查评论是否已通过审核
func (c *Comment) IsApproved() bool {
	return c.Status == CommentStatusApproved
}

// IsVisible 检查评论是否对外可见
func (c *Comment) IsVisible() bool {
	return c.Status == CommentStatusApproved && c.DeletedAt.Time.IsZero()
}

// IsReply 检查是否为回复评论
func (c *Comment) IsReply() bool {
	return c.ParentID != nil && *c.ParentID > 0
}

// IsRootComment 检查是否为根评论
func (c *Comment) IsRootComment() bool {
	return c.ParentID == nil || *c.ParentID == 0
}

// GetAuthorName 获取评论者名称（注册用户优先使用昵称）
func (c *Comment) GetAuthorName() string {
	if c.User != nil && c.User.Nickname != "" {
		return c.User.Nickname
	}
	if c.User != nil {
		return c.User.Username
	}
	return c.AuthorName
}

// GetAuthorAvatar 获取评论者头像
func (c *Comment) GetAuthorAvatar() string {
	if c.User != nil && c.User.Avatar != "" {
		return c.User.Avatar
	}
	// 返回默认头像或基于邮箱生成的Gravatar
	return ""
}

// CanEdit 检查是否可以编辑（作者本人或管理员）
func (c *Comment) CanEdit(currentUser *User) bool {
	if currentUser == nil {
		return false
	}

	// 管理员可以编辑所有评论
	if currentUser.IsAdmin() {
		return true
	}

	// 评论者本人可以编辑
	if c.UserID != nil && *c.UserID == currentUser.ID {
		return true
	}

	return false
}

// CanDelete 检查是否可以删除
func (c *Comment) CanDelete(currentUser *User) bool {
	if currentUser == nil {
		return false
	}

	// 管理员可以删除所有评论
	if currentUser.IsAdmin() {
		return true
	}

	// 评论者本人可以删除
	if c.UserID != nil && *c.UserID == currentUser.ID {
		return true
	}

	return false
}

// IncrementReplyCount 增加回复数量
func (c *Comment) IncrementReplyCount() {
	c.ReplyCount++
}

// DecrementReplyCount 减少回复数量
func (c *Comment) DecrementReplyCount() {
	if c.ReplyCount > 0 {
		c.ReplyCount--
	}
}

// SetAsAuthorReply 标记为作者回复
func (c *Comment) SetAsAuthorReply() {
	c.IsAuthor = true
}

// Approve 通过审核
func (c *Comment) Approve() {
	c.Status = CommentStatusApproved
}

// Reject 拒绝评论
func (c *Comment) Reject() {
	c.Status = CommentStatusRejected
}

// MarkAsSpam 标记为垃圾评论
func (c *Comment) MarkAsSpam() {
	c.Status = CommentStatusSpam
}

// MoveToTrash 移至回收站
func (c *Comment) MoveToTrash() {
	c.Status = CommentStatusTrash
}

// Pin 置顶评论
func (c *Comment) Pin() {
	c.IsPinned = true
}

// Unpin 取消置顶
func (c *Comment) Unpin() {
	c.IsPinned = false
}
