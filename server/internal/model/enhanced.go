package model

import (
	"encoding/json"
	"time"
)

// ArticleCategory 文章分类关联模型（支持文章多分类）
type ArticleCategory struct {
	ID         uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	ArticleID  uint      `json:"articleId" gorm:"not null;index;comment:文章ID"`
	CategoryID uint      `json:"categoryId" gorm:"not null;index;comment:分类ID"`
	CreatedAt  time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`

	// 关联关系
	Article  Article  `json:"-" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE"`
	Category Category `json:"-" gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (ArticleCategory) TableName() string {
	return "article_categories"
}

// OperationLog 操作日志模型（用于安全审计）
type OperationLog struct {
	ID           uint            `json:"id" gorm:"primaryKey;comment:日志ID"`
	UserID       *uint           `json:"userId" gorm:"index;comment:操作用户ID"`
	Action       string          `json:"action" gorm:"not null;size:100;index;comment:操作类型"`
	ResourceType *string         `json:"resourceType" gorm:"size:50;index;comment:资源类型"`
	ResourceID   *uint           `json:"resourceId" gorm:"comment:资源ID"`
	IPAddress    *string         `json:"ipAddress" gorm:"size:45;comment:IP地址"`
	UserAgent    *string         `json:"userAgent" gorm:"type:text;comment:用户代理"`
	Details      json.RawMessage `json:"details" gorm:"type:json;comment:操作详情"`
	CreatedAt    time.Time       `json:"createdAt" gorm:"type:datetime(3);index;comment:创建时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// TableName 指定表名
func (OperationLog) TableName() string {
	return "operation_logs"
}

// 操作类型常量
const (
	ActionLogin         = "login"          // 用户登录
	ActionLogout        = "logout"         // 用户登出
	ActionCreateUser    = "create_user"    // 创建用户
	ActionUpdateUser    = "update_user"    // 更新用户
	ActionDeleteUser    = "delete_user"    // 删除用户
	ActionCreateRole    = "create_role"    // 分配角色
	ActionCreateArticle = "create_article" // 创建文章
	ActionUpdateArticle = "update_article" // 更新文章
	ActionDeleteArticle = "delete_article" // 删除文章
	ActionDeleteComment = "delete_comment" // 删除评论
	ActionSystemConfig  = "system_config"  // 系统配置
)

// ArticleLike 文章点赞模型
type ArticleLike struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:点赞ID"`
	ArticleID uint      `json:"articleId" gorm:"not null;index;comment:文章ID"`
	UserID    uint      `json:"userId" gorm:"not null;index;comment:用户ID"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:datetime(3);index;comment:点赞时间"`

	// 关联关系
	Article Article `json:"-" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE"`
	User    User    `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (ArticleLike) TableName() string {
	return "article_likes"
}

// CommentLike 评论点赞模型
type CommentLike struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:点赞ID"`
	CommentID uint      `json:"commentId" gorm:"not null;index;comment:评论ID"`
	UserID    uint      `json:"userId" gorm:"not null;index;comment:用户ID"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:datetime(3);index;comment:点赞时间"`

	// 关联关系
	Comment Comment `json:"-" gorm:"foreignKey:CommentID;constraint:OnDelete:CASCADE"`
	User    User    `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (CommentLike) TableName() string {
	return "comment_likes"
}

// ArticleBookmark 文章收藏模型
type ArticleBookmark struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:收藏ID"`
	ArticleID uint      `json:"articleId" gorm:"not null;index;comment:文章ID"`
	UserID    uint      `json:"userId" gorm:"not null;index;comment:用户ID"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:datetime(3);index;comment:收藏时间"`

	// 关联关系
	Article Article `json:"-" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE"`
	User    User    `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (ArticleBookmark) TableName() string {
	return "article_bookmarks"
}

// Notification 系统通知模型
type Notification struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:通知ID"`
	UserID      uint      `json:"userId" gorm:"not null;index;comment:接收用户ID"`
	Type        string    `json:"type" gorm:"not null;size:50;index;comment:通知类型"`
	Title       string    `json:"title" gorm:"not null;size:255;comment:通知标题"`
	Content     *string   `json:"content" gorm:"type:text;comment:通知内容"`
	RelatedType *string   `json:"relatedType" gorm:"size:50;index;comment:关联资源类型"`
	RelatedID   *uint     `json:"relatedId" gorm:"comment:关联资源ID"`
	IsRead      bool      `json:"isRead" gorm:"default:false;index;comment:是否已读"`
	CreatedAt   time.Time `json:"createdAt" gorm:"type:datetime(3);index;comment:创建时间"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}

// 通知类型常量
const (
	NotificationTypeCommentReply = "comment_reply" // 评论回复
	NotificationTypeArticleLike  = "article_like"  // 文章点赞
	NotificationTypeCommentLike  = "comment_like"  // 评论点赞
	NotificationTypeSystem       = "system"        // 系统通知
	NotificationTypeFollow       = "follow"        // 用户关注
	NotificationTypeArticleNew   = "article_new"   // 新文章发布
)

// MarkAsRead 标记通知为已读
func (n *Notification) MarkAsRead() {
	n.IsRead = true
	n.UpdatedAt = time.Now()
}

// SearchLog 搜索记录模型
type SearchLog struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:搜索记录ID"`
	UserID       *uint     `json:"userId" gorm:"index;comment:搜索用户ID"`
	Keyword      string    `json:"keyword" gorm:"not null;size:255;index;comment:搜索关键词"`
	ResultsCount int       `json:"resultsCount" gorm:"default:0;comment:搜索结果数量"`
	IPAddress    *string   `json:"ipAddress" gorm:"size:45;index;comment:IP地址"`
	UserAgent    *string   `json:"userAgent" gorm:"type:text;comment:用户代理"`
	CreatedAt    time.Time `json:"createdAt" gorm:"type:datetime(3);index;comment:搜索时间"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// TableName 指定表名
func (SearchLog) TableName() string {
	return "search_logs"
}

// ContentStats 内容统计模型（用于性能优化和热门推荐）
type ContentStats struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:统计ID"`
	ContentType string    `json:"contentType" gorm:"not null;size:50;index;comment:内容类型"`
	ContentID   uint      `json:"contentId" gorm:"not null;index;comment:内容ID"`
	StatType    string    `json:"statType" gorm:"not null;size:50;index;comment:统计类型"`
	StatValue   uint      `json:"statValue" gorm:"default:0;index;comment:统计值"`
	StatDate    time.Time `json:"statDate" gorm:"type:date;index;comment:统计日期"`
	CreatedAt   time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
}

