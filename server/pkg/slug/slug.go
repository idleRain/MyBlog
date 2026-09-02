// Package slug 提供 URL 友好标识生成工具，供文章、分类、标签等模块复用。
package slug

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

// Generate 从标题生成 slug，仅保留小写字母、数字与连字符。
// 中文标题过滤后可能为空，此时回退为 <前缀>-<纳秒时间戳>-<随机数> 确保非空且唯一。
func Generate(prefix, title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// 移除特殊字符，仅保留字母、数字与连字符。
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// 结果为空时回退为带时间戳与随机数的标识，避免唯一索引写入空值。
	if result.Len() == 0 {
		return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), rand.IntN(1_000_000))
	}

	return result.String()
}
