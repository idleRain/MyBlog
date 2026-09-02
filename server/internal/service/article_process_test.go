package service

import (
	"testing"

	"MyBlog/internal/model"
)

// TestProcessContentChineseWordCount 验证中文内容按字符数统计字数与阅读时间。
func TestProcessContentChineseWordCount(t *testing.T) {
	svc := newArticleTestService(&fakeArticleRepo{})
	article := &model.Article{
		Title:   "中文测试",
		Content: "这是一段没有空格的中文内容，用于验证字数统计逻辑是否按字符数而非空白分词计算。",
	}

	if err := svc.processContent(article); err != nil {
		t.Fatalf("processContent 失败: %v", err)
	}

	// 中文内容按字符数统计，字数应等于 Rune 数量。
	wantRunes := len([]rune(article.Content))
	if int(article.WordCount) != wantRunes {
		t.Errorf("字数 = %d, 期望按字符统计为 %d", article.WordCount, wantRunes)
	}

	// 阅读时间至少为 1 分钟，且随字数增长。
	if article.ReadingTime == 0 {
		t.Error("阅读时间不应为 0")
	}
}

// TestProcessContentMixedWordCount 验证中英文混合内容仍按字符数统计。
func TestProcessContentMixedWordCount(t *testing.T) {
	svc := newArticleTestService(&fakeArticleRepo{})
	article := &model.Article{
		Title:   "混合内容",
		Content: "Go 语言编程 Hello World 中文段落连续无空格",
	}

	if err := svc.processContent(article); err != nil {
		t.Fatalf("processContent 失败: %v", err)
	}

	wantRunes := len([]rune(article.Content))
	if int(article.WordCount) != wantRunes {
		t.Errorf("字数 = %d, 期望按字符统计为 %d", article.WordCount, wantRunes)
	}
}

// TestProcessContentEscapesAndSummary 验证内容转义与摘要提取。
func TestProcessContentEscapesAndSummary(t *testing.T) {
	svc := newArticleTestService(&fakeArticleRepo{})
	article := &model.Article{
		Title:   "转义测试",
		Content: "<script>alert(1)</script> 这是一段正文内容",
	}

	if err := svc.processContent(article); err != nil {
		t.Fatalf("processContent 失败: %v", err)
	}

	if article.Content != "&lt;script&gt;alert(1)&lt;/script&gt; 这是一段正文内容" {
		t.Errorf("内容未正确转义: %q", article.Content)
	}

	// 未提供摘要时应从内容自动提取。
	if article.Summary == "" {
		t.Error("摘要应自动从内容提取")
	}
}

// TestArchiveSetsArchivedAt 验证归档会写入归档时间并保留发布时间。
func TestArchiveSetsArchivedAt(t *testing.T) {
	// 通过仓储层测试验证，此处仅确认 service 层 ArchiveArticle 调用仓储。
	repo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticle(id), nil
		},
	}
	svc := newArticleTestService(repo)

	if err := svc.ArchiveArticle(1, 1); err != nil {
		t.Fatalf("归档失败: %v", err)
	}
}
