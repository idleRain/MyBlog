package service

import (
	"encoding/json"
	"strings"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeLinkRepo 友情链接仓储的测试替身。
type fakeLinkRepo struct {
	repository.FriendlyLinkRepositoryInterface
	links    []*model.FriendlyLink
	getByID  func(id uint) (*model.FriendlyLink, error)
	getByURL func(url string) (*model.FriendlyLink, error)
}

// ListVisible 返回已展示链接的替身实现，显式覆写而非依赖内嵌接口的静默兼容。
func (f *fakeLinkRepo) ListVisible() ([]*model.FriendlyLink, error) {
	visible := make([]*model.FriendlyLink, 0, len(f.links))
	for _, link := range f.links {
		if link.Status == model.LinkStatusActive {
			visible = append(visible, link)
		}
	}
	return visible, nil
}

// List 返回全部链接的替身实现，管理端口径不做状态过滤。
func (f *fakeLinkRepo) List(_ *repository.FriendlyLinkListParams) ([]*model.FriendlyLink, int64, error) {
	return f.links, int64(len(f.links)), nil
}

func (f *fakeLinkRepo) GetByID(id uint) (*model.FriendlyLink, error) {
	if f.getByID != nil {
		return f.getByID(id)
	}
	for _, link := range f.links {
		if link.ID == id {
			return link, nil
		}
	}
	return nil, domain.ErrFriendlyLinkNotFound
}

func (f *fakeLinkRepo) GetByURL(url string) (*model.FriendlyLink, error) {
	if f.getByURL != nil {
		return f.getByURL(url)
	}
	for _, link := range f.links {
		if link.URL == url {
			return link, nil
		}
	}
	return nil, domain.ErrFriendlyLinkNotFound
}

func (f *fakeLinkRepo) Create(link *model.FriendlyLink) error {
	link.ID = uint(len(f.links) + 1)
	f.links = append(f.links, link)
	return nil
}

func (f *fakeLinkRepo) UpdateStatus(id uint, status model.LinkStatus) error {
	for _, link := range f.links {
		if link.ID == id {
			link.Status = status
			return nil
		}
	}
	return domain.ErrFriendlyLinkNotFound
}

// TestCreateLinkDefaultPending 验证新友情链接默认待审核。
func TestCreateLinkDefaultPending(t *testing.T) {
	repo := &fakeLinkRepo{}
	svc := NewFriendlyLinkService(repo)

	req := &CreateFriendlyLinkRequest{
		Name: "示例站点",
		URL:  "https://example.com",
	}
	link, err := svc.CreateLink(req)
	if err != nil {
		t.Fatalf("创建友情链接失败: %v", err)
	}

	if link.Status != model.LinkStatusPending {
		t.Errorf("默认状态 = %s, 期望 pending", link.Status)
	}
}

// TestCreateLinkDuplicateURL 验证重复站点 URL 被拒绝。
func TestCreateLinkDuplicateURL(t *testing.T) {
	repo := &fakeLinkRepo{
		links: []*model.FriendlyLink{
			{ID: 1, Name: "已有站点", URL: "https://example.com"},
		},
	}
	svc := NewFriendlyLinkService(repo)

	req := &CreateFriendlyLinkRequest{
		Name: "重复站点",
		URL:  "https://example.com",
	}
	_, err := svc.CreateLink(req)
	if err == nil {
		t.Fatal("重复 URL 创建应返回错误")
	}
}

// TestApproveLinkStatus 验证审核通过更新状态为 active。
func TestApproveLinkStatus(t *testing.T) {
	repo := &fakeLinkRepo{
		links: []*model.FriendlyLink{
			{ID: 1, Name: "待审核", URL: "https://example.com", Status: model.LinkStatusPending},
		},
	}
	svc := NewFriendlyLinkService(repo)

	if err := svc.ApproveLink(1, 1); err != nil {
		t.Fatalf("审核通过失败: %v", err)
	}

	if repo.links[0].Status != model.LinkStatusActive {
		t.Errorf("审核后状态 = %s, 期望 active", repo.links[0].Status)
	}
}

// TestApplyLinkCreatesPendingRecord 验证访客申请创建待审核记录，且响应不含联系邮箱。
func TestApplyLinkCreatesPendingRecord(t *testing.T) {
	repo := &fakeLinkRepo{}
	svc := NewFriendlyLinkService(repo)

	link, err := svc.ApplyLink(&ApplyFriendlyLinkRequest{
		Name:         "朋友的博客",
		URL:          "https://friend.example.com",
		Description:  "一个朋友的站点",
		ContactEmail: "friend@example.com",
	})
	if err != nil {
		t.Fatalf("提交友链申请失败: %v", err)
	}
	if len(repo.links) != 1 {
		t.Fatalf("仓储记录数 = %d, 期望 1", len(repo.links))
	}
	// 待审核状态以落库记录为准，公开视图不再输出状态字段。
	if repo.links[0].Status != model.LinkStatusPending {
		t.Errorf("申请状态 = %s, 期望 pending", repo.links[0].Status)
	}
	if repo.links[0].ContactEmail != "friend@example.com" {
		t.Errorf("联系邮箱 = %q, 期望落库保留 friend@example.com", repo.links[0].ContactEmail)
	}

	output := marshalLinkPublic(t, link)
	if strings.Contains(output, "contactEmail") || strings.Contains(output, "friend@example.com") {
		t.Errorf("友链申请响应不应回显联系邮箱，实际为 %s", output)
	}
}

// TestApplyLinkRejectsDuplicateURL 验证同一站点 URL 不论状态均拒绝重复申请。
func TestApplyLinkRejectsDuplicateURL(t *testing.T) {
	repo := &fakeLinkRepo{
		links: []*model.FriendlyLink{
			{ID: 1, Name: "已有站点", URL: "https://friend.example.com", Status: model.LinkStatusRejected},
		},
	}
	svc := NewFriendlyLinkService(repo)

	_, err := svc.ApplyLink(&ApplyFriendlyLinkRequest{
		Name:         "再次申请",
		URL:          "https://friend.example.com",
		ContactEmail: "friend@example.com",
	})
	if err == nil {
		t.Fatal("重复申请应返回错误")
	}
}

// TestListVisibleLinksHidesContactFields 验证公开友链列表不含联系邮箱与管理员备注。
func TestListVisibleLinksHidesContactFields(t *testing.T) {
	repo := &fakeLinkRepo{
		links: []*model.FriendlyLink{
			{
				ID:           1,
				Name:         "示例站点",
				URL:          "https://example.com",
				Description:  "示例简介",
				ContactEmail: "owner@example.com",
				Status:       model.LinkStatusActive,
				IsReciprocal: true,
				Note:         "内部备注：2026 年收录",
			},
			{ID: 2, Name: "待审核站点", URL: "https://pending.example.com", Status: model.LinkStatusPending},
		},
	}
	svc := NewFriendlyLinkService(repo)

	links, err := svc.ListVisibleLinks()
	if err != nil {
		t.Fatalf("查询公开友链列表失败: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("公开友链数 = %d, 期望仅展示 active 的 1 条", len(links))
	}

	data, err := json.Marshal(links)
	if err != nil {
		t.Fatalf("序列化公开友链列表失败: %v", err)
	}
	output := string(data)

	// 联系邮箱与管理员备注属个人信息与内部审计信息，公开列表一律不得输出。
	for _, forbidden := range []string{"contactEmail", "owner@example.com", "note", "内部备注"} {
		if strings.Contains(output, forbidden) {
			t.Errorf("公开友链列表不应包含 %s，实际为 %s", forbidden, output)
		}
	}
	if !strings.Contains(output, `"name":"示例站点"`) || !strings.Contains(output, `"isReciprocal":true`) {
		t.Errorf("公开友链列表应保留展示所需字段，实际为 %s", output)
	}

	// 公开视图输出键集须与白名单一致，防止个人信息字段被悄然加回。
	var decoded []map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("反序列化公开友链列表失败: %v", err)
	}
	if len(decoded) != 1 {
		t.Fatalf("公开友链条目数 = %d, 期望 1", len(decoded))
	}
	allowedKeys := map[string]bool{
		"id": true, "name": true, "url": true, "logo": true,
		"description": true, "sortOrder": true, "isReciprocal": true,
	}
	for key := range decoded[0] {
		if !allowedKeys[key] {
			t.Errorf("公开友链视图不应输出未知字段 %q，实际键集为 %v", key, linkViewKeys(decoded[0]))
		}
	}
}

// TestListLinksKeepsAdminAuditFields 验证管理端列表仍输出联系邮箱与备注供审核使用。
func TestListLinksKeepsAdminAuditFields(t *testing.T) {
	repo := &fakeLinkRepo{
		links: []*model.FriendlyLink{
			{ID: 1, Name: "示例站点", URL: "https://example.com", ContactEmail: "owner@example.com", Note: "内部备注"},
		},
	}
	svc := NewFriendlyLinkService(repo)

	result, err := svc.ListLinks(&ListFriendlyLinksRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("查询管理端友链列表失败: %v", err)
	}
	if len(result.Links) != 1 || result.Links[0].ContactEmail != "owner@example.com" {
		t.Fatal("管理端列表应保留联系邮箱供审核使用")
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("序列化管理端友链列表失败: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, `"contactEmail":"owner@example.com"`) || !strings.Contains(output, `"note":"内部备注"`) {
		t.Errorf("管理端列表应输出联系邮箱与备注，实际为 %s", output)
	}
}

// marshalLinkPublic 序列化公开友链视图，仅用于断言白名单字段。
func marshalLinkPublic(t *testing.T, link *FriendlyLinkPublic) string {
	t.Helper()
	data, err := json.Marshal(link)
	if err != nil {
		t.Fatalf("序列化公开友链视图失败: %v", err)
	}
	return string(data)
}

// linkViewKeys 列出公开视图解码后的键列表，仅用于测试失败信息展示。
func linkViewKeys(source map[string]any) []string {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	return keys
}
