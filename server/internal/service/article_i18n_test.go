package service

import (
	"errors"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
)

// TestBuildArticleTranslationRowsRejectsDefaultLanguage 验证缺省语言翻译键被拒绝。
func TestBuildArticleTranslationRowsRejectsDefaultLanguage(t *testing.T) {
	svc := newCreateArticleService(&fakeCreateArticleRepo{})

	title := "中文标题"
	_, err := svc.buildArticleTranslationRows(nil, map[domain.Language]ArticleI18nPayload{
		domain.DefaultLanguage: {Title: &title},
	})

	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("缺省语言翻译键应返回 ErrInvalidRequest, 实际: %v", err)
	}
}

// TestBuildArticleTranslationRowsRejectsUnsupportedLanguage 验证白名单外语言键被拒绝。
func TestBuildArticleTranslationRowsRejectsUnsupportedLanguage(t *testing.T) {
	svc := newCreateArticleService(&fakeCreateArticleRepo{})

	title := "Bonjour"
	_, err := svc.buildArticleTranslationRows(nil, map[domain.Language]ArticleI18nPayload{
		domain.Language("fr"): {Title: &title},
	})

	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("白名单外语言键应返回 ErrInvalidRequest, 实际: %v", err)
	}
}

// TestBuildArticleTranslationRowsMergesPatch 验证补丁合并语义，提供即更新，未提供保留原值。
func TestBuildArticleTranslationRowsMergesPatch(t *testing.T) {
	svc := newCreateArticleService(&fakeCreateArticleRepo{})

	existingTitle := "Old Title"
	rows, err := svc.buildArticleTranslationRows(
		[]model.ArticleTranslation{{Locale: "en", Title: existingTitle, Summary: "old summary"}},
		map[domain.Language]ArticleI18nPayload{
			domain.LanguageEnglish: {Summary: strPtr("new summary")},
		},
	)
	if err != nil {
		t.Fatalf("合并翻译补丁不应返回错误: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("翻译行数量 = %d, 期望 1", len(rows))
	}
	if rows[0].Title != existingTitle {
		t.Errorf("Title = %q, 期望保留既有值 %q", rows[0].Title, existingTitle)
	}
	if rows[0].Summary != "new summary" {
		t.Errorf("Summary = %q, 期望更新为 new summary", rows[0].Summary)
	}
}

// TestBuildArticleTranslationRowsRendersContent 验证翻译正文随写入渲染缓存并统计字数。
func TestBuildArticleTranslationRowsRendersContent(t *testing.T) {
	svc := newCreateArticleService(&fakeCreateArticleRepo{})

	rows, err := svc.buildArticleTranslationRows(nil, map[domain.Language]ArticleI18nPayload{
		domain.LanguageEnglish: {Content: strPtr("hello world")},
	})
	if err != nil {
		t.Fatalf("构造翻译行不应返回错误: %v", err)
	}

	if rows[0].WordCount != 11 {
		t.Errorf("WordCount = %d, 期望 11", rows[0].WordCount)
	}
	if rows[0].ContentHTML == "" {
		t.Error("翻译正文应生成渲染缓存")
	}
}

// TestCreateArticlePassesTranslations 验证创建文章时翻译行透传至仓储事务方法。
func TestCreateArticlePassesTranslations(t *testing.T) {
	articleRepo := &fakeCreateArticleRepo{}
	svc := newCreateArticleService(articleRepo)

	englishTitle := "Hello"
	_, err := svc.CreateArticle(&CreateArticleRequest{
		Title:   "测试文章",
		Content: "正文内容",
		I18n: map[domain.Language]ArticleI18nPayload{
			domain.LanguageEnglish: {Title: &englishTitle},
		},
	}, 1)
	if err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	if len(articleRepo.lastTranslations) != 1 {
		t.Fatalf("翻译行未透传至仓储, 实际数量 %d", len(articleRepo.lastTranslations))
	}
	if articleRepo.lastTranslations[0].Locale != "en" || articleRepo.lastTranslations[0].Title != englishTitle {
		t.Errorf("翻译行内容不符: %+v", articleRepo.lastTranslations[0])
	}
}

// strPtr 返回字符串指针，供构造可选补丁字段使用。
func strPtr(value string) *string {
	return &value
}
