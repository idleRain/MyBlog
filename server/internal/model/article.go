package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 文章分类模型，采用 parent_id、root_id、level 与 path 描述树形结构。
type Category struct {
	ID             uint           `json:"id" gorm:"primaryKey;comment:分类ID"`
	Name           string         `json:"name" gorm:"not null;size:50;comment:分类名称"`
	Slug           string         `json:"slug" gorm:"uniqueIndex;not null;size:50;comment:URL友好标识"`
	Description    string         `json:"description" gorm:"type:text;comment:分类描述"`
	CoverImage     string         `json:"coverImage" gorm:"size:255;comment:分类封面图"`
	ParentID       *uint          `json:"parentId" gorm:"index;comment:父分类ID，顶级分类为空"`
	RootID         *uint          `json:"rootId" gorm:"index;comment:根分类ID，用于整棵子树的聚合查询"`
	Level          uint8          `json:"level" gorm:"default:1;index;comment:分类层级，顶级为 1"`
	Path           string         `json:"path" gorm:"size:100;comment:物化路径，形如 /1/5/12，用于一次查询取整棵子树"`
	SortOrder      int            `json:"sortOrder" gorm:"default:0;index;comment:排序权重，数值小的靠前"`
	Status         int            `json:"status" gorm:"type:tinyint;default:1;index;comment:分类状态：1-显示 0-隐藏"`
	ArticleCount   uint           `json:"articleCount" gorm:"default:0;comment:文章数量，发布时异步维护"`
	IsFeatured     bool           `json:"isFeatured" gorm:"default:false;index;comment:是否为精选分类"`
	SEOTitle       string         `json:"seoTitle" gorm:"size:100;comment:SEO标题"`
	SEODescription string         `json:"seoDescription" gorm:"size:255;comment:SEO描述"`
	CreatedAt      time.Time      `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt      time.Time      `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID;constraint:OnDelete:SET NULL"`
	Children []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Articles []Article  `json:"-" gorm:"foreignKey:CategoryID"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}

// 定义分类状态常量
const (
	CategoryStatusHidden = 0 // 隐藏
	CategoryStatusShown  = 1 // 显示
)

// IsVisible 检查分类是否对外展示。
func (c *Category) IsVisible() bool {
	return c.Status == CategoryStatusShown
}

// Tag 文章标签模型
type Tag struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:标签ID"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null;size:30;comment:标签名称"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null;size:30;comment:URL友好标识"`
	Color       string    `json:"color" gorm:"size:7;default:#808080;comment:标签颜色，HEX 格式"`
	Description string    `json:"description" gorm:"size:200;comment:标签描述"`
	Status      int       `json:"status" gorm:"type:tinyint;default:1;index;comment:标签状态：1-启用 0-隐藏"`
	UsageCount  uint      `json:"usageCount" gorm:"default:0;index;comment:使用次数，文章挂载时异步维护"`
	IsHot       bool      `json:"isHot" gorm:"default:false;index;comment:是否热门标签"`
	CreatedAt   time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`

	// 关联关系
	Articles []Article `json:"-" gorm:"many2many:article_tags"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

// 定义标签状态常量
const (
	TagStatusHidden  = 0 // 隐藏
	TagStatusEnabled = 1 // 启用
)

// IsEnabled 检查标签是否处于启用状态。
func (t *Tag) IsEnabled() bool {
	return t.Status == TagStatusEnabled
}

// Article 文章模型
type Article struct {
	ID             uint           `json:"id" gorm:"primaryKey;comment:文章ID"`
	Title          string         `json:"title" gorm:"not null;size:200;comment:文章标题"`
	Slug           string         `json:"slug" gorm:"uniqueIndex;not null;size:200;comment:URL友好标识"`
	Summary        string         `json:"summary" gorm:"type:text;comment:文章摘要"`
	Content        string         `json:"content" gorm:"type:longtext;not null;comment:文章内容，Markdown 格式"`
	ContentHTML    string         `json:"contentHtml" gorm:"type:longtext;comment:文章内容，渲染后的 HTML 缓存"`
	CoverImage     string         `json:"coverImage" gorm:"size:500;comment:封面图片URL"`
	AuthorID       uint           `json:"authorId" gorm:"not null;index;index:idx_author_status,priority:1;comment:作者ID"`
	CategoryID     *uint          `json:"categoryId" gorm:"index;comment:主分类ID"`
	Status         ArticleStatus  `json:"status" gorm:"default:draft;size:20;index;index:idx_status_published,priority:1;index:idx_author_status,priority:2;comment:文章状态：draft/published/archived/private"`
	OriginType     string         `json:"originType" gorm:"size:20;default:original;index;comment:来源类型：original-原创 translation-翻译 reprint-转载"`
	SourceURL      string         `json:"sourceUrl" gorm:"size:500;comment:原文链接，原创文章为空"`
	SourceAuthor   string         `json:"sourceAuthor" gorm:"size:50;comment:原文作者，原创文章为空"`
	AccessPassword string         `json:"-" gorm:"size:255;comment:访问密码哈希，仅私密文章生效，为空表示仅登录可见"`
	IsFeatured     bool           `json:"isFeatured" gorm:"default:false;index;comment:是否精选文章"`
	IsTop          bool           `json:"isTop" gorm:"default:false;index;comment:是否置顶"`
	CommentEnabled bool           `json:"commentEnabled" gorm:"default:true;comment:是否允许评论"`
	ViewCount      uint           `json:"viewCount" gorm:"default:0;index;comment:浏览量"`
	LikeCount      uint           `json:"likeCount" gorm:"default:0;comment:点赞数"`
	BookmarkCount  uint           `json:"bookmarkCount" gorm:"default:0;comment:收藏数"`
	CommentCount   uint           `json:"commentCount" gorm:"default:0;comment:评论数"`
	WordCount      uint           `json:"wordCount" gorm:"default:0;comment:字数统计"`
	ReadingTime    uint           `json:"readingTime" gorm:"default:0;comment:预计阅读时间，单位分钟"`
	Version        uint           `json:"version" gorm:"default:1;comment:内容版本号，每次保存正文递增，对应 article_revisions"`
	SEOTitle       string         `json:"seoTitle" gorm:"size:100;comment:SEO标题"`
	SEODescription string         `json:"seoDescription" gorm:"size:255;comment:SEO描述"`
	SEOKeywords    string         `json:"seoKeywords" gorm:"size:200;comment:SEO关键词"`
	ScheduledAt    *time.Time     `json:"scheduledAt" gorm:"type:datetime(3);index;comment:定时发布时间，到期后由调度任务发布"`
	PublishedAt    *time.Time     `json:"publishedAt" gorm:"type:datetime(3);index;index:idx_status_published,priority:2;comment:发布时间"`
	EditedAt       *time.Time     `json:"editedAt" gorm:"type:datetime(3);comment:正文最后编辑时间，用于展示已编辑标记"`
	ArchivedAt     *time.Time     `json:"archivedAt" gorm:"type:datetime(3);comment:归档时间，进入归档状态时写入"`
	LastCommentAt  *time.Time     `json:"lastCommentAt" gorm:"type:datetime(3);comment:最新评论时间，用于评论排序展示"`
	CreatedAt      time.Time      `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt      time.Time      `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Author     User              `json:"author" gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE"`
	Category   *Category         `json:"category,omitempty" gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL"`
	Categories []Category        `json:"categories,omitempty" gorm:"many2many:article_categories"`
	Tags       []Tag             `json:"tags,omitempty" gorm:"many2many:article_tags"`
	Comments   []Comment         `json:"-" gorm:"foreignKey:ArticleID"`
	Views      []ArticleView     `json:"-" gorm:"foreignKey:ArticleID"`
	Likes      []ArticleLike     `json:"-" gorm:"foreignKey:ArticleID"`
	Bookmarks  []ArticleBookmark `json:"-" gorm:"foreignKey:ArticleID"`
}

// TableName 指定表名
func (Article) TableName() string {
	return "articles"
}

// ArticleTag 文章标签关联模型，复合唯一索引确保同一文章不重复挂载同一标签。
type ArticleTag struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	ArticleID uint      `json:"articleId" gorm:"not null;uniqueIndex:uk_article_tag,priority:1;comment:文章ID"`
	TagID     uint      `json:"tagId" gorm:"not null;uniqueIndex:uk_article_tag,priority:2;comment:标签ID"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`

	// 关联关系
	Article Article `json:"-" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE"`
	Tag     Tag     `json:"-" gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (ArticleTag) TableName() string {
	return "article_tags"
}

// ArticleView 文章浏览统计模型，按文章、访客与日期三元组去重计数。
type ArticleView struct {
	ID              uint      `json:"id" gorm:"primaryKey;comment:浏览记录ID"`
	ArticleID       uint      `json:"articleId" gorm:"not null;uniqueIndex:uk_article_visitor_date,priority:1;comment:文章ID"`
	UserID          *uint     `json:"userId" gorm:"index;comment:用户ID，注册用户填写"`
	VisitorID       string    `json:"visitorId" gorm:"size:64;uniqueIndex:uk_article_visitor_date,priority:2;comment:访客标识，匿名用户填写"`
	IPAddress       string    `json:"ipAddress" gorm:"size:45;index;comment:IP地址"`
	UserAgent       string    `json:"userAgent" gorm:"type:text;comment:用户代理"`
	Referer         string    `json:"referer" gorm:"size:500;comment:来源页面"`
	ViewDate        time.Time `json:"viewDate" gorm:"type:date;uniqueIndex:uk_article_visitor_date,priority:3;comment:浏览日期"`
	ViewCount       uint      `json:"viewCount" gorm:"default:1;comment:当日浏览次数"`
	DurationSeconds uint      `json:"durationSeconds" gorm:"default:0;comment:页面停留时长，单位秒，由前端埋点上报"`
	CreatedAt       time.Time `json:"createdAt" gorm:"type:datetime(3);comment:首次浏览时间"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:最后浏览时间"`

	// 关联关系
	Article Article `json:"-" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE"`
	User    *User   `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// TableName 指定表名
func (ArticleView) TableName() string {
	return "article_views"
}

// 定义文章状态枚举
type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"     // 草稿
	ArticleStatusPublished ArticleStatus = "published" // 已发布
	ArticleStatusArchived  ArticleStatus = "archived"  // 已归档
	ArticleStatusPrivate   ArticleStatus = "private"   // 私有
)

// 定义文章来源类型常量
const (
	ArticleOriginOriginal    = "original"    // 原创
	ArticleOriginTranslation = "translation" // 翻译
	ArticleOriginReprint     = "reprint"     // 转载
)

// IsPublished 检查文章是否已发布
func (a *Article) IsPublished() bool {
	return a.Status == ArticleStatusPublished && a.PublishedAt != nil
}

// IsPublic 检查文章是否公开可访问
func (a *Article) IsPublic() bool {
	return a.Status == ArticleStatusPublished
}

// IsScheduled 检查文章是否处于待发布的定时状态。
func (a *Article) IsScheduled() bool {
	return a.ScheduledAt != nil && a.ScheduledAt.After(time.Now()) && a.Status == ArticleStatusDraft
}

// IsReprint 检查文章是否为转载或翻译内容，此类文章必须补充原文链接。
func (a *Article) IsReprint() bool {
	return a.OriginType == ArticleOriginReprint || a.OriginType == ArticleOriginTranslation
}

// CanComment 检查文章是否允许评论
func (a *Article) CanComment() bool {
	return a.CommentEnabled && a.IsPublished()
}

// GetURL 获取文章的URL路径
func (a *Article) GetURL() string {
	return "/articles/" + a.Slug
}

// CalculateReadingTime 计算阅读时间（基于字数，平均每分钟200字）
func (a *Article) CalculateReadingTime() uint {
	if a.WordCount == 0 {
		return 1
	}
	readingTime := a.WordCount / 200
	if readingTime == 0 {
		return 1
	}
	return readingTime
}

// IncrementViewCount 增加浏览量
func (a *Article) IncrementViewCount() {
	a.ViewCount++
}

// IncrementCommentCount 增加评论数
func (a *Article) IncrementCommentCount() {
	a.CommentCount++
}

// DecrementCommentCount 减少评论数
func (a *Article) DecrementCommentCount() {
	if a.CommentCount > 0 {
		a.CommentCount--
	}
}

// IncrementVersion 递增内容版本号，保存正文快照后调用。
func (a *Article) IncrementVersion() {
	a.Version++
}
