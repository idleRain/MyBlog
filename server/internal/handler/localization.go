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
