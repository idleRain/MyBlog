# Repository 模块

数据访问层（Repository Layer），负责与数据库交互的底层操作。

## 设计原则

- 单一职责：每个Repository只负责一个实体的数据操作
- 接口抽象：通过接口定义操作规范，便于测试和扩展
- 错误处理：提供详细的错误信息和上下文

## 用户仓库 (UserRepository)

### 模型定义

```go
type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Username  string         `json:"username" gorm:"uniqueIndex;not null;size:50"`
    Email     string         `json:"email" gorm:"uniqueIndex;not null;size:100"`
    Password  string         `json:"-" gorm:"not null;size:255"`
    Nickname  string         `json:"nickname" gorm:"size:50"`
    Avatar    string         `json:"avatar" gorm:"size:255"`
    Status    int            `json:"status" gorm:"default:1;comment:状态 1-正常 0-禁用"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
```

### 接口方法

- `Create(user *User)`: 创建用户
- `GetByID(id uint)`: 根据ID获取用户
- `GetByUsername(username string)`: 根据用户名获取用户
- `GetByEmail(email string)`: 根据邮箱获取用户
- `Update(user *User)`: 更新用户
- `Delete(id uint)`: 删除用户（软删除）
- `List(offset, limit int)`: 获取用户列表

### 请求结构

```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=1,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,max=100"`
    Nickname string `json:"nickname" binding:"max=50"`
}
```

### 使用示例

```go
// 初始化仓库
userRepo := repository.NewUserRepository(db)

// 创建用户
user := &repository.User{
    Username: "john_doe",
    Email:    "john@example.com",
    Password: "hashed_password",
    Nickname: "John",
    Status:   1,
}
err := userRepo.Create(user)

// 查询用户
user, err := userRepo.GetByID(1)
user, err := userRepo.GetByUsername("john_doe")
user, err := userRepo.GetByEmail("john@example.com")

// 分页查询
users, total, err := userRepo.List(0, 10)

// 更新用户
user.Nickname = "John Doe"
err := userRepo.Update(user)

// 删除用户
err := userRepo.Delete(1)
```

## 注意事项

1. **数据验证**: Repository层只进行基本的数据库约束验证
2. **错误处理**: 所有数据库错误都会被包装并返回详细信息
3. **软删除**: 删除操作使用GORM的软删除功能
4. **唯一索引**: 用户名和邮箱都有唯一索引约束
5. **密码安全**: 密码字段在JSON序列化时会被忽略

## 扩展指南

添加新的Repository时，请遵循以下步骤：

1. 定义模型结构体
2. 定义Repository接口
3. 实现Repository接口
4. 编写单元测试
5. 更新文档