package service

import (
	"errors"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeArticleRepo 文章仓储的测试替身，记录互动调用并返回可配置结果。
type fakeArticleRepo struct {
	repository.ArticleRepositoryInterface
	getByID        func(id uint) (*model.Article, error)
	getByAuthor    func(authorID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error)
	search         func(keyword string, params *repository.ArticleListParams) ([]*model.Article, int64, error)
	incrementView  func(id uint) error
	recordView     func(view *model.ArticleView) error
	addLike        func(articleID, userID uint) (bool, error)
	removeLike     func(articleID, userID uint) (bool, error)
	addBookmark    func(articleID, userID uint) (bool, error)
	removeBookmark func(articleID, userID uint) (bool, error)
	existsLike     func(articleID, userID uint) (bool, error)
	existsBookmark func(articleID, userID uint) (bool, error)
	listBookmarks  func(userID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error)
}

func (f *fakeArticleRepo) ExistsLike(articleID, userID uint) (bool, error) {
	if f.existsLike != nil {
		return f.existsLike(articleID, userID)
	}
	return false, errors.New("未实现的测试替身方法")
}

func (f *fakeArticleRepo) ExistsBookmark(articleID, userID uint) (bool, error) {
	if f.existsBookmark != nil {
		return f.existsBookmark(articleID, userID)
	}
	return false, errors.New("未实现的测试替身方法")
}

func (f *fakeArticleRepo) ListBookmarks(userID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
	if f.listBookmarks != nil {
		return f.listBookmarks(userID, params)
	}
	return nil, 0, errors.New("未实现的测试替身方法")
}

func (f *fakeArticleRepo) Search(keyword string, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
	if f.search != nil {
		return f.search(keyword, params)
	}
	return nil, 0, errors.New("未实现的测试替身方法")
}

func (f *fakeArticleRepo) GetByID(id uint) (*model.Article, error) {
	if f.getByID != nil {
		return f.getByID(id)
	}
	return nil, errors.New("未实现的测试替身方法")
}

func (f *fakeArticleRepo) GetByAuthor(authorID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
	if f.getByAuthor != nil {
		return f.getByAuthor(authorID, params)
	}
	return nil, 0, errors.New("未实现的测试替身方法")
}

func (f *fakeArticleRepo) AddLike(articleID, userID uint) (bool, error) {
	if f.addLike != nil {
		return f.addLike(articleID, userID)
	}
	return false, nil
}

func (f *fakeArticleRepo) RemoveLike(articleID, userID uint) (bool, error) {
	if f.removeLike != nil {
		return f.removeLike(articleID, userID)
	}
	return false, nil
}

func (f *fakeArticleRepo) AddBookmark(articleID, userID uint) (bool, error) {
	if f.addBookmark != nil {
		return f.addBookmark(articleID, userID)
	}
	return false, nil
}

func (f *fakeArticleRepo) RemoveBookmark(articleID, userID uint) (bool, error) {
	if f.removeBookmark != nil {
		return f.removeBookmark(articleID, userID)
	}
	return false, nil
}

// Archive 记录归档操作，测试替身默认成功。
func (f *fakeArticleRepo) Archive(id uint) error {
	return nil
}

// IncrementViewCount 递增浏览计数，测试替身默认成功。
func (f *fakeArticleRepo) IncrementViewCount(id uint) error {
	if f.incrementView != nil {
		return f.incrementView(id)
	}
	return nil
}

// RecordArticleView 记录浏览明细，测试替身默认成功。
func (f *fakeArticleRepo) RecordArticleView(view *model.ArticleView) error {
	if f.recordView != nil {
		return f.recordView(view)
	}
	return nil
}

// captureAuthorStatus 构造捕获查询状态的作者文章测试替身，供可见性断言使用。
func captureAuthorStatus(captured *model.ArticleStatus) *fakeArticleRepo {
	return &fakeArticleRepo{
		getByAuthor: func(authorID uint, params *repository.ArticleListParams) ([]*model.Article, int64, error) {
			*captured = params.Status
			return []*model.Article{publishedArticle(1)}, 1, nil
		},
	}
}

// fakeUserRepo 用户仓储的测试替身。
type fakeUserRepo struct {
	repository.UserRepository
	user *domain.User
}

func (f *fakeUserRepo) GetByID(id uint) (*domain.User, error) {
	if f.user == nil {
		return nil, errors.New("用户不存在")
	}
	// 校验目标用户 ID 一致，模拟不存在的用户查询。
	if f.user.ID != id {
		return nil, errors.New("用户不存在")
	}
	return f.user, nil
}

// publishedArticle 构造一篇已发布文章的测试数据。
func publishedArticle(id uint) *model.Article {
	return &model.Article{
		ID:     id,
		Title:  "测试文章",
		Status: model.ArticleStatusPublished,
	}
}

// newArticleTestService 创建注入测试替身的文章服务实例。
func newArticleTestService(articleRepo repository.ArticleRepositoryInterface) *ArticleService {
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(articleRepo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{})
	return svc.(*ArticleService)
}

// TestLikeArticleCallsAddLike 验证点赞时调用仓储新增并校验文章可见性。
func TestLikeArticleCallsAddLike(t *testing.T) {
	calledAddLike := false
	repo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticle(id), nil
		},
		addLike: func(articleID, userID uint) (bool, error) {
			calledAddLike = true
			return true, nil
		},
	}
	svc := newArticleTestService(repo)

	if err := svc.LikeArticle(1, 1); err != nil {
		t.Fatalf("点赞失败: %v", err)
	}
	if !calledAddLike {
		t.Error("点赞应调用仓储 AddLike")
	}
}

