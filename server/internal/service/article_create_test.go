package service

import (
	"errors"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeCreateArticleRepo 创建文章场景的仓储替身，覆盖创建与关联同步方法。
type fakeCreateArticleRepo struct {
	repository.ArticleRepositoryInterface
	createErr       error
	createdArticle  *model.Article
	lastCategoryIDs []uint
	lastTagIDs      []uint
}

func (f *fakeCreateArticleRepo) CreateWithRelations(article *model.Article, categoryIDs, tagIDs []uint) error {
	if f.createErr != nil {
		return f.createErr
	}
	article.ID = 1
	f.createdArticle = article
	f.lastCategoryIDs = categoryIDs
	f.lastTagIDs = tagIDs
	return nil
}

func (f *fakeCreateArticleRepo) GetByID(id uint) (*model.Article, error) {
	if f.createdArticle == nil {
		return nil, repository.ErrArticleNotFound
	}
	return f.createdArticle, nil
}

// newCreateArticleService 创建注入创建场景替身的文章服务实例。
func newCreateArticleService(articleRepo *fakeCreateArticleRepo) *ArticleService {
	userRepo := &fakeUserRepo{user: &domain.User{ID: 1, Username: "admin", Role: "admin", Status: 1}}
	return NewArticleService(articleRepo, userRepo, NewRBACService(), &fakeStatsRepo{}, &fakeNotificationRepo{}).(*ArticleService)
}

// TestCreateArticleSyncFailurePropagates 验证关联同步失败时创建文章返回错误。
// 事务化后该错误由仓储层回滚保证，service 层不再产生孤儿文章。
func TestCreateArticleSyncFailurePropagates(t *testing.T) {
	articleRepo := &fakeCreateArticleRepo{createErr: errors.New("分类关联写入失败")}
	svc := newCreateArticleService(articleRepo)

	_, err := svc.CreateArticle(&CreateArticleRequest{
		Title:       "测试文章",
		Content:     "正文内容",
		CategoryIDs: []uint{1},
	}, 1)

	if err == nil {
		t.Fatal("关联同步失败时 CreateArticle 应返回错误")
	}
	if articleRepo.createdArticle != nil {
		t.Error("失败路径不应残留已创建文章")
	}
}

// TestCreateArticleDelegatesRelations 验证创建成功后分类与标签关联经事务方法写入。
func TestCreateArticleDelegatesRelations(t *testing.T) {
	articleRepo := &fakeCreateArticleRepo{}
	svc := newCreateArticleService(articleRepo)

	article, err := svc.CreateArticle(&CreateArticleRequest{
		Title:       "测试文章",
		Content:     "正文内容",
		CategoryIDs: []uint{2, 3},
		TagIDs:      []uint{4},
		Status:      "draft",
	}, 1)
	if err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	if article.ID != 1 {
		t.Errorf("文章 ID = %d，期望 1", article.ID)
	}
	if len(articleRepo.lastCategoryIDs) != 2 || articleRepo.lastCategoryIDs[0] != 2 {
		t.Errorf("分类关联未透传: %v", articleRepo.lastCategoryIDs)
	}
	if len(articleRepo.lastTagIDs) != 1 || articleRepo.lastTagIDs[0] != 4 {
		t.Errorf("标签关联未透传: %v", articleRepo.lastTagIDs)
	}
}
