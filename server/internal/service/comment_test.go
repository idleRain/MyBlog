package service

import (
	"testing"
	"time"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// modelTime 用于构造已发布文章的测试时间。
var modelTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// fakeCommentRepo 评论仓储的测试替身。
type fakeCommentRepo struct {
	repository.CommentRepositoryInterface
	comments       []*model.Comment
	getByID        func(id uint) (*model.Comment, error)
	updateStatus   func(id uint, status model.CommentStatus) error
	incrementReply func(id uint) error
	decrementReply func(id uint) error
}

func (f *fakeCommentRepo) GetByID(id uint) (*model.Comment, error) {
	if f.getByID != nil {
		return f.getByID(id)
	}
	for _, comment := range f.comments {
		if comment.ID == id {
			return comment, nil
		}
	}
	return nil, repository.ErrCommentNotFound
}

func (f *fakeCommentRepo) Create(comment *model.Comment) error {
	comment.ID = uint(len(f.comments) + 1)
	f.comments = append(f.comments, comment)
	return nil
}

func (f *fakeCommentRepo) UpdateStatus(id uint, status model.CommentStatus) error {
	if f.updateStatus != nil {
		return f.updateStatus(id, status)
	}
	for _, comment := range f.comments {
		if comment.ID == id {
			comment.Status = status
			return nil
		}
	}
	return repository.ErrCommentNotFound
}

func (f *fakeCommentRepo) IncrementReplyCount(id uint) error {
	if f.incrementReply != nil {
		return f.incrementReply(id)
	}
	return nil
}

func (f *fakeCommentRepo) DecrementReplyCount(id uint) error {
	if f.decrementReply != nil {
		return f.decrementReply(id)
	}
	return nil
}

// commentTestService 创建注入评论与文章仓储替身的服务实例。
func commentTestService(commentRepo repository.CommentRepositoryInterface, articleRepo repository.ArticleRepositoryInterface) *CommentService {
	return NewCommentService(commentRepo, articleRepo, &fakeSettingRepo{}).(*CommentService)
}

// commentTestServiceWithSettings 创建携带站点设置替身的服务实例，设置键缺失时走默认行为。
func commentTestServiceWithSettings(
	commentRepo repository.CommentRepositoryInterface,
	articleRepo repository.ArticleRepositoryInterface,
	settings map[string]string,
) *CommentService {
	return NewCommentService(commentRepo, articleRepo, settingRepoWith(settings)).(*CommentService)
}

// settingRepoWith 构造携带键值设置的设置仓储替身。
func settingRepoWith(settings map[string]string) *fakeSettingRepo {
	repo := &fakeSettingRepo{}
	for key, value := range settings {
		repo.settings = append(repo.settings, &model.Setting{KeyName: key, Value: value})
	}
	return repo
}

// publishedArticleForComment 构造允许评论的已发布文章。
func publishedArticleForComment(id uint) *model.Article {
	return &model.Article{
		ID:             id,
		Title:          "测试文章",
		Status:         model.ArticleStatusPublished,
		CommentEnabled: true,
		PublishedAt:    &modelTime,
	}
}

// TestCreateCommentGuestRequiresName 验证游客评论必须提供姓名。
func TestCreateCommentGuestRequiresName(t *testing.T) {
	articleRepo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticleForComment(id), nil
		},
	}
	svc := commentTestService(&fakeCommentRepo{}, articleRepo)

	req := &CreateCommentRequest{
		ArticleID: 1,
		Content:   "这是一条评论",
	}
	_, err := svc.CreateComment(req, nil)
	if err == nil {
		t.Fatal("游客评论未填写姓名应返回错误")
	}
}

// TestCreateCommentBindsLoginIdentity 验证登录用户评论绑定身份且忽略游客字段。
func TestCreateCommentBindsLoginIdentity(t *testing.T) {
	articleRepo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticleForComment(id), nil
		},
	}
	commentRepo := &fakeCommentRepo{}
	svc := commentTestService(commentRepo, articleRepo)

	userID := uint(7)
	req := &CreateCommentRequest{
		ArticleID:  1,
		Content:    "登录用户评论",
		AuthorName: "伪造姓名",
	}
	comment, err := svc.CreateComment(req, &userID)
	if err != nil {
		t.Fatalf("登录评论创建失败: %v", err)
	}
	if comment.UserID == nil || *comment.UserID != 7 {
		t.Errorf("UserID = %v, 期望 7", comment.UserID)
	}
	if comment.AuthorName != "" {
		t.Errorf("登录评论的游客姓名应被忽略，实际为 %s", comment.AuthorName)
	}
	if comment.Status != model.CommentStatusPending {
		t.Errorf("Status = %s, 期望 pending", comment.Status)
	}
}

