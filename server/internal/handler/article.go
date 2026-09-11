package handler

import (
	"net/http"

	"MyBlog/internal/service"
	"MyBlog/pkg/response"

	"github.com/gin-gonic/gin"
)

// ArticleHandlerInterface 文章处理器接口
type ArticleHandlerInterface interface {
	CreateArticle(c *gin.Context)
	GetArticle(c *gin.Context)
	GetArticleBySlug(c *gin.Context)
	UpdateArticle(c *gin.Context)
	DeleteArticle(c *gin.Context)
	GetArticleList(c *gin.Context)
	GetArticlesByAuthor(c *gin.Context)
	GetArticlesByCategory(c *gin.Context)
	GetArticlesByTag(c *gin.Context)
	SearchArticles(c *gin.Context)
	GetPopularArticles(c *gin.Context)
	GetRecentArticles(c *gin.Context)
	GetRelatedArticles(c *gin.Context)
	GetArticleArchives(c *gin.Context)
	ViewArticle(c *gin.Context)
	LikeArticle(c *gin.Context)
	UnlikeArticle(c *gin.Context)
	BookmarkArticle(c *gin.Context)
	UnbookmarkArticle(c *gin.Context)
	IsArticleLiked(c *gin.Context)
	IsArticleBookmarked(c *gin.Context)
	GetArticleBookmarks(c *gin.Context)
	PublishArticle(c *gin.Context)
	UnpublishArticle(c *gin.Context)
	ArchiveArticle(c *gin.Context)
	SetArticlePrivate(c *gin.Context)
}

// ArticleHandler 文章处理器实现
type ArticleHandler struct {
	articleService service.ArticleServiceInterface
}

// NewArticleHandler 创建文章处理器实例
func NewArticleHandler(articleService service.ArticleServiceInterface) ArticleHandlerInterface {
	return &ArticleHandler{
		articleService: articleService,
	}
}

// CreateArticle 创建文章
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	var req service.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 创建文章
	article, err := h.articleService.CreateArticle(&req, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, article)
}

// GetArticle 获取文章详情
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	// 绑定请求参数
	type GetArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req GetArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 读取当前用户ID，未登录时为空。
	userID := getOptionalUserID(c)

	// 获取文章
	article, err := h.articleService.GetArticle(req.ID, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, article)
}

// GetArticleBySlug 根据Slug获取文章
func (h *ArticleHandler) GetArticleBySlug(c *gin.Context) {
	// 绑定请求参数
	type GetArticleBySlugRequest struct {
		Slug string `json:"slug" binding:"required"`
	}

	var req GetArticleBySlugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 读取当前用户ID，未登录时为空。
	userID := getOptionalUserID(c)

	// 获取文章
	article, err := h.articleService.GetArticleBySlug(req.Slug, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, article)
}

// UpdateArticle 更新文章
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type UpdateArticleRequestWithID struct {
		ID uint `json:"id" binding:"required"`
		service.UpdateArticleRequest
	}

	var req UpdateArticleRequestWithID
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 更新文章
	article, err := h.articleService.UpdateArticle(req.ID, &req.UpdateArticleRequest, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, article)
}

// DeleteArticle 删除文章
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type DeleteArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req DeleteArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 删除文章
	err := h.articleService.DeleteArticle(req.ID, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "文章删除成功"})
}

// GetArticleList 获取文章列表
func (h *ArticleHandler) GetArticleList(c *gin.Context) {
	// 绑定请求参数
	var req service.GetArticleListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 读取当前用户ID，未登录时为空。
	userID := getOptionalUserID(c)

	// 获取文章列表
	result, err := h.articleService.GetArticleList(&req, userID)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// GetArticlesByAuthor 获取指定作者的文章
func (h *ArticleHandler) GetArticlesByAuthor(c *gin.Context) {
	// 绑定请求参数
	type GetArticlesByAuthorRequest struct {
		AuthorID uint `json:"authorId" binding:"required"`
		service.GetArticleListRequest
	}

	var req GetArticlesByAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 读取当前用户ID，未登录时为空，供服务层做角色化可见性判断。
	userID := getOptionalUserID(c)

	// 获取文章列表
	result, err := h.articleService.GetArticlesByAuthor(req.AuthorID, &req.GetArticleListRequest, userID)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// GetArticlesByCategory 获取指定分类的文章
func (h *ArticleHandler) GetArticlesByCategory(c *gin.Context) {
	// 绑定请求参数
	type GetArticlesByCategoryRequest struct {
		CategoryID uint `json:"categoryId" binding:"required"`
		service.GetArticleListRequest
	}

	var req GetArticlesByCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 获取文章列表
	result, err := h.articleService.GetArticlesByCategory(req.CategoryID, &req.GetArticleListRequest)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// GetArticlesByTag 获取指定标签的文章
func (h *ArticleHandler) GetArticlesByTag(c *gin.Context) {
	// 绑定请求参数
	type GetArticlesByTagRequest struct {
		TagID uint `json:"tagId" binding:"required"`
		service.GetArticleListRequest
	}

	var req GetArticlesByTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 获取文章列表
	result, err := h.articleService.GetArticlesByTag(req.TagID, &req.GetArticleListRequest)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// SearchArticles 搜索文章
func (h *ArticleHandler) SearchArticles(c *gin.Context) {
	// 获取搜索关键词
	type SearchRequest struct {
		Keyword string `json:"keyword" binding:"required"`
		service.GetArticleListRequest
	}

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 搜索文章
	result, err := h.articleService.SearchArticles(req.Keyword, &req.GetArticleListRequest)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// GetPopularArticles 获取热门文章
func (h *ArticleHandler) GetPopularArticles(c *gin.Context) {
	type PopularRequest struct {
		Limit int `json:"limit"`
	}

	var req PopularRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// 获取热门文章
	articles, err := h.articleService.GetPopularArticles(req.Limit)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"articles": articles})
}

// GetRecentArticles 获取最新文章
func (h *ArticleHandler) GetRecentArticles(c *gin.Context) {
	type RecentRequest struct {
		Limit int `json:"limit"`
	}

	var req RecentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// 获取最新文章
	articles, err := h.articleService.GetRecentArticles(req.Limit)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"articles": articles})
}

// GetArticleArchives 获取按年月分组的公开文章归档，无请求参数。
func (h *ArticleHandler) GetArticleArchives(c *gin.Context) {
	groups, err := h.articleService.GetArticleArchives()
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, groups)
}

