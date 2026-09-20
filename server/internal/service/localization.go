// Package service 业务逻辑层
package service

import (
	"sort"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/pkg/markdown"
)

// 内容本地化输出逻辑。
// 翻译行允许按字段部分填写，本地化以字段为单位回退：
// 目标语言提供了该字段则使用翻译值，否则保留主列的缺省语言内容。

// LocalizeArticle 按目标语言就地替换文章的本地化字段，返回实际输出语言。
// 文章本体未命中任何翻译字段时实际语言回退缺省语言，内嵌分类与标签按字段独立回退；
// 输出前装配轻量翻译状态标记，具体语言请求额外清空翻译行集合，公共响应不携带全量翻译包。
func LocalizeArticle(article *model.Article, language domain.Language) domain.Language {
	if article == nil {
		return language
	}

	// 全量包模式保留翻译行原样输出，仅装配轻量翻译状态标记。
	if language == domain.LanguageAll {
		article.TranslationLocales = collectArticleTranslationLocales(article.Translations)
		return language
	}

	localized := applyArticleTranslation(article, findArticleTranslation(article.Translations, language))
	article.TranslationLocales = collectArticleTranslationLocales(article.Translations)
	article.Translations = nil

	// 内嵌分类与标签的翻译缺失不影响文章本体的语言标注。
	for index := range article.Categories {
		localizeCategoryInPlace(&article.Categories[index], language)
	}
	if article.Category != nil {
		localizeCategoryInPlace(article.Category, language)
	}
	for index := range article.Tags {
		localizeTagInPlace(&article.Tags[index], language)
	}

	if localized {
		return language
	}
	return domain.DefaultLanguage
}

// LocalizeCategory 按目标语言就地替换分类的本地化字段，返回实际输出语言。
// 未命中任何翻译字段时实际语言回退缺省语言。
func LocalizeCategory(category *model.Category, language domain.Language) domain.Language {
	if language == domain.LanguageAll || category == nil {
		return language
	}
	if localizeCategoryInPlace(category, language) {
		return language
	}
	return domain.DefaultLanguage
}

// LocalizeTag 按目标语言就地替换标签的本地化字段，返回实际输出语言。
// 未命中任何翻译字段时实际语言回退缺省语言。
func LocalizeTag(tag *model.Tag, language domain.Language) domain.Language {
	if language == domain.LanguageAll || tag == nil {
		return language
	}
	if localizeTagInPlace(tag, language) {
		return language
	}
	return domain.DefaultLanguage
}

// LocalizeCategoryTree 按目标语言本地化分类树的全部节点，缺失翻译的节点保留缺省语言内容。
func LocalizeCategoryTree(nodes []*CategoryTreeNode, language domain.Language) {
	if language == domain.LanguageAll {
		return
	}
	for _, node := range nodes {
		if node == nil || node.Category == nil {
			continue
		}
		localizeCategoryInPlace(node.Category, language)
		LocalizeCategoryTree(node.Children, language)
	}
}

// applyArticleTranslation 将翻译行的非空字段替换到文章本体，返回是否存在有效翻译字段。
// 正文替换时同步替换渲染缓存与字数统计，渲染缓存缺失时按主列一致管线即时补渲染。
func applyArticleTranslation(article *model.Article, translation *model.ArticleTranslation) bool {
	if translation == nil {
		return false
	}

	localized := false
	if translation.Title != "" {
		article.Title = translation.Title
		localized = true
	}
	if translation.Summary != "" {
		article.Summary = translation.Summary
		localized = true
	}
	if translation.Content != "" {
		article.Content = translation.Content
		article.ContentHTML = translation.ContentHTML
		ensureArticleContentHTML(article)
		article.WordCount = translation.WordCount
		article.ReadingTime = article.CalculateReadingTime()
		localized = true
	}
	if translation.SEOTitle != "" {
		article.SEOTitle = translation.SEOTitle
		localized = true
	}
	if translation.SEODescription != "" {
		article.SEODescription = translation.SEODescription
		localized = true
	}
	if translation.SEOKeywords != "" {
		article.SEOKeywords = translation.SEOKeywords
		localized = true
	}
	return localized
}

// localizeCategoryInPlace 就地替换分类的本地化字段并清空翻译行，返回是否存在有效翻译字段。
func localizeCategoryInPlace(category *model.Category, language domain.Language) bool {
	translation := findCategoryTranslation(category.Translations, language)
	if translation == nil {
		return false
	}

	localized := false
	if translation.Name != "" {
		category.Name = translation.Name
		localized = true
	}
	if translation.Description != "" {
		category.Description = translation.Description
		localized = true
	}
	if translation.SEOTitle != "" {
		category.SEOTitle = translation.SEOTitle
		localized = true
	}
	if translation.SEODescription != "" {
		category.SEODescription = translation.SEODescription
		localized = true
	}
	category.Translations = nil
	return localized
}

// localizeTagInPlace 就地替换标签的本地化字段并清空翻译行，返回是否存在有效翻译字段。
func localizeTagInPlace(tag *model.Tag, language domain.Language) bool {
	translation := findTagTranslation(tag.Translations, language)
	if translation == nil {
		return false
	}

	localized := false
	if translation.Name != "" {
		tag.Name = translation.Name
		localized = true
	}
	if translation.Description != "" {
		tag.Description = translation.Description
		localized = true
	}
	tag.Translations = nil
	return localized
}

// ensureArticleContentHTML 为缺失渲染缓存的翻译正文即时补渲染。
// 渲染失败不阻断读取，空缓存交由展示层按纯文本降级处理。
func ensureArticleContentHTML(article *model.Article) {
	if article.ContentHTML != "" {
		return
	}
	if rendered, err := markdown.Render(article.Content); err == nil {
		article.ContentHTML = rendered
	}
}

// collectArticleTranslationLocales 提取存在有效翻译内容的语言列表，按字典序输出保证响应稳定。
func collectArticleTranslationLocales(rows []model.ArticleTranslation) []string {
	locales := make([]string, 0, len(rows))
	for _, row := range rows {
		if articleTranslationHasContent(row) {
			locales = append(locales, row.Locale)
		}
	}
	sort.Strings(locales)
	if len(locales) == 0 {
		return nil
	}
	return locales
}

// articleTranslationHasContent 判断翻译行是否携带有效内容，空行在输出时不计入翻译状态。
func articleTranslationHasContent(row model.ArticleTranslation) bool {
	return row.Title != "" || row.Summary != "" || row.Content != "" ||
		row.SEOTitle != "" || row.SEODescription != "" || row.SEOKeywords != ""
}

// findArticleTranslation 按语言定位翻译行，未命中时返回 nil。
func findArticleTranslation(rows []model.ArticleTranslation, language domain.Language) *model.ArticleTranslation {
	for index := range rows {
		if rows[index].Locale == string(language) {
			return &rows[index]
		}
	}
	return nil
}

// findCategoryTranslation 按语言定位分类翻译行，未命中时返回 nil。
func findCategoryTranslation(rows []model.CategoryTranslation, language domain.Language) *model.CategoryTranslation {
	for index := range rows {
		if rows[index].Locale == string(language) {
			return &rows[index]
		}
	}
	return nil
}

// findTagTranslation 按语言定位标签翻译行，未命中时返回 nil。
func findTagTranslation(rows []model.TagTranslation, language domain.Language) *model.TagTranslation {
	for index := range rows {
		if rows[index].Locale == string(language) {
			return &rows[index]
		}
	}
	return nil
}
