package model

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCommentMarshalHidesAuditFields 验证评论公开序列化不含游客邮箱、IP 与 UserAgent 审计字段。
func TestCommentMarshalHidesAuditFields(t *testing.T) {
	comment := &Comment{
		ID:          1,
		ArticleID:   1,
		Content:     "公开评论内容",
		AuthorName:  "游客甲",
		AuthorEmail: "guest@example.com",
		AuthorIP:    "203.0.113.10",
		UserAgent:   "Mozilla/5.0 (Test)",
		Status:      CommentStatusApproved,
	}

	data, err := json.Marshal(comment)
	if err != nil {
		t.Fatalf("序列化评论失败: %v", err)
	}
	output := string(data)

	for _, forbidden := range []string{"authorEmail", "authorIP", "userAgent"} {
		if strings.Contains(output, `"`+forbidden+`"`) {
			t.Errorf("评论公开输出不应包含字段 %q，实际为 %s", forbidden, output)
		}
	}
}

// TestCommentAfterFindExposesNarrowedUser 验证评论关联用户经窄化视图输出且不含 email。
func TestCommentAfterFindExposesNarrowedUser(t *testing.T) {
	comment := &Comment{
		ID:     1,
		UserID: &[]uint{7}[0],
		User: &User{
			ID:       7,
			Username: "alice",
			Email:    "alice@example.com",
			Nickname: "爱丽丝",
			Role:     "user",
		},
	}

	if err := comment.AfterFind(nil); err != nil {
		t.Fatalf("AfterFind 不应返回错误: %v", err)
	}
	if comment.AuthorPublic == nil {
		t.Fatal("关联用户存在时公开作者视图不应为空")
	}

	data, err := json.Marshal(comment)
	if err != nil {
		t.Fatalf("序列化评论失败: %v", err)
	}
	output := string(data)

	if strings.Contains(output, `"email"`) {
		t.Errorf("评论的 user 输出不应包含 email，实际为 %s", output)
	}
	if !strings.Contains(output, `"nickname":"爱丽丝"`) {
		t.Errorf("评论的 user 输出应包含窄化后的昵称，实际为 %s", output)
	}
}

// TestCommentAfterFindWithoutUserOmitsUserKey 验证游客评论不输出 user 键。
func TestCommentAfterFindWithoutUserOmitsUserKey(t *testing.T) {
	comment := &Comment{ID: 1, Content: "游客评论"}

	if err := comment.AfterFind(nil); err != nil {
		t.Fatalf("AfterFind 不应返回错误: %v", err)
	}

	data, err := json.Marshal(comment)
	if err != nil {
		t.Fatalf("序列化评论失败: %v", err)
	}
	if strings.Contains(string(data), `"user"`) {
		t.Errorf("游客评论不应输出 user 键，实际为 %s", string(data))
	}
}

// TestArticleAfterFindExposesNarrowedAuthor 验证文章作者经窄化视图输出且不含 email。
func TestArticleAfterFindExposesNarrowedAuthor(t *testing.T) {
	article := &Article{
		ID:       1,
		Title:    "测试文章",
		AuthorID: 7,
		Author: User{
			ID:       7,
			Username: "alice",
			Email:    "alice@example.com",
			Nickname: "爱丽丝",
			Role:     "admin",
		},
	}

	if err := article.AfterFind(nil); err != nil {
		t.Fatalf("AfterFind 不应返回错误: %v", err)
	}

	data, err := json.Marshal(article)
	if err != nil {
		t.Fatalf("序列化文章失败: %v", err)
	}
	output := string(data)

	if strings.Contains(output, `"email"`) {
		t.Errorf("文章的 author 输出不应包含 email，实际为 %s", output)
	}
	if !strings.Contains(output, `"nickname":"爱丽丝"`) {
		t.Errorf("文章的 author 输出应包含窄化后的昵称，实际为 %s", output)
	}
}

// TestArticleAfterFindWithoutAuthorOmitsAuthorKey 验证作者未预加载时 author 键整体省略。
func TestArticleAfterFindWithoutAuthorOmitsAuthorKey(t *testing.T) {
	article := &Article{ID: 1, Title: "测试文章"}

	if err := article.AfterFind(nil); err != nil {
		t.Fatalf("AfterFind 不应返回错误: %v", err)
	}

	data, err := json.Marshal(article)
	if err != nil {
		t.Fatalf("序列化文章失败: %v", err)
	}
	if strings.Contains(string(data), `"author"`) {
		t.Errorf("作者未预加载时不应输出 author 键，实际为 %s", string(data))
	}
}
