// Package markdown 提供 Markdown 到 HTML 的服务端渲染能力。
package markdown

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// defaultRenderer 全站默认渲染器：启用 GFM 扩展与自动标题锚点，
// 保持 Unsafe 关闭，原始 HTML 不写入输出以阻断脚本注入。
var defaultRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

// Render 将 Markdown 源文本渲染为可安全挂载的 HTML 片段。
func Render(source string) (string, error) {
	var buffer bytes.Buffer
	if err := defaultRenderer.Convert([]byte(source), &buffer); err != nil {
		return "", fmt.Errorf("渲染 Markdown 失败: %w", err)
	}
	return buffer.String(), nil
}
