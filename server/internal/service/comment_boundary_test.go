package service

import (
	"testing"

	"MyBlog/internal/model"
)

// publishedCommentArticleRepo 构造按 ID 返回可评论已发布文章的文章仓储替身。
func publishedCommentArticleRepo() *fakeArticleRepo {
	return &fakeArticleRepo{
		getByID: func(id uint) (*model.Article, error) {
			return publishedArticleForComment(id), nil
		},
	}
}

// TestCreateCommentHoistsNestedReplyToRoot 回复二级评论时归位到根评论下，
// 树深度锁定两级，嵌套深度不会随回复链增长。
func TestCreateCommentHoistsNestedReplyToRoot(t *testing.T) {
	rootID := uint(1)
	root := &model.Comment{ID: rootID, ArticleID: 1, Level: 1, Status: model.CommentStatusApproved}
	level2 := &model.Comment{ID: 2, ArticleID: 1, Level: 2, Status: model.CommentStatusApproved, RootID: &rootID}
	repo := &fakeCommentRepo{comments: []*model.Comment{root, level2}}
	svc := commentTestService(repo, publishedCommentArticleRepo())

	reply, err := svc.CreateComment(&CreateCommentRequest{
		ArticleID:  1,
		ParentID:   &level2.ID,
		Content:    "对二级评论的回复",
		AuthorName: "游客甲",
	}, nil)
	if err != nil {
		t.Fatalf("创建二级回复失败: %v", err)
	}

	if reply.Level != 2 {
		t.Errorf("回复层级 = %d, 期望锁定为二级", reply.Level)
	}
	if reply.RootID == nil || *reply.RootID != rootID {
		t.Errorf("回复根 ID = %v, 期望归位到根评论 %d", reply.RootID, rootID)
	}
}

// TestCreateCommentRejectsParentFromOtherArticle 父评论归属其他文章时拒绝回复。
func TestCreateCommentRejectsParentFromOtherArticle(t *testing.T) {
	parent := &model.Comment{ID: 5, ArticleID: 9, Level: 1, Status: model.CommentStatusApproved}
	repo := &fakeCommentRepo{comments: []*model.Comment{parent}}
	svc := commentTestService(repo, publishedCommentArticleRepo())

	_, err := svc.CreateComment(&CreateCommentRequest{
		ArticleID:  1,
		ParentID:   &parent.ID,
		Content:    "跨文章回复",
		AuthorName: "游客甲",
	}, nil)
	if err == nil {
		t.Fatal("跨文章回复应被拒绝")
	}
}