// TableName 指定表名
func (ContentStats) TableName() string {
	return "content_stats"
}

// 内容类型和统计类型常量
const (
	ContentTypeArticle  = "article"  // 文章
	ContentTypeTag      = "tag"      // 标签
	ContentTypeCategory = "category" // 分类

	StatTypeDailyViews   = "daily_views"   // 日浏览量
	StatTypeWeeklyViews  = "weekly_views"  // 周浏览量
	StatTypeMonthlyViews = "monthly_views" // 月浏览量
	StatTypeLikesCount   = "likes_count"   // 点赞数
)

// UserFollow 用户关注关系模型
type UserFollow struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:关注关系ID"`
	FollowerID  uint      `json:"followerId" gorm:"not null;index;comment:关注者ID"`
	FollowingID uint      `json:"followingId" gorm:"not null;index;comment:被关注者ID"`
	CreatedAt   time.Time `json:"createdAt" gorm:"type:datetime(3);index;comment:关注时间"`

	// 关联关系
	Follower  User `json:"-" gorm:"foreignKey:FollowerID;constraint:OnDelete:CASCADE"`
	Following User `json:"-" gorm:"foreignKey:FollowingID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (UserFollow) TableName() string {
	return "user_follows"
}

// IsValidFollow 检查关注关系是否有效（不能自己关注自己）
func (uf *UserFollow) IsValidFollow() bool {
	return uf.FollowerID != uf.FollowingID
}
