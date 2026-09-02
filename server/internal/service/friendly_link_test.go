package service

import (
	"testing"

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

func (f *fakeLinkRepo) GetByID(id uint) (*model.FriendlyLink, error) {
	if f.getByID != nil {
		return f.getByID(id)
	}
	for _, link := range f.links {
		if link.ID == id {
			return link, nil
		}
	}
	return nil, repository.ErrFriendlyLinkNotFound
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
	return nil, repository.ErrFriendlyLinkNotFound
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
	return repository.ErrFriendlyLinkNotFound
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
