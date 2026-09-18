# Config 模块

配置管理模块，负责应用程序配置的加载、解析和验证。

## 功能特性

- 支持YAML格式配置文件
- 单例模式确保配置全局唯一
- 配置参数验证
- 默认值设置
- 类型安全的配置访问

## 配置结构

### 服务器配置 (ServerConfig)

- `host`: 服务器监听主机
- `port`: 服务器监听端口
- `mode`: 运行模式 (debug/release/test)

### 数据库配置 (DatabaseConfig)

- `host`: 数据库主机地址
- `port`: 数据库端口
- `username`: 数据库用户名
- `password`: 数据库密码
- `dbname`: 数据库名称
- `charset`: 字符集
- `parse_time`: 是否解析时间类型
- `loc`: 时区设置
- `max_idle_conns`: 最大空闲连接数
- `max_open_conns`: 最大打开连接数

### 日志配置 (LoggerConfig)

- `level`: 日志级别 (debug/info/warn/error)
- `output`: 输出方式 (stdout/file)
- `file_path`: 日志文件路径

### API配置 (APIConfig)

- `version`: API版本
- `timeout`: 请求超时时间（秒）

### JWT配置 (JWTConfig)

- `access_secret` / `refresh_secret`: 双令牌签名密钥，**无代码默认值**，缺失或使用已公开弱密钥时启动直接失败，生产经 `MYBLOG_JWT_ACCESS_SECRET` / `MYBLOG_JWT_REFRESH_SECRET` 环境变量注入，双密钥必须互异
- `access_expire`: 访问令牌有效期（分钟）
- `refresh_expire`: 刷新令牌有效期（小时）
- `issuer`: 签发者标识

### 安全配置 (SecurityConfig)

- `rate_limit`: IP 与用户两级频率限制
- `security_headers`: CSP、X-Frame-Options 等安全响应头
- `input_validation`: 请求体大小上限与 User-Agent 黑名单（子串匹配，禁止收录泛化词）
- `admin_security`: 管理员接口更严格的频率限制与 IP 白名单
- `login_lockout`: 登录失败锁定策略（`enabled`、`max_failed_attempts`、`lock_minutes`），连续失败达阈值锁定账户，到期自动解除

### CORS配置 (CORSConfig)

- `allowed_origins`: 允许跨域的 Origin 白名单，精确匹配，空列表拒绝所有跨域（nginx 同源网关部署形态保持为空即可）
- `allow_credentials`: 是否允许携带凭证，与 Origin 全放行互斥
- `allowed_methods`: 允许的 HTTP 方法，POST-Only 规范下为 POST 与预检 OPTIONS
- `allowed_headers`: 允许跨域携带的请求头

### 媒体配置 (MediaConfig)

- `upload_dir`: 本地存储目录
- `base_url`: 文件访问 URL 前缀
- `max_size_mb`: 单文件大小上限（MB）
- `allowed_types`: 允许的 MIME 类型

### RBAC配置 (RBACConfig)

- `role_hierarchy`: 角色层级映射，数值越大权限越高，四类角色必须全部登记
- `role_permissions`: 角色到权限列表的映射，生产环境权限唯一权威

## 使用示例

```go
package main

import (
  "MyBlog/internal/config"
  "log"
)

func main() {
  // 加载配置
  cfg, err := config.Load("configs/config.yaml")
  if err != nil {
    log.Fatal("配置加载失败:", err)
  }

  // 获取数据库连接串
  dsn := cfg.GetDSN()

  // 获取服务器地址
  addr := cfg.GetServerAddress()

  // 在其他地方获取全局配置
  globalCfg := config.Get()
}
```

## 配置文件示例

参见 `configs/config.yaml` 文件。

## 注意事项

1. 必须先调用 `Load()` 方法初始化配置
2. 配置采用单例模式，整个应用生命周期内只加载一次
3. 配置文件路径相对于项目根目录
4. 配置项均有代码默认值，**JWT 双密钥除外**：密钥不设默认值，缺失即启动失败（OPS-08 定案）
5. 标量配置项可经 `MYBLOG_<SECTION>_<FIELD>` 环境变量覆盖（如 `MYBLOG_JWT_ACCESS_SECRET`），列表与映射保持 YAML 原值
