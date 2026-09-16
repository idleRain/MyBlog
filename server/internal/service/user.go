// Package service 业务逻辑层
package service

import (
	"MyBlog/internal/domain"
	"MyBlog/internal/model"
	"MyBlog/internal/repository"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// 密码加密相关常量
const (
	// bcrypt 加密成本，推荐值为 12
	BcryptCost = 12
)

// LoginResponse 登录响应
type LoginResponse struct {
	User         *domain.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	Permissions  []string     `json:"permissions"` // 当前角色权限列表，由后端下发的唯一权威
}

// UserService 用户服务接口
type UserService interface {
	CreateUser(req *domain.CreateUserRequest) (*domain.User, error)
	UpdateUser(req *domain.UpdateUserRequest) (*domain.User, error)
	GetUserByID(id uint) (*domain.User, error)
	GetUserList(page, pageSize int, keyword string) ([]*domain.User, int64, error)
	DeleteUser(id uint) error
	Login(username, password string) (*LoginResponse, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
	Logout(accessToken, refreshToken string) error
	// 自助资料
	GetProfile(userID uint) (*domain.User, error)
	UpdateProfile(userID uint, req *UpdateProfileRequest) (*domain.User, error)
	ChangePassword(userID uint, req *ChangePasswordRequest) error
	// 权限相关方法
	CanUserManageRole(managerRole, targetRole string) bool
	ValidateRoleTransition(currentRole, newRole string) error
}

// UpdateProfileRequest 自助资料更新请求，字段显式传入才更新。
type UpdateProfileRequest struct {
	Nickname *string `json:"nickname" binding:"omitempty,max=50"`
	Avatar   *string `json:"avatar" binding:"omitempty,max=255"`
	Bio      *string `json:"bio" binding:"omitempty,max=500"`
	Website  *string `json:"website" binding:"omitempty,max=255"`
}

// ChangePasswordRequest 修改密码请求。
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8,max=64"`
}

// 登录锁定策略默认值，组合根未注入策略或配置缺省时生效。
const (
	// DefaultMaxFailedLogins 连续登录失败阈值，达到即触发锁定。
	DefaultMaxFailedLogins = 5
	// DefaultLockDuration 账户锁定时长，到期后自动解除。
	DefaultLockDuration = 15 * time.Minute
)

// LoginLockoutPolicy 登录锁定策略，阈值与时长来源于 security.login_lockout 配置。
type LoginLockoutPolicy struct {
	Enabled         bool          // 是否启用失败锁定
	MaxFailedLogins uint          // 连续失败阈值
	LockDuration    time.Duration // 锁定时长
}

// defaultLoginLockoutPolicy 默认锁定策略，保证未配置时的生产安全底线。
func defaultLoginLockoutPolicy() LoginLockoutPolicy {
	return LoginLockoutPolicy{
		Enabled:         true,
		MaxFailedLogins: DefaultMaxFailedLogins,
		LockDuration:    DefaultLockDuration,
	}
}

// UserServiceOption 用户服务构造选项，可选依赖经选项注入而不破坏既有构造点。
type UserServiceOption func(*userService)

// WithLoginLockoutPolicy 注入登录锁定策略，覆盖默认策略。
func WithLoginLockoutPolicy(policy LoginLockoutPolicy) UserServiceOption {
	return func(s *userService) {
		s.lockoutPolicy = policy
	}
}

// userService 用户服务实现
type userService struct {
	userRepo      repository.UserRepository
	jwtService    JWTService
	rbacService   RBACService
	lockoutPolicy LoginLockoutPolicy
}

// NewUserService 创建用户服务实例，依赖由组合根注入，禁止内部私自实例化。
func NewUserService(userRepo repository.UserRepository, jwtService JWTService, rbacService RBACService,
	opts ...UserServiceOption) UserService {
	svc := &userService{
		userRepo:      userRepo,
		jwtService:    jwtService,
		rbacService:   rbacService,
		lockoutPolicy: defaultLoginLockoutPolicy(),
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// CreateUser 创建用户
func (s *userService) CreateUser(req *domain.CreateUserRequest) (*domain.User, error) {
	// 检查用户名是否已存在
	if existUser, _ := s.userRepo.GetByUsername(req.Username); existUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查邮箱是否已存在
	if existUser, _ := s.userRepo.GetByEmail(req.Email); existUser != nil {
		return nil, fmt.Errorf("邮箱已存在")
	}

	// 验证密码强度
	if err := s.validatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// 使用bcrypt加密密码
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 构建用户对象
	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Nickname: req.Nickname,
		Birthday: req.Birthday,
		Role:     req.Role,
		Status:   1,
	}

	// 如果昵称为空，使用用户名
	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	// 如果角色为空，设置默认角色为 user
	if user.Role == "" {
		user.Role = "user"
	}

	// 验证角色是否有效
	if !s.rbacService.IsValidRole(user.Role) {
		return nil, fmt.Errorf("无效的用户角色: %s", user.Role)
	}

	// 创建用户
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(req *domain.UpdateUserRequest) (*domain.User, error) {
	// 获取现有用户信息
	existingUser, err := s.userRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 检查用户名是否被其他用户占用
	if userByName, _ := s.userRepo.GetByUsername(req.Username); userByName != nil && userByName.ID != req.ID {
		return nil, fmt.Errorf("用户名已被其他用户使用")
	}

	// 检查邮箱是否被其他用户占用
	if userByEmail, _ := s.userRepo.GetByEmail(req.Email); userByEmail != nil && userByEmail.ID != req.ID {
		return nil, fmt.Errorf("邮箱已被其他用户使用")
	}

	// 更新基本信息
	existingUser.Username = req.Username
	existingUser.Email = req.Email
	existingUser.Nickname = req.Nickname
	existingUser.Birthday = req.Birthday
	existingUser.Role = req.Role

	// 仅当提供了状态字段时更新用户状态。
	if req.Status == 0 || req.Status == 1 {
		existingUser.Status = req.Status
	}

	// 如果昵称为空，使用用户名
	if existingUser.Nickname == "" {
		existingUser.Nickname = existingUser.Username
	}

	// 如果角色为空，设置默认角色为 user
	if existingUser.Role == "" {
		existingUser.Role = "user"
	}

	// 验证角色是否有效
	if !s.rbacService.IsValidRole(existingUser.Role) {
		return nil, fmt.Errorf("无效的用户角色: %s", existingUser.Role)
	}

	// 如果提供了新密码，则更新密码
	if req.Password != "" {
		// 验证密码强度
		if err := s.validatePasswordStrength(req.Password); err != nil {
			return nil, err
		}

		// 使用bcrypt加密密码
		hashedPassword, err := s.hashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败: %w", err)
		}

		existingUser.Password = hashedPassword
	}

	// 更新用户
	if err := s.userRepo.Update(existingUser); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return existingUser, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*domain.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserList 获取用户列表，keyword 非空时按用户名、邮箱或昵称模糊匹配。
func (s *userService) GetUserList(page, pageSize int, keyword string) ([]*domain.User, int64, error) {
	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, total, err := s.userRepo.List(offset, pageSize, keyword)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 删除用户
	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// GetProfile 获取当前登录用户的资料。
func (s *userService) GetProfile(userID uint) (*domain.User, error) {
	return s.userRepo.GetByID(userID)
}

// UpdateProfile 更新当前登录用户的自助资料字段，显式传入才更新。
func (s *userService) UpdateProfile(userID uint, req *UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.Website != nil {
		user.Website = *req.Website
	}

	// 如果昵称为空，使用用户名
	if user.Nickname == "" {
		user.Nickname = user.Username
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("更新资料失败: %w", err)
	}
	return user, nil
}

// ChangePassword 校验旧密码后更新为新密码，成功后撤销该用户全部既有令牌，
// 持有旧令牌的会话随之失效，当前请求使用的令牌同样被撤销，客户端需重新登录。
func (s *userService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if !s.verifyPassword(req.OldPassword, user.Password) {
		return fmt.Errorf("旧密码不正确")
	}

	if err := s.validatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	hashedPassword, err := s.hashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	// 撤销动作在密码已更新后执行，失败不回滚改密结果，内存实现当前无失败路径。
	_ = s.jwtService.RevokeUserTokens(userID)
	return nil
}

// hashPassword 使用bcrypt加密密码
func (s *userService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}
	return string(hashedBytes), nil
}

// verifyPassword 验证密码
func (s *userService) verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// validatePasswordStrength 验证密码强度
func (s *userService) validatePasswordStrength(password string) error {
	// 最小长度检查
	if len(password) < 6 {
		return fmt.Errorf("密码长度不能少于6位")
	}

	// 最大长度检查
	if len(password) > 100 {
		return fmt.Errorf("密码长度不能超过100位")
	}

	// 检查是否包含数字
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// 检查是否包含字母
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)

	// 要求至少包含数字和字母
	if !hasNumber || !hasLetter {
		return fmt.Errorf("密码必须包含字母和数字")
	}

	// 检查常见弱密码
	weakPasswords := []string{
		"123456", "password", "123456789", "12345678", "12345",
		"1234567", "1234567890", "qwerty", "abc123", "admin",
		"root", "guest", "test", "user", "demo", "login",
	}

	for _, weak := range weakPasswords {
		if password == weak {
			return fmt.Errorf("密码过于简单，请使用更复杂的密码")
		}
	}

	return nil
}

// Login 用户登录
func (s *userService) Login(username, password string) (*LoginResponse, error) {
	// 先尝试通过用户名查找
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		// 如果用户名未找到，尝试通过邮箱查找
		user, err = s.userRepo.GetByEmail(username)
		if err != nil {
			return nil, fmt.Errorf("用户不存在")
		}
	}

	// 锁定检查先于密码校验，锁定期间即使密码正确也拒绝，阻断持续爆破。
	if user.IsLocked() {
		return nil, fmt.Errorf("账户已被锁定，请稍后再试")
	}

	// 验证密码
	if !s.verifyPassword(password, user.Password) {
		return nil, s.recordLoginFailure(user)
	}

	// 检查用户状态
	if user.Status != model.UserStatusActive {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 登录成功后清零失败计数并解除历史锁定标记。
	s.resetLoginFailure(user)

	// 生成JWT令牌对
	tokenPair, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		Permissions:  permissionStrings(s.rbacService.GetUserPermissions(user.Role)),
	}, nil
}

// recordLoginFailure 累计连续登录失败次数，达到阈值时锁定账户一段时间。
// 计数落库失败不掩盖原始的密码错误结果，仅代表本次计数未持久化。
func (s *userService) recordLoginFailure(user *domain.User) error {
	user.FailedLoginCount++
	locked := false
	if s.lockoutPolicy.Enabled && user.FailedLoginCount >= s.lockoutPolicy.MaxFailedLogins {
		lockedUntil := time.Now().Add(s.lockoutPolicy.LockDuration)
		user.LockedUntil = &lockedUntil
		locked = true
	}

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("密码错误")
	}

	if locked {
		return fmt.Errorf("密码错误，失败次数过多，账户已被锁定，请稍后再试")
	}
	return fmt.Errorf("密码错误")
}

