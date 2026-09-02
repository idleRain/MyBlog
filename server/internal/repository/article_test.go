package repository

import (
	"strings"
	"testing"
)

// TestGenerateSlug 验证仓储层委托公共 slug 工具生成文章 slug。
func TestGenerateSlug(t *testing.T) {
	// 英文标题保留字母与连字符。
	englishSlug := generateSlug("Hello World!")
	if englishSlug != "hello-world" {
		t.Errorf("英文标题 slug = %q, 期望 hello-world", englishSlug)
	}

	// 中文标题回退为带 article 前缀的标识。
	chineseSlug := generateSlug("中文标题测试")
	if !strings.HasPrefix(chineseSlug, "article-") {
		t.Errorf("中文标题回退 slug 应以 article- 开头，实际为 %q", chineseSlug)
	}
}
