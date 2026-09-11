package markdown

import (
	"strings"
	"testing"
)

// TestRenderBasicMarkdown 验证标题与段落渲染为基础 HTML 结构。
func TestRenderBasicMarkdown(t *testing.T) {
	output, err := Render("# 标题\n\n正文段落")
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}
	if !strings.Contains(output, "<h1") {
		t.Errorf("输出应包含标题元素: %q", output)
	}
	if !strings.Contains(output, "<p>正文段落</p>") {
		t.Errorf("输出应包含段落元素: %q", output)
	}
}

// TestRenderStripsRawScript 验证原始脚本标签不会进入渲染输出。
func TestRenderStripsRawScript(t *testing.T) {
	output, err := Render("<script>alert(1)</script>\n\n正文段落")
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}
	if strings.Contains(output, "<script") {
		t.Errorf("输出不应包含脚本标签: %q", output)
	}
}

// TestRenderTableExtension 验证 GFM 表格扩展可用。
func TestRenderTableExtension(t *testing.T) {
	source := "| a | b |\n| --- | --- |\n| 1 | 2 |"
	output, err := Render(source)
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}
	if !strings.Contains(output, "<table>") {
		t.Errorf("输出应包含表格元素: %q", output)
	}
}