// GetRelatedArticles 获取相关文章
func (h *ArticleHandler) GetRelatedArticles(c *gin.Context) {
	// 绑定请求参数
	type RelatedRequest struct {
		ID    uint `json:"id" binding:"required"`
		Limit int  `json:"limit"`
	}

	var req RelatedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 5
	}

	// 获取相关文章
	articles, err := h.articleService.GetRelatedArticles(req.ID, req.Limit)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"articles": articles})
}

// ViewArticle 记录文章浏览
func (h *ArticleHandler) ViewArticle(c *gin.Context) {
	// 绑定请求参数
	type ViewArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req ViewArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 读取当前用户ID，未登录时为空。
	userID := getOptionalUserID(c)

	// 获取访客ID和IP地址
	visitorID := c.GetHeader("Visitor-ID")
	ipAddress := c.ClientIP()

	// 记录浏览
	err := h.articleService.ViewArticle(req.ID, userID, visitorID, ipAddress)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "浏览记录成功"})
}

// LikeArticle 点赞文章
func (h *ArticleHandler) LikeArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type LikeArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req LikeArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 点赞文章
	err := h.articleService.LikeArticle(req.ID, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "点赞成功"})
}

// UnlikeArticle 取消点赞文章
func (h *ArticleHandler) UnlikeArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type UnlikeArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req UnlikeArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 取消点赞文章
	err := h.articleService.UnlikeArticle(req.ID, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "取消点赞成功"})
}

// IsArticleLiked 查询当前用户对文章的点赞状态 POST /api/articles/isLiked
func (h *ArticleHandler) IsArticleLiked(c *gin.Context) {
	type IsLikedRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req IsLikedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	isLiked, err := h.articleService.IsArticleLiked(req.ID, userID)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"isLiked": isLiked})
}

// IsArticleBookmarked 查询当前用户对文章的收藏状态 POST /api/articles/isBookmarked
func (h *ArticleHandler) IsArticleBookmarked(c *gin.Context) {
	type IsBookmarkedRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req IsBookmarkedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	isBookmarked, err := h.articleService.IsArticleBookmarked(req.ID, userID)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"isBookmarked": isBookmarked})
}

// GetArticleBookmarks 分页查询当前用户的收藏文章 POST /api/articles/bookmarks
func (h *ArticleHandler) GetArticleBookmarks(c *gin.Context) {
	var req service.GetArticleListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	userID, ok := getOperatorID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}

	result, err := h.articleService.GetArticleBookmarks(userID, &req)
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, result)
}

// BookmarkArticle 收藏文章
func (h *ArticleHandler) BookmarkArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type BookmarkArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req BookmarkArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 收藏文章
	err := h.articleService.BookmarkArticle(req.ID, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "收藏成功"})
}

// UnbookmarkArticle 取消收藏文章
func (h *ArticleHandler) UnbookmarkArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type UnbookmarkArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req UnbookmarkArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 取消收藏文章
	err := h.articleService.UnbookmarkArticle(req.ID, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "取消收藏成功"})
}

// PublishArticle 发布文章
func (h *ArticleHandler) PublishArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type PublishArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req PublishArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 发布文章
	err := h.articleService.PublishArticle(req.ID, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "文章发布成功"})
}

// UnpublishArticle 取消发布文章
func (h *ArticleHandler) UnpublishArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type UnpublishArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req UnpublishArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 取消发布文章
	err := h.articleService.UnpublishArticle(req.ID, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "取消发布成功"})
}

// ArchiveArticle 归档文章
func (h *ArticleHandler) ArchiveArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type ArchiveArticleRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req ArchiveArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 归档文章
	err := h.articleService.ArchiveArticle(req.ID, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "文章归档成功"})
}

// SetArticlePrivate 设置文章为私有
func (h *ArticleHandler) SetArticlePrivate(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 绑定请求参数
	type SetArticlePrivateRequest struct {
		ID uint `json:"id" binding:"required"`
	}

	var req SetArticlePrivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置文章为私有
	err := h.articleService.SetArticlePrivate(req.ID, userID.(uint))
	if err != nil {
		HandleServiceError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "文章设置为私有成功"})
}
