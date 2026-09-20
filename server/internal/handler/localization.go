// Package handler HTTP请求处理层
package handler

import (
	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// 内容本地化响应输出助手。
// 全量包请求保留翻译行原样输出，其余语言本地化后输出单语言内容；
// 单资源响应的 Content-Language 标注实际输出语言，集合响应标注请求语言。

// respondLocalizedArticle 按请求语言本地化后输出文章，响应头标注实际输出语言。
func respondLocalizedArticle(c *gin.Context, article *model.Article) {
	language := getRequestLanguage(c)
	actual := service.LocalizeArticle(article, language)
	c.Header(ContentLanguageHeader, string(actual))
	response.Success(c, article)
}

// respondLocalizedArticleList 按请求语言本地化后输出文章列表。
func respondLocalizedArticleList(c *gin.Context, list *service.ArticleListResponse) {
	language := localizeArticles(list.Articles, getRequestLanguage(c))
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, list)
}

// respondLocalizedArticleCollection 本地化后按 articles 包裹形状输出文章集合。
func respondLocalizedArticleCollection(c *gin.Context, articles []*model.Article) {
	language := localizeArticles(articles, getRequestLanguage(c))
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, gin.H{"articles": articles})
}

// respondLocalizedArchives 按请求语言本地化后输出归档分组内的文章。
func respondLocalizedArchives(c *gin.Context, years []service.ArticleArchiveYear) {
	language := getRequestLanguage(c)
	if language != domain.LanguageAll {
		for _, year := range years {
			for _, month := range year.Months {
				localizeArticles(month.Articles, language)
			}
		}
	}
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, years)
}

// localizeArticles 就地本地化文章集合，全量包请求原样保留。
func localizeArticles(articles []*model.Article, language domain.Language) domain.Language {
	if language == domain.LanguageAll {
		return language
	}
	for _, article := range articles {
		service.LocalizeArticle(article, language)
	}
	return language
}

// respondLocalizedCategory 按请求语言本地化后输出分类，响应头标注实际输出语言。
func respondLocalizedCategory(c *gin.Context, category *model.Category) {
	language := getRequestLanguage(c)
	actual := service.LocalizeCategory(category, language)
	c.Header(ContentLanguageHeader, string(actual))
	response.Success(c, category)
}

// respondLocalizedCategoryMessage 本地化后输出带提示消息的分类响应。
func respondLocalizedCategoryMessage(c *gin.Context, message string, category *model.Category) {
	language := getRequestLanguage(c)
	actual := service.LocalizeCategory(category, language)
	c.Header(ContentLanguageHeader, string(actual))
	response.SuccessWithMessage(c, message, category)
}

// respondLocalizedCategoryList 按请求语言本地化后输出分类列表。
func respondLocalizedCategoryList(c *gin.Context, list *service.CategoryListResponse) {
	language := getRequestLanguage(c)
	if language != domain.LanguageAll {
		for _, category := range list.Categories {
			service.LocalizeCategory(category, language)
		}
	}
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, list)
}

// respondLocalizedCategoryTree 按请求语言本地化后输出分类树。
func respondLocalizedCategoryTree(c *gin.Context, tree gin.H) {
	language := getRequestLanguage(c)
	if nodes, ok := tree["tree"].([]*service.CategoryTreeNode); ok {
		service.LocalizeCategoryTree(nodes, language)
	}
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, tree)
}

// respondLocalizedTag 按请求语言本地化后输出标签，响应头标注实际输出语言。
func respondLocalizedTag(c *gin.Context, tag *model.Tag) {
	language := getRequestLanguage(c)
	actual := service.LocalizeTag(tag, language)
	c.Header(ContentLanguageHeader, string(actual))
	response.Success(c, tag)
}

// respondLocalizedTagMessage 本地化后输出带提示消息的标签响应。
func respondLocalizedTagMessage(c *gin.Context, message string, tag *model.Tag) {
	language := getRequestLanguage(c)
	actual := service.LocalizeTag(tag, language)
	c.Header(ContentLanguageHeader, string(actual))
	response.SuccessWithMessage(c, message, tag)
}

// respondLocalizedTagList 按请求语言本地化后输出标签列表。
func respondLocalizedTagList(c *gin.Context, list *service.TagListResponse) {
	language := localizeTags(list.Tags, getRequestLanguage(c))
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, list)
}

