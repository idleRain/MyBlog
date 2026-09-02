package service

import (
	"errors"
	"testing"

	"MyBlog/internal/model"
	"MyBlog/internal/repository"
)

// fakeCategoryRepo 分类仓储的测试替身，记录调用并返回可配置结果。
type fakeCategoryRepo struct {
	repository.CategoryRepositoryInterface
	categories   []*model.Category
	getByID      func(id uint) (*model.Category, error)
	countByParent func(parentID uint) (int64, error)
}

func (f *fakeCategoryRepo) GetByID(id uint) (*model.Category, error) {
	if f.getByID != nil {
		return f.getByID(id)
	}
	for _, category := range f.categories {
		if category.ID == id {
			return category, nil
		}
	}
	return nil, repository.ErrCategoryNotFound
}

func (f *fakeCategoryRepo) CountByParentID(parentID uint) (int64, error) {
	if f.countByParent != nil {
		return f.countByParent(parentID)
	}
	var count int64
	for _, category := range f.categories {
		if category.ParentID != nil && *category.ParentID == parentID {
			count++
		}
	}
	return count, nil
}

func (f *fakeCategoryRepo) Create(category *model.Category) error {
	category.ID = uint(len(f.categories) + 1)
	f.categories = append(f.categories, category)
	return nil
}

func (f *fakeCategoryRepo) UpdatePath(id uint, path string) error {
	for _, category := range f.categories {
		if category.ID == id {
			category.Path = path
		}
	}
	return nil
}

func (f *fakeCategoryRepo) ListAll() ([]*model.Category, error) {
	return f.categories, nil
}

func (f *fakeCategoryRepo) Delete(id uint) error {
	for i, category := range f.categories {
		if category.ID == id {
			f.categories = append(f.categories[:i], f.categories[i+1:]...)
			return nil
		}
	}
	return repository.ErrCategoryNotFound
}

// TestBuildCategoryTree 验证扁平分类列表正确构建为树形结构。
func TestBuildCategoryTree(t *testing.T) {
	parentID := uint(1)
	categories := []*model.Category{
		{ID: 1, Name: "技术分享"},
		{ID: 2, Name: "后端开发", ParentID: &parentID},
		{ID: 3, Name: "前端开发", ParentID: &parentID},
		{ID: 4, Name: "生活随笔"},
	}

	tree := buildCategoryTree(categories)

	if len(tree) != 2 {
		t.Fatalf("根节点数 = %d, 期望 2", len(tree))
	}

	// 根节点 1 应包含两个子节点。
	if len(tree[0].Children) != 2 {
		t.Errorf("根节点 1 子节点数 = %d, 期望 2", len(tree[0].Children))
	}
}

// TestCreateCategoryWithParent 验证带父分类时推导 root 与 level 字段。
func TestCreateCategoryWithParent(t *testing.T) {
	parentID := uint(1)
	parentRoot := uint(1)
	repo := &fakeCategoryRepo{
		categories: []*model.Category{
			{ID: 1, Name: "技术分享", RootID: &parentRoot, Level: 1, Path: "/1"},
		},
	}
	svc := NewCategoryService(repo)

	req := &CreateCategoryRequest{
		Name:     "后端开发",
		ParentID: &parentID,
	}
	category, err := svc.CreateCategory(req, 1)
	if err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}

	if category.RootID == nil || *category.RootID != 1 {
		t.Errorf("RootID = %v, 期望继承父分类的 1", category.RootID)
	}
	if category.Level != 2 {
		t.Errorf("Level = %d, 期望父分类层级 +1 为 2", category.Level)
	}
	if category.Path != "/1/2" {
		t.Errorf("Path = %q, 期望 /1/2", category.Path)
	}
}

// TestDeleteCategoryWithChildren 验证存在子分类时禁止删除。
func TestDeleteCategoryWithChildren(t *testing.T) {
	parentID := uint(1)
	repo := &fakeCategoryRepo{
		categories: []*model.Category{
			{ID: 1, Name: "技术分享"},
			{ID: 2, Name: "后端开发", ParentID: &parentID},
		},
	}
	svc := NewCategoryService(repo)

	err := svc.DeleteCategory(1, 1)
	if err == nil {
		t.Fatal("存在子分类时删除应返回错误")
	}
}

// TestDeleteCategoryNotFound 验证删除不存在的分类返回哨兵错误。
func TestDeleteCategoryNotFound(t *testing.T) {
	repo := &fakeCategoryRepo{
		getByID: func(id uint) (*model.Category, error) {
			return nil, repository.ErrCategoryNotFound
		},
	}
	svc := NewCategoryService(repo)

	err := svc.DeleteCategory(999, 1)
	if !errors.Is(err, repository.ErrCategoryNotFound) {
		t.Errorf("删除不存在分类应返回 ErrCategoryNotFound，实际为 %v", err)
	}
}
