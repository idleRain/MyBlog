package model

import (
	"time"
)

// ArticleRevision 文章修订历史模型，保存每次正文保存后的快照，支持版本回滚与差异对比。
type ArticleRevision struct {
	ID            uint      `json:"id" gorm:"primaryKey;comment:修订ID"`
	ArticleID     uint      `json:"articleId" gorm:"not null;index;uniqueIndex:uk_article_revision,priority:1;comment:文章ID"`
	RevisionNo    uint      `json:"revisionNo" gorm:"not null;uniqueIndex:uk_article_revision,priority:2;comment:修订版本号，从 1 开始随文章 version 递增"`
	Title         string    `json:"title" gorm:"size:200;comment:该版本文章标题快照"`
	Summary       string    `json:"summary" gorm:"type:text;comment:该版本摘要快照"`
	Content       string    `json:"content" gorm:"type:longtext;not null;comment:该版本正文快照，Markdown 格式"`
	ContentHTML   string    `json:"contentHtml" gorm:"type:longtext;comment:该版本渲染后 HTML 快照"`
	WordCount     uint      `json:"wordCount" gorm:"default:0;comment:该版本字数统计"`
	ChangeSummary string    `json:"changeSummary" gorm:"size:255;comment:变更说明，由作者填写"`
	EditorID      *uint     `json:"editorId" gorm:"index;comment:执行本次修订的用户ID"`
	IsAutosave    bool      `json:"isAutosave" gorm:"default:false;index;comment:是否编辑器自动保存产生的快照"`
	CreatedAt     time.Time `json:"createdAt" gorm:"type:datetime(3);index;comment:修订时间"`

	// 关联关系
	Article Article `json:"-" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE"`
	Editor  *User   `json:"editor,omitempty" gorm:"foreignKey:EditorID;constraint:OnDelete:SET NULL"`
}

// TableName 指定表名
func (ArticleRevision) TableName() string {
	return "article_revisions"
}