// TestCreateCommentParentMismatch 验证回复评论的文章归属校验。
func TestCreateCommentParentMismatch(t *testing.T) {
	articleRepo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticleForComment(id), nil
		},
	}
	parentID := uint(100)
	commentRepo := &fakeCommentRepo{
		comments: []*model.Comment{
			{ID: 100, ArticleID: 999, Content: "其他文章评论"},
		},
	}
	svc := commentTestService(commentRepo, articleRepo)

	req := &CreateCommentRequest{
		ArticleID:  1,
		ParentID:   &parentID,
		Content:    "回复评论",
		AuthorName: "游客甲",
	}
	_, err := svc.CreateComment(req, nil)
	if err == nil {
		t.Fatal("父评论与文章不匹配应返回错误")
	}
}

// TestCreateReplySetsLevel 验证回复根评论时派生 root 与 level 字段。
func TestCreateReplySetsLevel(t *testing.T) {
	articleRepo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticleForComment(id), nil
		},
	}
	parentID := uint(100)
	commentRepo := &fakeCommentRepo{
		comments: []*model.Comment{
			{ID: 100, ArticleID: 1, Content: "根评论", Level: 1},
		},
	}
	svc := commentTestService(commentRepo, articleRepo)

	req := &CreateCommentRequest{
		ArticleID:  1,
		ParentID:   &parentID,
		Content:    "回复根评论",
		AuthorName: "游客甲",
	}
	comment, err := svc.CreateComment(req, nil)
	if err != nil {
		t.Fatalf("创建回复评论失败: %v", err)
	}

	if comment.RootID == nil || *comment.RootID != 100 {
		t.Errorf("RootID = %v, 期望指向根评论 100", comment.RootID)
	}
	if comment.Level != 2 {
		t.Errorf("Level = %d, 期望 2", comment.Level)
	}
}

// TestCreateCommentRejectsGuestWhenDisabled 验证站点关闭游客评论后游客通道被拒。
func TestCreateCommentRejectsGuestWhenDisabled(t *testing.T) {
	articleRepo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticleForComment(id), nil
		},
	}
	svc := commentTestServiceWithSettings(&fakeCommentRepo{}, articleRepo, map[string]string{
		model.SettingAllowGuestComment: "false",
	})

	req := &CreateCommentRequest{
		ArticleID:  1,
		Content:    "游客评论",
		AuthorName: "游客甲",
	}
	if _, err := svc.CreateComment(req, nil); err == nil {
		t.Fatal("关闭游客评论后应拒绝游客发言")
	}
}

// TestCreateCommentAutoApprove 验证站点开启自动通过后评论直接进入已审核状态。
func TestCreateCommentAutoApprove(t *testing.T) {
	articleRepo := &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticleForComment(id), nil
		},
	}
	commentRepo := &fakeCommentRepo{}
	svc := commentTestServiceWithSettings(commentRepo, articleRepo, map[string]string{
		model.SettingCommentAutoApprove: "true",
	})

	userID := uint(7)
	comment, err := svc.CreateComment(&CreateCommentRequest{ArticleID: 1, Content: "登录评论"}, &userID)
	if err != nil {
		t.Fatalf("创建评论失败: %v", err)
	}
	if comment.Status != model.CommentStatusApproved {
		t.Errorf("Status = %s, 期望 approved", comment.Status)
	}
}

// TestApproveCommentStatus 验证审核通过更新评论状态。
func TestApproveCommentStatus(t *testing.T) {
	commentRepo := &fakeCommentRepo{
		comments: []*model.Comment{
			{ID: 1, ArticleID: 1, Content: "待审核评论", Status: model.CommentStatusPending},
		},
	}
	svc := commentTestService(commentRepo, &fakeArticleRepo{})

	if err := svc.ApproveComment(1, 1); err != nil {
		t.Fatalf("审核通过失败: %v", err)
	}

	if commentRepo.comments[0].Status != model.CommentStatusApproved {
		t.Errorf("审核后状态 = %s, 期望 approved", commentRepo.comments[0].Status)
	}
}
