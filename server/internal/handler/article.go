package handler

import (
	"net/http"
	"strconv"

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
	ViewArticle(c *gin.Context)
	LikeArticle(c *gin.Context)
	UnlikeArticle(c *gin.Context)
	BookmarkArticle(c *gin.Context)
	UnbookmarkArticle(c *gin.Context)
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
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, article)
}

// GetArticle 获取文章详情
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 获取当前用户ID（可选）
	var userID *uint
	if uid, exists := c.Get("userID"); exists {
		uidUint := uid.(uint)
		userID = &uidUint
	}

	// 获取文章
	article, err := h.articleService.GetArticle(uint(id), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, article)
}

// GetArticleBySlug 根据Slug获取文章
func (h *ArticleHandler) GetArticleBySlug(c *gin.Context) {
	// 获取slug参数
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "文章slug不能为空")
		return
	}

	// 获取当前用户ID（可选）
	var userID *uint
	if uid, exists := c.Get("userID"); exists {
		uidUint := uid.(uint)
		userID = &uidUint
	}

	// 获取文章
	article, err := h.articleService.GetArticleBySlug(slug, userID)
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

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 绑定请求参数
	var req service.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 更新文章
	article, err := h.articleService.UpdateArticle(uint(id), &req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 删除文章
	err = h.articleService.DeleteArticle(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	// 获取当前用户ID（可选）
	var userID *uint
	if uid, exists := c.Get("userID"); exists {
		uidUint := uid.(uint)
		userID = &uidUint
	}

	// 获取文章列表
	result, err := h.articleService.GetArticleList(&req, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

// GetArticlesByAuthor 获取指定作者的文章
func (h *ArticleHandler) GetArticlesByAuthor(c *gin.Context) {
	// 获取作者ID
	authorIDParam := c.Param("authorId")
	if authorIDParam == "" {
		response.Error(c, http.StatusBadRequest, "作者ID不能为空")
		return
	}

	authorID, err := strconv.ParseUint(authorIDParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的作者ID")
		return
	}

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

	// 获取文章列表
	result, err := h.articleService.GetArticlesByAuthor(uint(authorID), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

// GetArticlesByCategory 获取指定分类的文章
func (h *ArticleHandler) GetArticlesByCategory(c *gin.Context) {
	// 获取分类ID
	categoryIDParam := c.Param("categoryId")
	if categoryIDParam == "" {
		response.Error(c, http.StatusBadRequest, "分类ID不能为空")
		return
	}

	categoryID, err := strconv.ParseUint(categoryIDParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的分类ID")
		return
	}

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

	// 获取文章列表
	result, err := h.articleService.GetArticlesByCategory(uint(categoryID), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

// GetArticlesByTag 获取指定标签的文章
func (h *ArticleHandler) GetArticlesByTag(c *gin.Context) {
	// 获取标签ID
	tagIDParam := c.Param("tagId")
	if tagIDParam == "" {
		response.Error(c, http.StatusBadRequest, "标签ID不能为空")
		return
	}

	tagID, err := strconv.ParseUint(tagIDParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的标签ID")
		return
	}

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

	// 获取文章列表
	result, err := h.articleService.GetArticlesByTag(uint(tagID), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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
		response.Error(c, http.StatusInternalServerError, err.Error())
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
		response.Error(c, http.StatusInternalServerError, err.Error())
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
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"articles": articles})
}

// GetRelatedArticles 获取相关文章
func (h *ArticleHandler) GetRelatedArticles(c *gin.Context) {
	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	type RelatedRequest struct {
		Limit int `json:"limit"`
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
	articles, err := h.articleService.GetRelatedArticles(uint(id), req.Limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"articles": articles})
}

// ViewArticle 记录文章浏览
func (h *ArticleHandler) ViewArticle(c *gin.Context) {
	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 获取当前用户ID（可选）
	var userID *uint
	if uid, exists := c.Get("userID"); exists {
		uidUint := uid.(uint)
		userID = &uidUint
	}

	// 获取访客ID和IP地址
	visitorID := c.GetHeader("Visitor-ID")
	ipAddress := c.ClientIP()

	// 记录浏览
	err = h.articleService.ViewArticle(uint(id), userID, visitorID, ipAddress)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 点赞文章
	err = h.articleService.LikeArticle(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 取消点赞文章
	err = h.articleService.UnlikeArticle(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "取消点赞成功"})
}

// BookmarkArticle 收藏文章
func (h *ArticleHandler) BookmarkArticle(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 收藏文章
	err = h.articleService.BookmarkArticle(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 取消收藏文章
	err = h.articleService.UnbookmarkArticle(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 发布文章
	err = h.articleService.PublishArticle(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 取消发布文章
	err = h.articleService.UnpublishArticle(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 归档文章
	err = h.articleService.ArchiveArticle(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	// 获取文章ID
	idParam := c.Param("id")
	if idParam == "" {
		response.Error(c, http.StatusBadRequest, "文章ID不能为空")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文章ID")
		return
	}

	// 设置文章为私有
	err = h.articleService.SetArticlePrivate(uint(id), userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "文章设置为私有成功"})
}
