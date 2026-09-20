package service

import (
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
)

// newLocalizedArticleFixture 构造带中文主列与英文翻译行的文章测试数据。
func newLocalizedArticleFixture() *model.Article {
	article := &model.Article{
		ID:        1,
		Title:     "你好世界",
		Summary:   "中文摘要",
		Content:   "# 中文正文",
		WordCount: 5,
		Category: &model.Category{
			ID:   2,
			Name: "技术",
		},
		Tags: []model.Tag{
			{ID: 3, Name: "前端"},
		},
		Translations: []model.ArticleTranslation{
			{Locale: "en", Title: "Hello World", Summary: "English summary", Content: "English content", WordCount: 15},
		},
	}
	article.Category.Translations = []model.CategoryTranslation{
		{Locale: "en", Name: "Technology"},
	}
	article.Tags[0].Translations = []model.TagTranslation{
		{Locale: "en", Name: "frontend"},
	}
	return article
}

// TestLocalizeArticleSwapsTranslatedFields 验证请求英文时主列字段被翻译值替换。
func TestLocalizeArticleSwapsTranslatedFields(t *testing.T) {
	article := newLocalizedArticleFixture()

	actual := LocalizeArticle(article, domain.LanguageEnglish)

	if actual != domain.LanguageEnglish {
		t.Errorf("实际语言 = %q, 期望 en", actual)
	}
	if article.Title != "Hello World" || article.Summary != "English summary" || article.Content != "English content" {
		t.Errorf("翻译字段未替换: title=%q summary=%q content=%q", article.Title, article.Summary, article.Content)
	}
	if article.WordCount != 15 {
		t.Errorf("WordCount = %d, 期望使用翻译行的 15", article.WordCount)
	}
	if article.ReadingTime != 1 {
		t.Errorf("ReadingTime = %d, 期望按翻译字数估算为 1", article.ReadingTime)
	}
	if article.Category.Name != "Technology" {
		t.Errorf("内嵌分类名称 = %q, 期望翻译值 Technology", article.Category.Name)
	}
	if article.Tags[0].Name != "frontend" {
		t.Errorf("内嵌标签名称 = %q, 期望翻译值 frontend", article.Tags[0].Name)
	}
	if article.Translations != nil {
		t.Error("公共响应不应携带翻译行集合")
	}
	if article.Category.Translations != nil {
		t.Error("内嵌分类不应携带翻译行集合")
	}
	if len(article.TranslationLocales) != 1 || article.TranslationLocales[0] != "en" {
		t.Errorf("TranslationLocales = %v, 期望 [en]", article.TranslationLocales)
	}
}

// TestLocalizeArticleFallsBackPerField 验证缺失翻译的字段回退缺省语言内容。
func TestLocalizeArticleFallsBackPerField(t *testing.T) {
	article := newLocalizedArticleFixture()
	article.Translations = []model.ArticleTranslation{
		{Locale: "en", Title: "Hello World"},
	}

	actual := LocalizeArticle(article, domain.LanguageEnglish)

	if actual != domain.LanguageEnglish {
		t.Errorf("存在有效翻译时实际语言 = %q, 期望 en", actual)
	}
	if article.Title != "Hello World" {
		t.Errorf("Title = %q, 期望翻译值", article.Title)
	}
	if article.Summary != "中文摘要" || article.Content != "# 中文正文" {
		t.Error("缺失翻译的字段应保留缺省语言内容")
	}
}

// TestLocalizeArticleMissingTranslationFallsBackToDefault 验证完全缺失翻译时实际语言回退缺省语言。
func TestLocalizeArticleMissingTranslationFallsBackToDefault(t *testing.T) {
	article := newLocalizedArticleFixture()
	article.Translations = nil

	actual := LocalizeArticle(article, domain.LanguageEnglish)

	if actual != domain.DefaultLanguage {
		t.Errorf("实际语言 = %q, 期望回退 zh", actual)
	}
	if article.Title != "你好世界" {
		t.Error("完全缺失翻译时应保留缺省语言内容")
	}
	if article.TranslationLocales != nil {
		t.Errorf("无翻译行时 TranslationLocales = %v, 期望省略", article.TranslationLocales)
	}
}

// TestLocalizeArticleKeepsBundleForAllMode 验证全量包请求保留翻译行不做替换。
func TestLocalizeArticleKeepsBundleForAllMode(t *testing.T) {
	article := newLocalizedArticleFixture()

	actual := LocalizeArticle(article, domain.LanguageAll)

	if actual != domain.LanguageAll {
		t.Errorf("实际语言 = %q, 期望 *", actual)
	}
	if article.Title != "你好世界" {
		t.Error("全量包请求不应替换主列字段")
	}
	if len(article.Translations) != 1 {
		t.Error("全量包请求应保留翻译行集合")
	}
}

// TestLocalizeArticleFiltersEmptyTranslationRows 验证空翻译行不计入翻译状态标记。
func TestLocalizeArticleFiltersEmptyTranslationRows(t *testing.T) {
	article := newLocalizedArticleFixture()
	article.Translations = []model.ArticleTranslation{
		{Locale: "ja"},
		{Locale: "en", Title: "Hello"},
	}

	LocalizeArticle(article, domain.LanguageEnglish)

	if len(article.TranslationLocales) != 1 || article.TranslationLocales[0] != "en" {
		t.Errorf("TranslationLocales = %v, 期望仅含 en", article.TranslationLocales)
	}
}

// TestLocalizeCategoryAndTag 验证分类与标签的本地化与回退语义。
func TestLocalizeCategoryAndTag(t *testing.T) {
	category := &model.Category{
		ID:   1,
		Name: "技术",
		Translations: []model.CategoryTranslation{
			{Locale: "en", Name: "Technology", Description: "Tech articles"},
		},
	}
	tag := &model.Tag{
		ID:   2,
		Name: "前端",
		Translations: []model.TagTranslation{
			{Locale: "en", Name: "frontend"},
		},
	}

	if actual := LocalizeCategory(category, domain.LanguageEnglish); actual != domain.LanguageEnglish {
		t.Errorf("分类实际语言 = %q, 期望 en", actual)
	}
	if category.Name != "Technology" || category.Description != "Tech articles" {
		t.Error("分类翻译字段未替换")
	}

	if actual := LocalizeTag(tag, domain.LanguageEnglish); actual != domain.LanguageEnglish {
		t.Errorf("标签实际语言 = %q, 期望 en", actual)
	}
	if tag.Name != "frontend" {
		t.Error("标签翻译字段未替换")
	}

	// 缺失翻译时回退缺省语言。
	plain := &model.Category{ID: 3, Name: "生活"}
	if actual := LocalizeCategory(plain, domain.LanguageEnglish); actual != domain.DefaultLanguage {
		t.Errorf("缺失翻译的分类实际语言 = %q, 期望回退 zh", actual)
	}
	if plain.Name != "生活" {
		t.Error("缺失翻译的分类应保留缺省语言名称")
	}
}
