package domain

import "strings"

// Language 内容语言标识，取值与前台 paraglide 的 locales 保持一致。
type Language string

// 语言标识常量，LanguageAll 供管理端一次读取全量翻译包，不属于语言白名单成员。
const (
	LanguageChinese Language = "zh"
	LanguageEnglish Language = "en"
	LanguageAll     Language = "*"
)

// DefaultLanguage 缺省语言，所有多语言字段的主列均以该语言存储。
const DefaultLanguage = LanguageChinese

// LanguageContextKey 语言标识在 gin 上下文中的存取键，由语言中间件写入、handler 层读取。
const LanguageContextKey = "language"

// ParseLanguage 解析 Accept-Language 头并返回受支持的语言标识。
// 解析仅遵循标签出现顺序，未实现质量值加权匹配，当前客户端均显式携带单一语言标签。
// 地区与书写变体折叠为主子标签，zh-CN 与 zh-Hans 均折叠为 zh；
// 整头为通配符时返回 LanguageAll，交由响应组装层输出全量翻译包；
// 请求头为空或白名单全部未命中时返回 fallback。
func ParseLanguage(raw string, supported []Language, fallback Language) Language {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback
	}
	if trimmed == string(LanguageAll) {
		return LanguageAll
	}

	for _, part := range strings.Split(trimmed, ",") {
		candidate := normalizeLanguageTag(part)
		if candidate == "" || candidate == LanguageAll {
			continue
		}
		for _, item := range supported {
			if candidate == item {
				return item
			}
		}
	}
	return fallback
}

// ParseLanguages 将配置中的语言字符串列表归一化为白名单成员。
// 归一化保证白名单元素与解析结果同形态可比，通配符与空值直接丢弃，重复标签只保留首个。
func ParseLanguages(values []string) []Language {
	normalized := make([]Language, 0, len(values))
	seen := make(map[Language]struct{}, len(values))
	for _, value := range values {
		tag := normalizeLanguageTag(value)
		if tag == "" || tag == LanguageAll {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	return normalized
}

// normalizeLanguageTag 剥离质量参数并折叠主子标签，输入 zh-CN;q=0.9 归一化为 zh。
func normalizeLanguageTag(tag string) Language {
	withoutQuality := tag
	if index := strings.Index(tag, ";"); index >= 0 {
		withoutQuality = tag[:index]
	}
	withoutQuality = strings.TrimSpace(withoutQuality)
	if index := strings.Index(withoutQuality, "-"); index >= 0 {
		withoutQuality = withoutQuality[:index]
	}
	return Language(strings.ToLower(withoutQuality))
}
