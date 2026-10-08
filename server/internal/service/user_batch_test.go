package service

import (
	"errors"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/repository"
)

// fakeBatchUserRepo 批量操作的用户仓储替身，记录批量方法的触达与传入集合。
type fakeBatchUserRepo struct {
	repository.UserRepository
	users        []*domain.User
	deleteBatch  func(ids []uint) error
	updateStatus func(ids []uint, status int) error
	deleteCalls  int
	statusCalls  int
	lastStatus   int
}

func (f *fakeBatchUserRepo) ListByIDs(ids []uint) ([]*domain.User, error) {
	var matched []*domain.User
	for _, id := range ids {
		for _, user := range f.users {
			if user.ID == id {
				matched = append(matched, user)
			}
		}
	}
	return matched, nil
}

func (f *fakeBatchUserRepo) DeleteBatch(ids []uint) error {
	f.deleteCalls++
	if f.deleteBatch != nil {
		return f.deleteBatch(ids)
	}
	return nil
}

func (f *fakeBatchUserRepo) UpdateStatusBatch(ids []uint, status int) error {
	f.statusCalls++
	f.lastStatus = status
	if f.updateStatus != nil {
		return f.updateStatus(ids, status)
	}
	return nil
}

// batchServiceUser 构造指定角色的用户测试数据。
func batchServiceUser(id uint, role string) *domain.User {
	return &domain.User{ID: id, Username: "user", Role: role, Status: domain.UserStatusActive}
}

// TestBatchDeleteUsersRejectsIncludingSelf 包含操作者本人的批量删除整批拒绝。
func TestBatchDeleteUsersRejectsIncludingSelf(t *testing.T) {
	repo := &fakeBatchUserRepo{users: []*domain.User{batchServiceUser(2, "user")}}
	svc := &userService{userRepo: repo, rbacService: NewRBACService()}

	if err := svc.BatchDeleteUsers([]uint{2, 7}, 7, "admin"); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("包含自己的批量删除应被拒绝，实际 %v", err)
	}
	if repo.deleteCalls != 0 {
		t.Error("校验失败不应触达删除")
	}
}

// TestBatchDeleteUsersRejectsUnmanagedRole 含不可管理角色目标的批量删除整批拒绝。
func TestBatchDeleteUsersRejectsUnmanagedRole(t *testing.T) {
	repo := &fakeBatchUserRepo{
		users: []*domain.User{
			batchServiceUser(2, "user"),
			batchServiceUser(3, "superadmin"),
		},
	}
	svc := &userService{userRepo: repo, rbacService: NewRBACService()}

	err := svc.BatchDeleteUsers([]uint{2, 3}, 7, "admin")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("含更高角色的批量删除应被拒绝，实际 %v", err)
	}
	if repo.deleteCalls != 0 {
		t.Error("校验失败不应触达删除")
	}
}

// TestBatchDeleteUsersRejectsUnknownUser 目标不存在的批量删除整批拒绝。
func TestBatchDeleteUsersRejectsUnknownUser(t *testing.T) {
	repo := &fakeBatchUserRepo{users: []*domain.User{batchServiceUser(2, "user")}}
	svc := &userService{userRepo: repo, rbacService: NewRBACService()}

	err := svc.BatchDeleteUsers([]uint{2, 99}, 7, "admin")
	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("含不存在目标的批量删除应被拒绝，实际 %v", err)
	}
	if repo.deleteCalls != 0 {
		t.Error("校验失败不应触达删除")
	}
}

// TestBatchDeleteUsersRejectsOversized 数量超出上限的批量删除被拒绝。
func TestBatchDeleteUsersRejectsOversized(t *testing.T) {
	repo := &fakeBatchUserRepo{}
	svc := &userService{userRepo: repo, rbacService: NewRBACService()}

	ids := make([]uint, batchUserMaxLimit+1)
	for i := range ids {
		ids[i] = uint(i + 1)
	}
	if err := svc.BatchDeleteUsers(ids, 1, "admin"); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("超限批量删除应被拒绝，实际 %v", err)
	}
}

// TestBatchDeleteUsersDeletesAtomicallyOnValidTargets 校验通过后进入原子删除路径。
func TestBatchDeleteUsersDeletesAtomicallyOnValidTargets(t *testing.T) {
	repo := &fakeBatchUserRepo{users: []*domain.User{batchServiceUser(2, "user")}}
	svc := &userService{userRepo: repo, rbacService: NewRBACService()}

	// 删除失败模拟数据库层异常，校验通过后错误上抛。
	repo.deleteBatch = func([]uint) error {
		return errors.New("模拟数据库删除失败")
	}

	if err := svc.BatchDeleteUsers([]uint{2}, 1, "admin"); err == nil {
		t.Fatal("数据库删除失败应上抛错误")
	}
	if repo.deleteCalls != 1 {
		t.Errorf("合法目标的删除应触达一次，实际 %d", repo.deleteCalls)
	}
}

// TestBatchUpdateUserStatusValidatesRange 批量状态仅允许启用与禁用两态。
func TestBatchUpdateUserStatusValidatesRange(t *testing.T) {
	repo := &fakeBatchUserRepo{users: []*domain.User{batchServiceUser(2, "user")}}
	svc := &userService{userRepo: repo, rbacService: NewRBACService()}

	if err := svc.BatchUpdateUserStatus([]uint{2}, domain.UserStatusLocked, 1, "admin"); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("锁定状态不应经批量路径写入，实际 %v", err)
	}
	if repo.statusCalls != 0 {
		t.Error("校验失败不应触达状态更新")
	}

	if err := svc.BatchUpdateUserStatus([]uint{2}, domain.UserStatusActive, 1, "admin"); err != nil {
		t.Errorf("合法状态应成功，实际 %v", err)
	}
	if repo.statusCalls != 1 || repo.lastStatus != domain.UserStatusActive {
		t.Errorf("状态更新应触达一次且写入启用值，实际 %d / %d", repo.statusCalls, repo.lastStatus)
	}
}

// TestBatchDeleteUsersRejectsEmptyList 空集合的批量删除被拒绝。
func TestBatchDeleteUsersRejectsEmptyList(t *testing.T) {
	repo := &fakeBatchUserRepo{}
	svc := &userService{userRepo: repo, rbacService: NewRBACService()}

	if err := svc.BatchDeleteUsers(nil, 1, "admin"); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("空集合应被拒绝，实际 %v", err)
	}
}
