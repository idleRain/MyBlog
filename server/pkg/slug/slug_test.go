package slug

import (
	"strings"
	"testing"
)

// TestGenerate 验证 slug 生成逻辑，确保返回结果非空且仅含合法字符。
func TestGenerate(t *testing.T) {
	testCases := []struct {
		name  string
		title string
	}{
		{
			name:  "英文标题保留字母与连字符",
			title: "Hello World!",
		},
		{
			name:  "纯中文标题回退为时间戳标识",
			title: "中文标题不传别名测试",
		},
		{
			name:  "符号标题回退为时间戳标识",
			title: "!!!@@@###",
		},
		{
			name:  "空标题回退为时间戳标识",
			title: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Generate("article", tc.title)
			if got == "" {
				t.Fatalf("Generate(%q, %q) 返回空字符串", "article", tc.title)
			}
			// 校验 slug 仅包含小写字母、数字与连字符，不残留非法字符。
			for _, r := range got {
				if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '-' {
					t.Errorf("Generate(%q) 包含非法字符 %q，结果为 %q", tc.title, r, got)
				}
			}
		})
	}
}

// TestGenerateFallbackPrefix 验证中文标题回退生成的 slug 使用指定前缀，便于前端识别。
func TestGenerateFallbackPrefix(t *testing.T) {
	got := Generate("article", "中文标题")
	if !strings.HasPrefix(got, "article-") {
		t.Errorf("中文标题回退 slug 应以 article- 开头，实际为 %q", got)
	}
}

// TestGenerateFallbackUnique 验证同一标题连续生成的 slug 各不相同，避免唯一索引冲突。
func TestGenerateFallbackUnique(t *testing.T) {
	first := Generate("article", "中文标题")
	second := Generate("article", "中文标题")
	if first == second {
		t.Errorf("同一标题两次生成的回退 slug 不应相同：%q", first)
	}
}