// TestLikeArticleRejectsInvisibleArticle 验证对不可见文章点赞返回权限错误且不落库。
func TestLikeArticleRejectsInvisibleArticle(t *testing.T) {
	calledAddLike := false
	repo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			// 私有文章，普通用户不可见。
			return &model.Article{ID: id, Status: model.ArticleStatusPrivate}, nil
		},
		addLike: func(articleID, userID uint) (bool, error) {
			calledAddLike = true
			return true, nil
		},
	}
	// 使用普通用户角色的服务实例，验证权限校验。
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "user", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)

	err := svc.LikeArticle(1, 1)
	if err == nil {
		t.Fatal("对不可见文章点赞应返回错误")
	}
	if calledAddLike {
		t.Error("不可见文章不应调用仓储 AddLike")
	}
}

// TestLikeArticleNotFound 验证文章不存在时返回仓储错误。
func TestLikeArticleNotFound(t *testing.T) {
	repo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return nil, repository.ErrArticleNotFound
		},
	}
	svc := newArticleTestService(repo)

	if err := svc.LikeArticle(999, 1); !errors.Is(err, repository.ErrArticleNotFound) {
		t.Errorf("点赞不存在文章应返回 ErrArticleNotFound，实际为 %v", err)
	}
}

// TestUnlikeArticleCallsRemoveLike 验证取消点赞时调用仓储移除。
func TestUnlikeArticleCallsRemoveLike(t *testing.T) {
	calledRemoveLike := false
	repo := &fakeArticleRepo{
		removeLike: func(articleID, userID uint) (bool, error) {
			calledRemoveLike = true
			return true, nil
		},
	}
	svc := newArticleTestService(repo)

	if err := svc.UnlikeArticle(1, 1); err != nil {
		t.Fatalf("取消点赞失败: %v", err)
	}
	if !calledRemoveLike {
		t.Error("取消点赞应调用仓储 RemoveLike")
	}
}

// TestLikeArticleNotifiesAuthor 验证首次点赞通知文章作者。
func TestLikeArticleNotifiesAuthor(t *testing.T) {
	notificationRepo := &fakeNotificationRepo{}
	repo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return &model.Article{ID: id, Title: "文章", Slug: "post", Status: model.ArticleStatusPublished, AuthorID: 2}, nil
		},
		addLike: func(articleID, userID uint) (bool, error) {
			return true, nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, notificationRepo).(*ArticleService)

	if err := svc.LikeArticle(1, 1); err != nil {
		t.Fatalf("点赞失败: %v", err)
	}
	if len(notificationRepo.created) != 1 {
		t.Fatalf("通知数 = %d, 期望 1", len(notificationRepo.created))
	}
	if notificationRepo.created[0].UserID != 2 {
		t.Errorf("通知接收者 = %d, 期望 2", notificationRepo.created[0].UserID)
	}
	if notificationRepo.created[0].Type != model.NotificationTypeArticleLike {
		t.Errorf("通知类型 = %s, 期望 article_like", notificationRepo.created[0].Type)
	}
}

// TestLikeArticleSelfLikeSkipsNotification 验证作者点赞自己的文章不产生通知。
func TestLikeArticleSelfLikeSkipsNotification(t *testing.T) {
	notificationRepo := &fakeNotificationRepo{}
	repo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return &model.Article{ID: id, Title: "文章", Slug: "post", Status: model.ArticleStatusPublished, AuthorID: 1}, nil
		},
		addLike: func(articleID, userID uint) (bool, error) {
			return true, nil
		},
	}
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Role: "admin", Status: 1}}
	svc := NewArticleService(repo, userRepo, NewRBACService(), &fakeStatsRepo{}, notificationRepo).(*ArticleService)

	if err := svc.LikeArticle(1, 1); err != nil {
		t.Fatalf("点赞失败: %v", err)
	}
	if len(notificationRepo.created) != 0 {
		t.Errorf("通知数 = %d, 期望 0", len(notificationRepo.created))
	}
}

// TestBookmarkArticleCallsAddBookmark 验证收藏时调用仓储新增。
func TestBookmarkArticleCallsAddBookmark(t *testing.T) {
	calledAddBookmark := false
	repo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticle(id), nil
		},
		addBookmark: func(articleID, userID uint) (bool, error) {
			calledAddBookmark = true
			return true, nil
		},
	}
	svc := newArticleTestService(repo)

	if err := svc.BookmarkArticle(1, 1); err != nil {
		t.Fatalf("收藏失败: %v", err)
	}
	if !calledAddBookmark {
		t.Error("收藏应调用仓储 AddBookmark")
	}
}

// TestUnbookmarkArticleCallsRemoveBookmark 验证取消收藏时调用仓储移除。
func TestUnbookmarkArticleCallsRemoveBookmark(t *testing.T) {
	calledRemoveBookmark := false
	repo := &fakeArticleRepo{
		removeBookmark: func(articleID, userID uint) (bool, error) {
			calledRemoveBookmark = true
			return true, nil
		},
	}
	svc := newArticleTestService(repo)

	if err := svc.UnbookmarkArticle(1, 1); err != nil {
		t.Fatalf("取消收藏失败: %v", err)
	}
	if !calledRemoveBookmark {
		t.Error("取消收藏应调用仓储 RemoveBookmark")
	}
}
