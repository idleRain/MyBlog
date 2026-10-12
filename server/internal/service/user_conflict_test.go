package service

import (
	"errors"
	"testing"

	"MyBlog/internal/domain"
	"MyBlog/internal/repository"
)

// conflictUserRepo 用户冲突场景的仓储替身。
// 按预置索引精确返回既有用户，命中不到时返回 not-found 哨兵，
// 被本文件测试路径触达的方法全部显式覆写，不依赖内嵌空接口的静默兼容。
type conflictUserRepo struct {
	repository.UserRepository
	target     *domain.User
	byUsername map[string]*domain.User
	byEmail    map[string]*domain.User
	created    []*domain.User
	updated    []*domain.User
}

func (r *conflictUserRepo) GetByID(id uint) (*domain.User, error) {
	if r.target == nil || r.target.ID != id {
		return nil, domain.ErrUserNotFound
	}
	return r.target, nil
}

func (r *conflictUserRepo) GetByUsername(username string) (*domain.User, error) {
	if user, ok := r.byUsername[username]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (r *conflictUserRepo) GetByEmail(email string) (*domain.User, error) {
	if user, ok := r.byEmail[email]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (r *conflictUserRepo) Create(user *domain.User) error {
	r.created = append(r.created, user)
	return nil
}

func (r *conflictUserRepo) Update(user *domain.User) error {
	r.updated = append(r.updated, user)
	return nil
}

// TestCreateUserRejectsDuplicateEmail 验证创建用户时邮箱冲突返回 ErrEmailTaken。
// 该路径原先以裸错误表达，handler 分档退化为 500 且丢失可修正提示。
func TestCreateUserRejectsDuplicateEmail(t *testing.T) {
	existing := &domain.User{ID: 1, Username: "existing", Email: "taken@example.com"}
	repo := &conflictUserRepo{
		byUsername: map[string]*domain.User{existing.Username: existing},
		byEmail:    map[string]*domain.User{existing.Email: existing},
	}
	svc := NewUserService(repo, &recordedTokenService{}, NewRBACService())

	_, err := svc.CreateUser(&domain.CreateUserRequest{
		Username: "fresh",
		Email:    existing.Email,
		Password: "k3Vb9xQ2mN",
	})

	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("创建时邮箱冲突错误 = %v，期望包装 ErrEmailTaken", err)
	}
	if len(repo.created) != 0 {
		t.Error("邮箱冲突时不应写入用户记录")
	}
}

// TestUpdateUserRejectsDuplicateUsername 验证更新用户时用户名被他人占用返回 ErrUsernameTaken。
func TestUpdateUserRejectsDuplicateUsername(t *testing.T) {
	target := &domain.User{ID: 1, Username: "self", Email: "self@example.com", Role: "user", Status: 1}
	other := &domain.User{ID: 2, Username: "taken", Email: "other@example.com", Role: "user", Status: 1}
	repo := &conflictUserRepo{
		target:     target,
		byUsername: map[string]*domain.User{other.Username: other},
		byEmail:    map[string]*domain.User{target.Email: target},
	}
	svc := NewUserService(repo, &recordedTokenService{}, NewRBACService())

	_, err := svc.UpdateUser(&domain.UpdateUserRequest{
		ID:       target.ID,
		Username: other.Username,
		Email:    target.Email,
		Role:     target.Role,
	})

	if !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("更新时用户名冲突错误 = %v，期望包装 ErrUsernameTaken", err)
	}
	if len(repo.updated) != 0 {
		t.Error("用户名冲突时不应写回用户记录")
	}
}
