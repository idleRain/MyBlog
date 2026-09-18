# Service 模块

业务逻辑层（Service Layer），负责处理业务规则和逻辑。

## 设计原则

- 业务封装：将复杂的业务逻辑封装在Service层
- 数据校验：对输入数据进行业务级别的验证
- 事务管理：处理跨多个Repository的事务操作
- 接口抽象：通过接口定义服务规范

## 用户服务 (UserService)

### 接口定义

```go
type UserService interface {
    CreateUser(req *domain.CreateUserRequest) (*domain.User, error)
    UpdateUser(req *domain.UpdateUserRequest) (*domain.User, error)
    GetUserByID(id uint) (*domain.User, error)
    GetUserList(page, pageSize int, keyword string) ([]*domain.User, int64, error)
    DeleteUser(id uint) error
    Login(username, password string) (*LoginResponse, error)
    RefreshToken(refreshToken string) (*TokenPair, error)
    Logout(accessToken, refreshToken string) error
    GetProfile(userID uint) (*domain.User, error)
    UpdateProfile(userID uint, req *UpdateProfileRequest) (*domain.User, error)
    ChangePassword(userID uint, req *ChangePasswordRequest) error
    CanUserManageRole(managerRole, targetRole string) bool
    ValidateRoleTransition(currentRole, newRole string) error
}
```

### 构造与依赖

依赖经构造函数注入，组合根可经选项注入登录锁定策略：

```go
// 组合根内构造（第三参起为可选选项，缺省时启用默认锁定策略）
userService := service.NewUserService(userRepo, jwtService, rbacService,
    service.WithLoginLockoutPolicy(service.LoginLockoutPolicy{
        Enabled:         true,
        MaxFailedLogins: 5,
        LockDuration:    15 * time.Minute,
    }))
```

### 核心功能

#### 创建用户

- 验证用户名和邮箱的唯一性
- 密码加密处理
- 设置默认昵称与默认角色
- 数据持久化

#### 用户查询

- 根据ID查询用户信息
- 分页查询用户列表，keyword 非空时按用户名、邮箱或昵称模糊匹配
- 参数验证和默认值处理

#### 用户删除

- 验证用户存在性
- 执行软删除操作

### 认证语义

- **登录**：用户名未命中时回退邮箱查找；连续密码失败达到 `security.login_lockout` 配置阈值后置 `locked_until` 锁定账户，到期自动解除，登录成功清零计数
- **刷新**：换取新令牌对前查库校验用户存在且状态正常，被禁用或已删除用户拒绝刷新
- **登出**：撤销访问与刷新令牌构成的对，刷新令牌经请求体可选提交
- **改密**：成功后撤销该用户全部既有令牌，当前会话一并失效

### 业务规则

1. **用户名唯一性**: 不允许重复的用户名
2. **邮箱唯一性**: 不允许重复的邮箱地址
3. **密码安全**: 使用bcrypt加密密码
4. **默认昵称**: 如果不提供昵称，使用用户名作为默认昵称
5. **分页限制**: 每页最多100条记录，默认10条

### 使用示例

```go
// 创建用户
req := &domain.CreateUserRequest{
    Username: "john_doe",
    Email:    "john@example.com",
    Password: "Passw0rd123",
    Nickname: "John",
}
user, err := userService.CreateUser(req)

// 登录
login, err := userService.Login("john_doe", "Passw0rd123")

// 获取用户列表（keyword 支持用户名、邮箱、昵称模糊匹配）
users, total, err := userService.GetUserList(1, 10, "john")

// 修改密码（成功后撤销该用户全部既有令牌）
err = userService.ChangePassword(1, &ChangePasswordRequest{
    OldPassword: "Passw0rd123",
    NewPassword: "N3wPassw0rd456",
})
```

### 密码加密

#### 加密方案（bcrypt）
系统使用安全的bcrypt加密：

```go
func (s *userService) hashPassword(password string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
    if err != nil {
        return "", fmt.Errorf("密码加密失败: %w", err)
    }
    return string(hashedBytes), nil
}
```

#### 密码强度验证
系统会验证密码强度，要求：
- 最少6位，最多100位
- 必须包含字母和数字
- 不能使用常见弱密码

#### 密码验证
系统使用bcrypt进行密码验证：

```go
func (s *userService) verifyPassword(password, hashedPassword string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
```

**安全特性**：
- ✅ 使用bcrypt加密（安全）
- ✅ 密码强度验证
- ✅ 防止常见弱密码

### 错误处理

Service层会返回具体的业务错误信息：

- "用户名已存在"
- "邮箱已存在"
- "用户不存在"
- "创建用户失败"
- "删除用户失败"

## 扩展指南

### 添加新的业务方法

1. 在接口中定义新方法
2. 在实现中添加业务逻辑
3. 编写单元测试
4. 更新文档

### 添加新的Service

1. 定义Service接口
2. 实现Service接口
3. 注入依赖的Repository
4. 在main.go中注册服务
5. 编写测试和文档

### 最佳实践

1. **单一职责**: 每个Service只处理一个业务域
2. **依赖注入**: 通过构造函数注入Repository依赖
3. **错误包装**: 将Repository错误包装为业务错误
4. **参数验证**: 对所有输入参数进行验证
5. **事务处理**: 复杂操作使用数据库事务
