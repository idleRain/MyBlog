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
	return NewCommentService(commentRepo, articleRepo).(*CommentService)
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
	_, err := svc.CreateComment(req)
	if err == nil {
		t.Fatal("游客评论未填写姓名应返回错误")
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
	_, err := svc.CreateComment(req)
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
	comment, err := svc.CreateComment(req)
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
