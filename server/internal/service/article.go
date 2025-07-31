package service

import (
	"errors"
	"html"
	"regexp"
	"strings"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// ArticleServiceInterface 文章服务接口
type ArticleServiceInterface interface {
	// 基础CRUD操作
	CreateArticle(req *CreateArticleRequest, authorID uint) (*model.Article, error)
	GetArticle(id uint, userID *uint) (*model.Article, error)
	GetArticleBySlug(slug string, userID *uint) (*model.Article, error)
	UpdateArticle(id uint, req *UpdateArticleRequest, userID uint) (*model.Article, error)
	DeleteArticle(id uint, userID uint) error

	// 查询操作
	GetArticleList(req *GetArticleListRequest, userID *uint) (*ArticleListResponse, error)
	GetArticlesByAuthor(authorID uint, req *GetArticleListRequest) (*ArticleListResponse, error)
	GetArticlesByCategory(categoryID uint, req *GetArticleListRequest) (*ArticleListResponse, error)
	GetArticlesByTag(tagID uint, req *GetArticleListRequest) (*ArticleListResponse, error)
	SearchArticles(keyword string, req *GetArticleListRequest) (*ArticleListResponse, error)

	// 统计和推荐
	GetPopularArticles(limit int) ([]*model.Article, error)
	GetRecentArticles(limit int) ([]*model.Article, error)
	GetRelatedArticles(articleID uint, limit int) ([]*model.Article, error)

	// 互动操作
	ViewArticle(articleID uint, userID *uint, visitorID string, ipAddress string) error
	LikeArticle(articleID uint, userID uint) error
	UnlikeArticle(articleID uint, userID uint) error
	BookmarkArticle(articleID uint, userID uint) error
	UnbookmarkArticle(articleID uint, userID uint) error

	// 状态管理
	PublishArticle(id uint, userID uint) error
	UnpublishArticle(id uint, userID uint) error
	ArchiveArticle(id uint, userID uint) error
	SetArticlePrivate(id uint, userID uint) error

	// 权限检查
	CanView(article *model.Article, userID *uint) bool
	CanEdit(article *model.Article, userID uint) bool
	CanDelete(article *model.Article, userID uint) bool
}

// 请求和响应结构体
type CreateArticleRequest struct {
	Title          string `json:"title" binding:"required,min=1,max=200"`
	Slug           string `json:"slug" binding:"max=200"`
	Summary        string `json:"summary" binding:"max=500"`
	Content        string `json:"content" binding:"required"`
	CoverImage     string `json:"coverImage" binding:"max=500"`
	CategoryID     *uint  `json:"categoryId"`
	CategoryIDs    []uint `json:"categoryIds"`
	TagIDs         []uint `json:"tagIds"`
	Status         string `json:"status" binding:"oneof=draft published private"`
	IsFeatured     bool   `json:"isFeatured"`
	IsTop          bool   `json:"isTop"`
	CommentEnabled bool   `json:"commentEnabled"`
	SEOTitle       string `json:"seoTitle" binding:"max=100"`
	SEODescription string `json:"seoDescription" binding:"max=255"`
	SEOKeywords    string `json:"seoKeywords" binding:"max=200"`
}

type UpdateArticleRequest struct {
	Title          string `json:"title" binding:"required,min=1,max=200"`
	Slug           string `json:"slug" binding:"max=200"`
	Summary        string `json:"summary" binding:"max=500"`
	Content        string `json:"content" binding:"required"`
	CoverImage     string `json:"coverImage" binding:"max=500"`
	CategoryID     *uint  `json:"categoryId"`
	CategoryIDs    []uint `json:"categoryIds"`
	TagIDs         []uint `json:"tagIds"`
	Status         string `json:"status" binding:"oneof=draft published archived private"`
	IsFeatured     bool   `json:"isFeatured"`
	IsTop          bool   `json:"isTop"`
	CommentEnabled bool   `json:"commentEnabled"`
	SEOTitle       string `json:"seoTitle" binding:"max=100"`
	SEODescription string `json:"seoDescription" binding:"max=255"`
	SEOKeywords    string `json:"seoKeywords" binding:"max=200"`
}

type GetArticleListRequest struct {
	Page     int    `json:"page" binding:"min=1"`
	PageSize int    `json:"pageSize" binding:"min=1,max=100"`
	Status   string `json:"status" binding:"oneof='' draft published archived private"`
	AuthorID uint   `json:"authorId"`
	SortBy   string `json:"sortBy" binding:"oneof='' created_at updated_at published_at view_count like_count"`
	Order    string `json:"order" binding:"oneof='' asc desc"`
	Search   string `json:"search"`
}

type ArticleListResponse struct {
	Articles []*model.Article `json:"articles"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// ArticleService 文章服务实现
type ArticleService struct {
	articleRepo repository.ArticleRepositoryInterface
	userRepo    repository.UserRepository
	rbacService RBACService
}

// NewArticleService 创建文章服务实例
func NewArticleService(
	articleRepo repository.ArticleRepositoryInterface,
	userRepo repository.UserRepository,
	rbacService RBACService,
) ArticleServiceInterface {
	return &ArticleService{
		articleRepo: articleRepo,
		userRepo:    userRepo,
		rbacService: rbacService,
	}
}

// CreateArticle 创建文章
func (s *ArticleService) CreateArticle(req *CreateArticleRequest, authorID uint) (*model.Article, error) {
	// 获取用户信息以检查角色权限
	user, err := s.userRepo.GetByID(authorID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 权限检查
	if !s.rbacService.HasPermission(user.Role, PermissionArticleCreate) {
		return nil, errors.New("没有创建文章的权限")
	}

	// 构建文章对象
	article := &model.Article{
		Title:          req.Title,
		Slug:           req.Slug,
		Summary:        req.Summary,
		Content:        req.Content,
		CoverImage:     req.CoverImage,
		AuthorID:       authorID,
		CategoryID:     req.CategoryID,
		Status:         model.ArticleStatus(req.Status),
		IsFeatured:     req.IsFeatured,
		IsTop:          req.IsTop,
		CommentEnabled: req.CommentEnabled,
		SEOTitle:       req.SEOTitle,
		SEODescription: req.SEODescription,
		SEOKeywords:    req.SEOKeywords,
	}

	// 处理内容
	if err := s.processContent(article); err != nil {
		return nil, err
	}

	// 创建文章
	if err := s.articleRepo.Create(article); err != nil {
		return nil, err
	}

	// 同步分类关联
	if len(req.CategoryIDs) > 0 {
		if err := s.articleRepo.SyncCategories(article.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	// 同步标签关联
	if len(req.TagIDs) > 0 {
		if err := s.articleRepo.SyncTags(article.ID, req.TagIDs); err != nil {
			return nil, err
		}
	}

	// 重新获取完整的文章信息
	return s.articleRepo.GetByID(article.ID)
}

// GetArticle 获取文章详情
func (s *ArticleService) GetArticle(id uint, userID *uint) (*model.Article, error) {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 权限检查
	if !s.CanView(article, userID) {
		return nil, errors.New("没有查看此文章的权限")
	}

	return article, nil
}

// GetArticleBySlug 根据Slug获取文章
func (s *ArticleService) GetArticleBySlug(slug string, userID *uint) (*model.Article, error) {
	article, err := s.articleRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	// 权限检查
	if !s.CanView(article, userID) {
		return nil, errors.New("没有查看此文章的权限")
	}

	return article, nil
}

// UpdateArticle 更新文章
func (s *ArticleService) UpdateArticle(id uint, req *UpdateArticleRequest, userID uint) (*model.Article, error) {
	// 获取现有文章
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 权限检查
	if !s.CanEdit(article, userID) {
		return nil, errors.New("没有编辑此文章的权限")
	}

	// 更新字段
	article.Title = req.Title
	article.Slug = req.Slug
	article.Summary = req.Summary
	article.Content = req.Content
	article.CoverImage = req.CoverImage
	article.CategoryID = req.CategoryID
	article.Status = model.ArticleStatus(req.Status)
	article.IsFeatured = req.IsFeatured
	article.IsTop = req.IsTop
	article.CommentEnabled = req.CommentEnabled
	article.SEOTitle = req.SEOTitle
	article.SEODescription = req.SEODescription
	article.SEOKeywords = req.SEOKeywords

	// 处理内容
	if err := s.processContent(article); err != nil {
		return nil, err
	}

	// 更新文章
	if err := s.articleRepo.Update(article); err != nil {
		return nil, err
	}

	// 同步分类关联
	if len(req.CategoryIDs) > 0 {
		if err := s.articleRepo.SyncCategories(article.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	// 同步标签关联
	if len(req.TagIDs) > 0 {
		if err := s.articleRepo.SyncTags(article.ID, req.TagIDs); err != nil {
			return nil, err
		}
	}

	// 重新获取完整的文章信息
	return s.articleRepo.GetByID(article.ID)
}

// DeleteArticle 删除文章
func (s *ArticleService) DeleteArticle(id uint, userID uint) error {
	// 获取现有文章
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 权限检查
	if !s.CanDelete(article, userID) {
		return errors.New("没有删除此文章的权限")
	}

	return s.articleRepo.Delete(id)
}

// GetArticleList 获取文章列表
func (s *ArticleService) GetArticleList(req *GetArticleListRequest, userID *uint) (*ArticleListResponse, error) {
	params := &repository.ArticleListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   model.ArticleStatus(req.Status),
		AuthorID: req.AuthorID,
		SortBy:   req.SortBy,
		Order:    req.Order,
		Search:   req.Search,
	}

	// 如果不是管理员，只能看到已发布的文章
	if userID == nil {
		params.Status = model.ArticleStatusPublished
	} else {
		user, err := s.userRepo.GetByID(*userID)
		if err != nil || !s.rbacService.HasPermission(user.Role, PermissionArticleManage) {
			params.Status = model.ArticleStatusPublished
		}
	}

	articles, total, err := s.articleRepo.List(params)
	if err != nil {
		return nil, err
	}

	// 过滤用户没有权限查看的文章
	var filteredArticles []*model.Article
	for _, article := range articles {
		if s.CanView(article, userID) {
			filteredArticles = append(filteredArticles, article)
		}
	}

	return &ArticleListResponse{
		Articles: filteredArticles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetArticlesByAuthor 获取指定作者的文章
func (s *ArticleService) GetArticlesByAuthor(authorID uint, req *GetArticleListRequest) (*ArticleListResponse, error) {
	params := &repository.ArticleListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   model.ArticleStatus(req.Status),
		SortBy:   req.SortBy,
		Order:    req.Order,
		Search:   req.Search,
	}

	articles, total, err := s.articleRepo.GetByAuthor(authorID, params)
	if err != nil {
		return nil, err
	}

	return &ArticleListResponse{
		Articles: articles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetArticlesByCategory 获取指定分类的文章
func (s *ArticleService) GetArticlesByCategory(categoryID uint, req *GetArticleListRequest) (*ArticleListResponse, error) {
	params := &repository.ArticleListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   model.ArticleStatusPublished,
		SortBy:   req.SortBy,
		Order:    req.Order,
		Search:   req.Search,
	}

	articles, total, err := s.articleRepo.GetByCategory(categoryID, params)
	if err != nil {
		return nil, err
	}

	return &ArticleListResponse{
		Articles: articles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetArticlesByTag 获取指定标签的文章
func (s *ArticleService) GetArticlesByTag(tagID uint, req *GetArticleListRequest) (*ArticleListResponse, error) {
	params := &repository.ArticleListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   model.ArticleStatusPublished,
		SortBy:   req.SortBy,
		Order:    req.Order,
		Search:   req.Search,
	}

	articles, total, err := s.articleRepo.GetByTag(tagID, params)
	if err != nil {
		return nil, err
	}

	return &ArticleListResponse{
		Articles: articles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// SearchArticles 搜索文章
func (s *ArticleService) SearchArticles(keyword string, req *GetArticleListRequest) (*ArticleListResponse, error) {
	params := &repository.ArticleListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   model.ArticleStatusPublished,
		SortBy:   req.SortBy,
		Order:    req.Order,
	}

	articles, total, err := s.articleRepo.Search(keyword, params)
	if err != nil {
		return nil, err
	}

	return &ArticleListResponse{
		Articles: articles,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetPopularArticles 获取热门文章
func (s *ArticleService) GetPopularArticles(limit int) ([]*model.Article, error) {
	return s.articleRepo.GetPopular(limit)
}

// GetRecentArticles 获取最新文章
func (s *ArticleService) GetRecentArticles(limit int) ([]*model.Article, error) {
	return s.articleRepo.GetRecent(limit)
}

// GetRelatedArticles 获取相关文章
func (s *ArticleService) GetRelatedArticles(articleID uint, limit int) ([]*model.Article, error) {
	// 获取当前文章
	article, err := s.articleRepo.GetByID(articleID)
	if err != nil {
		return nil, err
	}

	// 基于分类和标签获取相关文章
	params := &repository.ArticleListParams{
		Page:     1,
		PageSize: limit * 2, // 获取更多文章用于筛选
		Status:   model.ArticleStatusPublished,
		SortBy:   "view_count",
		Order:    "desc",
	}

	var relatedArticles []*model.Article

	// 优先获取同分类的文章
	if article.CategoryID != nil {
		articles, _, err := s.articleRepo.GetByCategory(*article.CategoryID, params)
		if err == nil {
			for _, a := range articles {
				if a.ID != articleID && len(relatedArticles) < limit {
					relatedArticles = append(relatedArticles, a)
				}
			}
		}
	}

	// 如果还不够，获取同标签的文章
	if len(relatedArticles) < limit && len(article.Tags) > 0 {
		for _, tag := range article.Tags {
			if len(relatedArticles) >= limit {
				break
			}

			articles, _, err := s.articleRepo.GetByTag(tag.ID, params)
			if err != nil {
				continue
			}

			for _, a := range articles {
				if a.ID != articleID && !containsArticle(relatedArticles, a.ID) && len(relatedArticles) < limit {
					relatedArticles = append(relatedArticles, a)
				}
			}
		}
	}

	return relatedArticles, nil
}

// ViewArticle 记录文章浏览
func (s *ArticleService) ViewArticle(articleID uint, userID *uint, visitorID string, ipAddress string) error {
	// 增加浏览量
	if err := s.articleRepo.IncrementViewCount(articleID); err != nil {
		return err
	}

	// TODO: 记录详细的浏览记录到 article_views 表
	// 这里可以异步处理，避免影响响应速度

	return nil
}

// LikeArticle 点赞文章
func (s *ArticleService) LikeArticle(articleID uint, userID uint) error {
	// TODO: 实现点赞逻辑，操作 article_likes 表
	// 更新文章的点赞数
	return s.articleRepo.UpdateLikeCount(articleID)
}

// UnlikeArticle 取消点赞文章
func (s *ArticleService) UnlikeArticle(articleID uint, userID uint) error {
	// TODO: 实现取消点赞逻辑
	return s.articleRepo.UpdateLikeCount(articleID)
}

// BookmarkArticle 收藏文章
func (s *ArticleService) BookmarkArticle(articleID uint, userID uint) error {
	// TODO: 实现收藏逻辑，操作 article_bookmarks 表
	return nil
}

// UnbookmarkArticle 取消收藏文章
func (s *ArticleService) UnbookmarkArticle(articleID uint, userID uint) error {
	// TODO: 实现取消收藏逻辑
	return nil
}

// PublishArticle 发布文章
func (s *ArticleService) PublishArticle(id uint, userID uint) error {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if !s.CanEdit(article, userID) {
		return errors.New("没有编辑此文章的权限")
	}

	return s.articleRepo.Publish(id)
}

// UnpublishArticle 取消发布文章
func (s *ArticleService) UnpublishArticle(id uint, userID uint) error {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if !s.CanEdit(article, userID) {
		return errors.New("没有编辑此文章的权限")
	}

	return s.articleRepo.Unpublish(id)
}

// ArchiveArticle 归档文章
func (s *ArticleService) ArchiveArticle(id uint, userID uint) error {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if !s.CanEdit(article, userID) {
		return errors.New("没有编辑此文章的权限")
	}

	return s.articleRepo.Archive(id)
}

// SetArticlePrivate 设置文章为私有
func (s *ArticleService) SetArticlePrivate(id uint, userID uint) error {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if !s.CanEdit(article, userID) {
		return errors.New("没有编辑此文章的权限")
	}

	return s.articleRepo.SetPrivate(id)
}

// CanView 检查用户是否可以查看文章
func (s *ArticleService) CanView(article *model.Article, userID *uint) bool {
	// 已发布的文章所有人都可以查看
	if article.Status == model.ArticleStatusPublished {
		return true
	}

	// 未登录用户只能查看已发布的文章
	if userID == nil {
		return false
	}

	// 作者可以查看自己的所有文章
	if article.AuthorID == *userID {
		return true
	}

	// 管理员可以查看所有文章
	user, err := s.userRepo.GetByID(*userID)
	if err == nil && s.rbacService.HasPermission(user.Role, PermissionArticleManage) {
		return true
	}

	return false
}

// CanEdit 检查用户是否可以编辑文章
func (s *ArticleService) CanEdit(article *model.Article, userID uint) bool {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false
	}

	// 作者可以编辑自己的文章
	if article.AuthorID == userID {
		return s.rbacService.HasPermission(user.Role, PermissionArticleCreate)
	}

	// 管理员可以编辑所有文章
	return s.rbacService.HasPermission(user.Role, PermissionArticleManage)
}

// CanDelete 检查用户是否可以删除文章
func (s *ArticleService) CanDelete(article *model.Article, userID uint) bool {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false
	}

	// 作者可以删除自己的文章
	if article.AuthorID == userID {
		return s.rbacService.HasPermission(user.Role, PermissionArticleCreate)
	}

	// 管理员可以删除所有文章
	return s.rbacService.HasPermission(user.Role, PermissionArticleManage)
}

// 私有辅助方法

// processContent 处理文章内容
func (s *ArticleService) processContent(article *model.Article) error {
	// 清理和转义内容
	article.Content = html.EscapeString(article.Content)
	article.Summary = html.EscapeString(article.Summary)

	// 计算字数
	article.WordCount = uint(len(strings.Fields(article.Content)))

	// 计算阅读时间（假设每分钟200字）
	article.ReadingTime = article.WordCount / 200
	if article.ReadingTime == 0 {
		article.ReadingTime = 1
	}

	// 如果没有摘要，从内容中提取
	if article.Summary == "" {
		article.Summary = s.extractSummary(article.Content, 200)
	}

	return nil
}

// extractSummary 从内容中提取摘要
func (s *ArticleService) extractSummary(content string, maxLength int) string {
	// 移除HTML标签
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(content, "")

	// 限制长度
	if len(text) > maxLength {
		text = text[:maxLength] + "..."
	}

	return text
}

// containsArticle 检查文章数组是否包含指定ID的文章
func containsArticle(articles []*model.Article, id uint) bool {
	for _, article := range articles {
		if article.ID == id {
			return true
		}
	}
	return false
}
