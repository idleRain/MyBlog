package model

import "time"

// 内容多语言翻译模型集合。
// 主表列恒为缺省语言内容，翻译行以 locale 区分其他语言；
// 翻译行允许按字段部分填写，输出时缺失翻译的字段回退主列。

// ArticleTranslation 文章翻译模型，复合唯一索引确保同一文章同一语言仅一行翻译。
type ArticleTranslation struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:翻译ID"`
	ArticleID      uint      `json:"articleId" gorm:"not null;uniqueIndex:uk_article_translation,priority:1;comment:文章ID"`
	Locale         string    `json:"locale" gorm:"not null;size:10;uniqueIndex:uk_article_translation,priority:2;index;comment:语言标识，主子标签形式，如 zh、en"`
	Title          string    `json:"title" gorm:"size:200;comment:该语言文章标题"`
	Summary        string    `json:"summary" gorm:"type:text;comment:该语言文章摘要"`
	Content        string    `json:"content" gorm:"type:longtext;comment:该语言文章正文，Markdown 格式"`
	ContentHTML    string    `json:"contentHtml" gorm:"type:longtext;comment:该语言正文渲染后的 HTML 缓存"`
	WordCount      uint      `json:"wordCount" gorm:"default:0;comment:该语言正文字数统计"`
	SEOTitle       string    `json:"seoTitle" gorm:"size:100;comment:该语言SEO标题"`
	SEODescription string    `json:"seoDescription" gorm:"size:255;comment:该语言SEO描述"`
	SEOKeywords    string    `json:"seoKeywords" gorm:"size:200;comment:该语言SEO关键词"`
	CreatedAt      time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:最后更新时间"`

	// 关联关系
	Article Article `json:"-" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (ArticleTranslation) TableName() string {
	return "article_translations"
}

// CategoryTranslation 分类翻译模型，复合唯一索引确保同一分类同一语言仅一行翻译。
type CategoryTranslation struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:翻译ID"`
	CategoryID     uint      `json:"categoryId" gorm:"not null;uniqueIndex:uk_category_translation,priority:1;comment:分类ID"`
	Locale         string    `json:"locale" gorm:"not null;size:10;uniqueIndex:uk_category_translation,priority:2;index;comment:语言标识，主子标签形式，如 zh、en"`
	Name           string    `json:"name" gorm:"size:50;comment:该语言分类名称"`
	Description    string    `json:"description" gorm:"type:text;comment:该语言分类描述"`
	SEOTitle       string    `json:"seoTitle" gorm:"size:100;comment:该语言SEO标题"`
	SEODescription string    `json:"seoDescription" gorm:"size:255;comment:该语言SEO描述"`
	CreatedAt      time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:最后更新时间"`

	// 关联关系
	Category Category `json:"-" gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (CategoryTranslation) TableName() string {
	return "category_translations"
}

// TagTranslation 标签翻译模型，复合唯一索引确保同一标签同一语言仅一行翻译。
type TagTranslation struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:翻译ID"`
	TagID       uint      `json:"tagId" gorm:"not null;uniqueIndex:uk_tag_translation,priority:1;comment:标签ID"`
	Locale      string    `json:"locale" gorm:"not null;size:10;uniqueIndex:uk_tag_translation,priority:2;index;comment:语言标识，主子标签形式，如 zh、en"`
	Name        string    `json:"name" gorm:"size:30;comment:该语言标签名称"`
	Description string    `json:"description" gorm:"size:200;comment:该语言标签描述"`
	CreatedAt   time.Time `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"type:datetime(3);comment:最后更新时间"`

	// 关联关系
	Tag Tag `json:"-" gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (TagTranslation) TableName() string {
	return "tag_translations"
}
