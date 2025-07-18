# Database 模块

数据库连接和管理模块，基于GORM提供MySQL数据库操作功能。

## 功能特性

- 单例模式确保数据库连接全局唯一
- 自动创建数据库（如果不存在）
- 连接池配置和管理
- 数据库健康检查
- 自动表结构迁移
- 优雅关闭连接

## 核心功能

### 数据库连接管理

- `InitMySQL(cfg *config.Config)`: 初始化MySQL连接
- `GetDB()`: 获取数据库实例
- `Close()`: 关闭数据库连接
- `HealthCheck()`: 数据库健康检查

### 数据库维护

- `AutoMigrate(models ...interface{})`: 自动迁移表结构
- `createDatabaseIfNotExists()`: 自动创建数据库

## 使用示例

### 初始化数据库连接

```go
package main

import (
  "MyBlog/internal/config"
  "MyBlog/internal/database"
  "log"
)

func main() {
  // 加载配置
  cfg, err := config.Load("configs/config.yaml")
  if err != nil {
    log.Fatal("配置加载失败:", err)
  }

  // 初始化数据库
  db, err := database.InitMySQL(cfg)
  if err != nil {
    log.Fatal("数据库初始化失败:", err)
  }

  // 在其他地方使用数据库
  db = database.GetDB()
}
```

### 数据库表迁移

```go
// 定义模型
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Username string `gorm:"uniqueIndex;not null"`
    Email    string `gorm:"uniqueIndex;not null"`
    Password string `gorm:"not null"`
}

// 执行迁移
err := database.AutoMigrate(&User{})
if err != nil {
    log.Fatal("表结构迁移失败:", err)
}
```

### 健康检查

```go
if err := database.HealthCheck(); err != nil {
    log.Printf("数据库健康检查失败: %v", err)
}
```

### 优雅关闭

```go
defer func() {
    if err := database.Close(); err != nil {
        log.Printf("关闭数据库连接失败: %v", err)
    }
}()
```

## 配置说明

数据库配置通过 `configs/config.yaml` 文件进行管理：

```yaml
database:
  host: "localhost"          # 数据库主机
  port: 3306                # 数据库端口
  username: "root"          # 用户名
  password: "123456"        # 密码
  dbname: "blog"           # 数据库名
  charset: "utf8mb4"       # 字符集
  parse_time: true         # 解析时间类型
  loc: "Local"            # 时区
  max_idle_conns: 10      # 最大空闲连接数
  max_open_conns: 100     # 最大打开连接数
```

## 注意事项

1. 必须先初始化配置，再初始化数据库
2. 数据库连接采用单例模式，整个应用生命周期内只初始化一次
3. 程序退出时建议调用 `Close()` 方法优雅关闭连接
4. 如果数据库不存在，模块会自动创建
5. 连接池参数可以根据实际需求调整

## 错误处理

模块提供详细的错误信息，包括：

- 连接失败原因
- 配置参数错误
- 数据库操作失败
- 健康检查失败

所有错误都会包含上下文信息，便于问题排查。