// respondLocalizedTagCollection 本地化后按 tags 包裹形状输出标签集合。
func respondLocalizedTagCollection(c *gin.Context, tags []*model.Tag) {
	language := localizeTags(tags, getRequestLanguage(c))
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, gin.H{"tags": tags})
}

// localizeTags 就地本地化标签集合，全量包请求原样保留。
func localizeTags(tags []*model.Tag, language domain.Language) domain.Language {
	if language == domain.LanguageAll {
		return language
	}
	for _, tag := range tags {
		service.LocalizeTag(tag, language)
	}
	return language
}

// respondLocalizedDictGroups 本地化后按 dicts 包裹形状输出全量字典分组。
func respondLocalizedDictGroups(c *gin.Context, groups []*service.EnabledDictGroup) {
	language := getRequestLanguage(c)
	if language != domain.LanguageAll {
		for _, group := range groups {
			service.LocalizeDictGroup(group, language)
		}
	}
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, gin.H{"dicts": groups})
}

// respondLocalizedDictGroup 本地化后输出单个字典分组。
// 分组属于集合形状，组内条目按字段独立回退，Content-Language 标注请求语言。
func respondLocalizedDictGroup(c *gin.Context, group *service.EnabledDictGroup) {
	language := getRequestLanguage(c)
	if language != domain.LanguageAll {
		service.LocalizeDictGroup(group, language)
	}
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, group)
}

// respondLocalizedDictType 按请求语言本地化后输出字典类型，响应头标注实际输出语言。
func respondLocalizedDictType(c *gin.Context, dictType *model.DictType) {
	language := getRequestLanguage(c)
	actual := service.LocalizeDictType(dictType, language)
	c.Header(ContentLanguageHeader, string(actual))
	response.Success(c, dictType)
}

// respondLocalizedDictTypeMessage 本地化后输出带提示消息的字典类型响应。
func respondLocalizedDictTypeMessage(c *gin.Context, message string, dictType *model.DictType) {
	language := getRequestLanguage(c)
	actual := service.LocalizeDictType(dictType, language)
	c.Header(ContentLanguageHeader, string(actual))
	response.SuccessWithMessage(c, message, dictType)
}

// respondLocalizedDictTypeList 按请求语言本地化后输出字典类型分页列表。
func respondLocalizedDictTypeList(c *gin.Context, list *service.DictTypeListResponse) {
	language := localizeDictTypes(list.Types, getRequestLanguage(c))
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, list)
}

// respondLocalizedDictItem 按请求语言本地化后输出字典项，响应头标注实际输出语言。
func respondLocalizedDictItem(c *gin.Context, item *model.DictItem) {
	language := getRequestLanguage(c)
	actual := service.LocalizeDictItem(item, language)
	c.Header(ContentLanguageHeader, string(actual))
	response.Success(c, item)
}

// respondLocalizedDictItemMessage 本地化后输出带提示消息的字典项响应。
func respondLocalizedDictItemMessage(c *gin.Context, message string, item *model.DictItem) {
	language := getRequestLanguage(c)
	actual := service.LocalizeDictItem(item, language)
	c.Header(ContentLanguageHeader, string(actual))
	response.SuccessWithMessage(c, message, item)
}

// respondLocalizedDictItemList 按请求语言本地化后输出字典项分页列表。
func respondLocalizedDictItemList(c *gin.Context, list *service.DictItemListResponse) {
	language := localizeDictItems(list.Items, getRequestLanguage(c))
	c.Header(ContentLanguageHeader, string(language))
	response.Success(c, list)
}

// localizeDictTypes 就地本地化字典类型集合，全量包请求原样保留。
func localizeDictTypes(types []*model.DictType, language domain.Language) domain.Language {
	if language == domain.LanguageAll {
		return language
	}
	for _, dictType := range types {
		service.LocalizeDictType(dictType, language)
	}
	return language
}

// localizeDictItems 就地本地化字典项集合，全量包请求原样保留。
func localizeDictItems(items []*model.DictItem, language domain.Language) domain.Language {
	if language == domain.LanguageAll {
		return language
	}
	for _, item := range items {
		service.LocalizeDictItem(item, language)
	}
	return language
}
