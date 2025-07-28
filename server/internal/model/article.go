package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 文章分类模型
type Category struct {
	ID             uint           `json:"id" gorm:"primaryKey;comment:分类ID"`
	Name           string         `json:"name" gorm:"not null;size:50;comment:分类名称"`
	Slug           string         `json:"slug" gorm:"uniqueIndex;not null;size:50;comment:URL友好标识"`
	Description    string         `json:"description" gorm:"type:text;comment:分类描述"`
	CoverImage     string         `json:"coverImage" gorm:"size:255;comment:分类封面图"`
	ParentID       *uint          `json:"parentId" gorm:"index;comment:父分类ID"`
	Level          uint8          `json:"level" gorm:"default:1;index;comment:分类层级"`
	SortOrder      int            `json:"sortOrder" gorm:"default:0;index;comment:排序权重"`
	ArticleCount   uint           `json:"articleCount" gorm:"default:0;comment:文章数量"`
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

// Tag 文章标签模型
type Tag struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:标签ID"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null;size:30;comment:标签名称"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null;size:30;comment:URL友好标识"`
	Color       string    `json:"color" gorm:"default:#808080;size:7;comment:标签颜色（HEX格式）"`
	Description string    `json:"description" gorm:"size:200;comment:标签描述"`
	UsageCount  uint      `json:"usageCount" gorm:"default:0;index;comment:使用次数"`
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

// Article 文章模型
type Article struct {
	ID             uint           `json:"id" gorm:"primaryKey;comment:文章ID"`
	Title          string         `json:"title" gorm:"not null;size:200;comment:文章标题"`
	Slug           string         `json:"slug" gorm:"uniqueIndex;not null;size:200;comment:URL友好标识"`
	Summary        string         `json:"summary" gorm:"type:text;comment:文章摘要"`
	Content        string         `json:"content" gorm:"type:longtext;not null;comment:文章内容（Markdown格式）"`
	ContentHTML    string         `json:"contentHtml" gorm:"type:longtext;comment:文章内容（HTML格式，缓存用）"`
	CoverImage     string         `json:"coverImage" gorm:"size:500;comment:封面图片URL"`
	AuthorID       uint           `json:"authorId" gorm:"not null;index;comment:作者ID"`
	CategoryID     *uint          `json:"categoryId" gorm:"index;comment:主分类ID"`
	Status         ArticleStatus  `json:"status" gorm:"default:draft;index;comment:文章状态"`
	IsFeatured     bool           `json:"isFeatured" gorm:"default:false;index;comment:是否精选文章"`
	IsTop          bool           `json:"isTop" gorm:"default:false;index;comment:是否置顶"`
	CommentEnabled bool           `json:"commentEnabled" gorm:"default:true;comment:是否允许评论"`
	ViewCount      uint           `json:"viewCount" gorm:"default:0;index;comment:浏览量"`
	LikeCount      uint           `json:"likeCount" gorm:"default:0;comment:点赞数"`
	CommentCount   uint           `json:"commentCount" gorm:"default:0;comment:评论数"`
	WordCount      uint           `json:"wordCount" gorm:"default:0;comment:字数统计"`
	ReadingTime    uint           `json:"readingTime" gorm:"default:0;comment:预计阅读时间（分钟）"`
	SEOTitle       string         `json:"seoTitle" gorm:"size:100;comment:SEO标题"`
	SEODescription string         `json:"seoDescription" gorm:"size:255;comment:SEO描述"`
	SEOKeywords    string         `json:"seoKeywords" gorm:"size:200;comment:SEO关键词"`
	PublishedAt    *time.Time     `json:"publishedAt" gorm:"type:datetime(3);index;comment:发布时间"`
	CreatedAt      time.Time      `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt      time.Time      `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Author   User          `json:"author" gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE"`
	Category *Category     `json:"category,omitempty" gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL"`
	Tags     []Tag         `json:"tags,omitempty" gorm:"many2many:article_tags"`
	Comments []Comment     `json:"-" gorm:"foreignKey:ArticleID"`
	Views    []ArticleView `json:"-" gorm:"foreignKey:ArticleID"`
}

// TableName 指定表名
func (Article) TableName() string {
	return "articles"
}

// ArticleTag 文章标签关联模型
type ArticleTag struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:关联ID"`
	ArticleID uint      `json:"articleId" gorm:"not null;comment:文章ID"`
	TagID     uint      `json:"tagId" gorm:"not null;comment:标签ID"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`

	// 关联关系
	Article Article `json:"-" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE"`
	Tag     Tag     `json:"-" gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (ArticleTag) TableName() string {
	return "article_tags"
}

// ArticleView 文章浏览统计模型
type ArticleView struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:浏览记录ID"`
	ArticleID uint      `json:"articleId" gorm:"not null;comment:文章ID"`
	UserID    *uint     `json:"userId" gorm:"index;comment:用户ID（注册用户）"`
	VisitorID string    `json:"visitorId" gorm:"size:64;comment:访客标识（匿名用户）"`
	IPAddress string    `json:"ipAddress" gorm:"size:45;index;comment:IP地址"`
	UserAgent string    `json:"userAgent" gorm:"type:text;comment:用户代理"`
	Referer   string    `json:"referer" gorm:"size:500;comment:来源页面"`
	ViewDate  time.Time `json:"viewDate" gorm:"type:date;index;comment:浏览日期"`
	ViewCount uint      `json:"viewCount" gorm:"default:1;comment:当日浏览次数"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:datetime(3);comment:首次浏览时间"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:最后浏览时间"`

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

// IsPublished 检查文章是否已发布
func (a *Article) IsPublished() bool {
	return a.Status == ArticleStatusPublished && a.PublishedAt != nil
}

// IsPublic 检查文章是否公开可访问
func (a *Article) IsPublic() bool {
	return a.Status == ArticleStatusPublished
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
