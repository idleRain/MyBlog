package service

import (
	"testing"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeTagRepo 标签仓储的测试替身，记录调用并返回可配置结果。
type fakeTagRepo struct {
	repository.TagRepositoryInterface
	tags    []*model.Tag
	getByID func(id uint) (*model.Tag, error)
}

func (f *fakeTagRepo) GetByID(id uint) (*model.Tag, error) {
	if f.getByID != nil {
		return f.getByID(id)
	}
	for _, tag := range f.tags {
		if tag.ID == id {
			return tag, nil
		}
	}
	return nil, repository.ErrTagNotFound
}

func (f *fakeTagRepo) GetByName(name string) (*model.Tag, error) {
	for _, tag := range f.tags {
		if tag.Name == name {
			return tag, nil
		}
	}
	return nil, repository.ErrTagNotFound
}

func (f *fakeTagRepo) Create(tag *model.Tag) error {
	tag.ID = uint(len(f.tags) + 1)
	f.tags = append(f.tags, tag)
	return nil
}

func (f *fakeTagRepo) Update(tag *model.Tag) error {
	for i, existing := range f.tags {
		if existing.ID == tag.ID {
			f.tags[i] = tag
			return nil
		}
	}
	return repository.ErrTagNotFound
}

// TestCreateTagUniqueName 验证标签名称重复时返回业务错误。
func TestCreateTagUniqueName(t *testing.T) {
	repo := &fakeTagRepo{
		tags: []*model.Tag{{ID: 1, Name: "Go语言"}},
	}
	svc := NewTagService(repo)

	req := &CreateTagRequest{Name: "Go语言"}
	_, err := svc.CreateTag(req, 1)
	if err == nil {
		t.Fatal("重复标签名称应返回错误")
	}
}

// TestCreateTagDefaults 验证创建标签时默认状态启用、颜色为默认值。
func TestCreateTagDefaults(t *testing.T) {
	repo := &fakeTagRepo{}
	svc := NewTagService(repo)

	req := &CreateTagRequest{Name: "Gin框架"}
	tag, err := svc.CreateTag(req, 1)
	if err != nil {
		t.Fatalf("创建标签失败: %v", err)
	}

	if tag.Status != model.TagStatusEnabled {
		t.Errorf("默认状态 = %d, 期望启用 1", tag.Status)
	}
	if tag.Color != "#808080" {
		t.Errorf("默认颜色 = %q, 期望 #808080", tag.Color)
	}
}

// TestUpdateTagOptionalFields 验证更新标签时省略字段保留原值。
func TestUpdateTagOptionalFields(t *testing.T) {
	repo := &fakeTagRepo{
		tags: []*model.Tag{{ID: 1, Name: "Go语言", Description: "原始描述"}},
	}
	svc := NewTagService(repo)

	name := "Go语言新版"
	req := &UpdateTagRequest{
		ID:   1,
		Name: &name,
	}
	tag, err := svc.UpdateTag(req, 1)
	if err != nil {
		t.Fatalf("更新标签失败: %v", err)
	}

	if tag.Name != "Go语言新版" {
		t.Errorf("更新后名称 = %q, 期望 Go语言新版", tag.Name)
	}
	if tag.Description != "原始描述" {
		t.Errorf("省略字段描述被清空 = %q, 期望保留原始描述", tag.Description)
	}
}
