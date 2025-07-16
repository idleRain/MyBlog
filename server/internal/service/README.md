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
    CreateUser(req *repository.CreateUserRequest) (*repository.User, error)
    GetUserByID(id uint) (*repository.User, error)
    GetUserList(page, pageSize int) ([]*repository.User, int64, error)
    DeleteUser(id uint) error
}
```

### 核心功能

#### 创建用户
- 验证用户名和邮箱的唯一性
- 密码加密处理
- 设置默认昵称
- 数据持久化

#### 用户查询
- 根据ID查询用户信息
- 分页查询用户列表
- 参数验证和默认值处理

#### 用户删除
- 验证用户存在性
- 执行软删除操作

### 业务规则

1. **用户名唯一性**: 不允许重复的用户名
2. **邮箱唯一性**: 不允许重复的邮箱地址
3. **密码安全**: 使用MD5+盐值加密密码
4. **默认昵称**: 如果不提供昵称，使用用户名作为默认昵称
5. **分页限制**: 每页最多100条记录，默认10条

### 使用示例

```go
// 初始化服务
userService := service.NewUserService(userRepo)

// 创建用户
req := &repository.CreateUserRequest{
    Username: "john_doe",
    Email:    "john@example.com",
    Password: "123456",
    Nickname: "John",
}
user, err := userService.CreateUser(req)

// 获取用户
user, err := userService.GetUserByID(1)

// 获取用户列表
users, total, err := userService.GetUserList(1, 10)

// 删除用户
err := userService.DeleteUser(1)
```

### 密码加密

当前使用MD5+盐值的方式加密密码：

```go
func (s *userService) hashPassword(password string) string {
    h := md5.New()
    h.Write([]byte(password + "myblog_salt"))
    return hex.EncodeToString(h.Sum(nil))
}
```

> 注意：生产环境建议使用更安全的加密算法，如bcrypt

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