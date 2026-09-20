// Package service 业务逻辑层
package service

import (
	"fmt"

	"MyBlog/internal/domain"
)

// 语言协商策略注入与服务端翻译写入校验。
// 白名单权威在配置文件，经组合根转换为 domain.LanguagePolicy 后注入各内容服务。

// languagePolicyHolder 承载服务层的语言协商策略，供翻译写入键校验复用。
type languagePolicyHolder struct {
	languagePolicy domain.LanguagePolicy
}

// requireWritableTranslation 校验翻译语言键可写，非法键返回映射 400 的请求错误。
// 缺省语言内容走主列，白名单外语言缺少翻译行承载，两者均拒绝写入。
func (h languagePolicyHolder) requireWritableTranslation(language domain.Language) error {
	if h.languagePolicy.IsWritableTranslation(language) {
		return nil
	}
	return fmt.Errorf("%w：不支持写入语言 %s 的翻译", ErrInvalidRequest, language)
}

// ArticleServiceOption 文章服务功能选项。
type ArticleServiceOption func(*ArticleService)

// WithArticleLanguagePolicy 注入文章服务的语言协商策略。
func WithArticleLanguagePolicy(policy domain.LanguagePolicy) ArticleServiceOption {
	return func(s *ArticleService) {
		s.languagePolicy = policy
	}
}

// CategoryServiceOption 分类服务功能选项。
type CategoryServiceOption func(*CategoryService)

// WithCategoryLanguagePolicy 注入分类服务的语言协商策略。
func WithCategoryLanguagePolicy(policy domain.LanguagePolicy) CategoryServiceOption {
	return func(s *CategoryService) {
		s.languagePolicy = policy
	}
}

// TagServiceOption 标签服务功能选项。
type TagServiceOption func(*TagService)

// WithTagLanguagePolicy 注入标签服务的语言协商策略。
func WithTagLanguagePolicy(policy domain.LanguagePolicy) TagServiceOption {
	return func(s *TagService) {
		s.languagePolicy = policy
	}
}

// DictServiceOption 字典服务功能选项。
type DictServiceOption func(*DictService)

// WithDictLanguagePolicy 注入字典服务的语言协商策略。
func WithDictLanguagePolicy(policy domain.LanguagePolicy) DictServiceOption {
	return func(s *DictService) {
		s.languagePolicy = policy
	}
}