// resetLoginFailure 登录成功后清零失败计数并清除锁定截止时间。
// 计数本就为零时跳过写库，写库失败不影响本次登录结果。
func (s *userService) resetLoginFailure(user *domain.User) {
	if user.FailedLoginCount == 0 && user.LockedUntil == nil {
		return
	}

	user.FailedLoginCount = 0
	user.LockedUntil = nil
	_ = s.userRepo.Update(user)
}

// RefreshToken 刷新令牌，换取新令牌对前必须查库校验用户存在且状态正常，
// 被禁用或已删除用户的 refresh 不得继续换取新令牌。
// 刷新为低频路径，查库成本可接受；access 链路信任短有效期不逐请求查库，
// 需要访问令牌实时失效时由 R4 会话方案承接。
func (s *userService) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	if user.Status != model.UserStatusActive {
		return nil, fmt.Errorf("用户已被禁用")
	}

	return s.jwtService.RefreshAccessToken(refreshToken)
}

// Logout 用户登出，撤销访问令牌与刷新令牌构成的对。
// 刷新令牌由客户端在登出请求体中提交，缺省时仅撤销访问令牌。
func (s *userService) Logout(accessToken, refreshToken string) error {
	if accessToken != "" {
		if err := s.jwtService.RevokeToken(accessToken); err != nil {
			return err
		}
	}
	if refreshToken == "" {
		return nil
	}
	return s.jwtService.RevokeToken(refreshToken)
}

// CanUserManageRole 检查用户是否可以管理指定角色
func (s *userService) CanUserManageRole(managerRole, targetRole string) bool {
	return s.rbacService.CanManageUser(managerRole, targetRole)
}

// ValidateRoleTransition 验证角色转换是否有效
func (s *userService) ValidateRoleTransition(currentRole, newRole string) error {
	// 检查新角色是否有效
	if !s.rbacService.IsValidRole(newRole) {
		return fmt.Errorf("无效的目标角色: %s", newRole)
	}

	// 检查当前角色是否有效
	if !s.rbacService.IsValidRole(currentRole) {
		return fmt.Errorf("无效的当前角色: %s", currentRole)
	}

	// 超级管理员角色只能由系统设置，不能通过API变更
	if newRole == string(RoleSuperAdmin) {
		return fmt.Errorf("超级管理员角色只能通过系统管理员设置")
	}

	// 从超级管理员降级需要特殊处理
	if currentRole == string(RoleSuperAdmin) {
		return fmt.Errorf("超级管理员角色不能被降级")
	}

	return nil
}
