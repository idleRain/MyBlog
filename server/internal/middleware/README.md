# Middleware 中间件模块

> MyBlog 项目的 HTTP 中间件集合，提供统一的请求处理、安全防护、日志记录等功能。

## 📋 目录结构

```
middleware/
├── README.md          # 本文档
├── auth.go           # JWT 认证中间件
├── rbac.go           # RBAC权限控制中间件（新增）
├── cors.go           # 跨域资源共享中间件
├── logger.go         # 请求日志和请求ID中间件
├── ratelimit.go      # 频率限制中间件（已废弃，建议使用 security.go）
└── security.go       # 综合安全中间件（推荐）
```

## 🔧 中间件列表

### 1. 认证中间件 (`auth.go`)

**功能**：JWT 令牌验证和用户身份认证

```go
// 基础认证中间件
func Auth(jwtService service.JWTService) gin.HandlerFunc

// 可选认证中间件（不强制要求认证）
func OptionalAuth(jwtService service.JWTService) gin.HandlerFunc
```

**特性**：
- ✅ 支持 `Authorization: Bearer <token>` 头认证
- ✅ JWT 令牌解析和验证
- ✅ 用户信息注入到 Context
- ✅ 令牌过期和无效检测
- ✅ 可选认证模式支持

**Context 注入**：
- `userID` (uint): 用户ID
- `username` (string): 用户名
- `user` (*repository.User): 完整用户对象

**使用示例**：
```go
// 必须认证的路由
authenticatedGroup := api.Group("/users")
authenticatedGroup.Use(middleware.Auth(jwtService))

// 可选认证的路由
publicGroup := api.Group("/public")
publicGroup.Use(middleware.OptionalAuth(jwtService))
```

### 2. CORS 中间件 (`cors.go`)

**功能**：跨域资源共享配置

```go
func CORS() gin.HandlerFunc
```

**配置**：
- **允许的源**：`http://localhost:5173` (开发环境前端)
- **允许的方法**：GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD
- **允许的头**：Authorization, Content-Type, Accept, Origin, User-Agent, Cache-Control, Keep-Alive
- **暴露的头**：Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers
- **凭据支持**：启用 (`Access-Control-Allow-Credentials: true`)
- **预检缓存**：24小时

### 3. 日志中间件 (`logger.go`)

**功能**：请求日志记录和请求ID生成

```go
// 请求日志中间件
func Logger() gin.HandlerFunc

// 请求ID中间件
func RequestID() gin.HandlerFunc
```

**日志字段**：
- **基础信息**：请求ID、时间戳、HTTP方法、路径
- **响应信息**：状态码、响应时间、响应大小
- **客户端信息**：客户端IP、User-Agent
- **错误信息**：错误详情（如有）

**请求ID**：
- 生成格式：`req_` + 16字符随机字符串
- 注入到 Context：`request_id`
- 响应头：`X-Request-Id`

### 4. 安全中间件 (`security.go`) ⭐ **推荐**

**功能**：综合安全防护，包含频率限制、安全头设置、输入验证

```go
// 默认安全配置
func SecurityMiddleware(config *SecurityConfig) gin.HandlerFunc
func DefaultSecurityConfig() *SecurityConfig

// 从配置文件创建
func SecurityMiddlewareFromConfig(cfg *config.Config) gin.HandlerFunc

// 管理员专用安全中间件
func AdminSecurityMiddleware() gin.HandlerFunc
func AdminSecurityMiddlewareFromConfig(cfg *config.Config) gin.HandlerFunc

// IP白名单中间件
func IPWhitelistMiddleware(whitelist []string) gin.HandlerFunc
```

#### 安全功能详解

##### 🛡️ 频率限制

**IP级别限制**：
- 默认：100次/分钟
- 管理员接口：30次/分钟
- 使用滑动窗口算法
- 自动清理过期记录

**用户级别限制**：
- 默认：300次/分钟  
- 管理员接口：50次/分钟
- 基于认证用户ID
- 与IP限制独立计算

##### 🔒 安全HTTP头

```http
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self' https:
X-Frame-Options: SAMEORIGIN
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
X-XSS-Protection: 1; mode=block
X-DNS-Prefetch-Control: off
X-Download-Options: noopen
X-Permitted-Cross-Domain-Policies: none
Strict-Transport-Security: max-age=31536000; includeSubDomains (仅HTTPS)
```

##### 🔍 输入验证

**恶意模式检测**：
- XSS攻击：`<script>` 标签、`javascript:` URL、事件处理器
- SQL注入：`UNION SELECT`、`INSERT INTO`、`DROP TABLE`、`OR/AND` 模式
- 命令执行：`exec()`、`system()` 调用
- 路径遍历：`../`、`..\` 模式

**检测范围**：
- URL查询参数
- JSON请求体内容
- 表单数据 (application/x-www-form-urlencoded)
- 可疑请求头

**User-Agent过滤**：
```yaml
blocked_user_agents:
  - "curl"
  - "wget"  
  - "python-requests"
  - "bot"
  - "crawler"
  - "spider"
```

**请求大小限制**：
- 普通接口：10MB
- 管理员接口：5MB

### 5. 频率限制中间件 (`ratelimit.go`) ⚠️ **已废弃**

> **注意**：此文件已被 `security.go` 中的频率限制功能替代，建议使用新的安全中间件。

## 📄 配置文件支持

安全中间件支持通过 `configs/config.yaml` 进行配置：

```yaml
# API安全配置
security:
  # 频率限制配置
  rate_limit:
    enabled: true
    max_requests: 100          # IP级别最大请求数/分钟
    window_minutes: 1          # 时间窗口（分钟）
    user_max_requests: 300     # 用户级别最大请求数/分钟
    user_window_minutes: 1     # 用户时间窗口（分钟）
    
  # 安全HTTP头配置
  security_headers:
    enabled: true
    content_security_policy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'..."
    x_frame_options: "SAMEORIGIN"
    x_content_type_options: "nosniff"
    referrer_policy: "strict-origin-when-cross-origin"
    strict_transport_security: "max-age=31536000; includeSubDomains"
    
  # 输入验证配置
  input_validation:
    enabled: true
    max_request_size_mb: 10    # 最大请求体大小（MB）
    blocked_user_agents:       # 阻止的User-Agent
      - "curl"
      - "wget"
      - "python-requests"
      
  # 管理员接口安全配置
  admin_security:
    enabled: true
    max_requests: 30           # 管理员接口IP级别限制
    user_max_requests: 50      # 管理员接口用户级别限制
    ip_whitelist: []           # IP白名单（空表示不启用）
```

## 🚀 使用指南

### 基础使用

```go
// router/router.go
func NewRouter() *Router {
    engine := gin.New()

    // 设置全局中间件（推荐顺序）
    engine.Use(middleware.Logger())                    // 1. 日志记录
    engine.Use(gin.Recovery())                         // 2. 恢复中间件
    engine.Use(middleware.RequestID())                 // 3. 请求ID
    engine.Use(middleware.CORS())                      // 4. CORS支持
    engine.Use(middleware.SecurityMiddleware(          // 5. 安全防护
        middleware.DefaultSecurityConfig()))

    return &Router{engine: engine}
}
```

### 从配置文件使用

```go
// 从配置文件创建安全中间件
cfg := config.Get()
engine.Use(middleware.SecurityMiddlewareFromConfig(cfg))
```

### 路由级别中间件

```go
// 用户路由
userGroup := api.Group("/users")
{
    // 公开路由
    userGroup.POST("/create", userHandler.CreateUser)
    userGroup.POST("/login", userHandler.Login)

    // 认证路由
    authGroup := userGroup.Group("")
    authGroup.Use(middleware.Auth(jwtService))
    {
        authGroup.POST("/get", userHandler.GetUserByID)
        
        // 管理员路由（更严格的安全策略）
        adminGroup := authGroup.Group("")
        adminGroup.Use(middleware.AdminSecurityMiddleware())
        {
            adminGroup.POST("/delete", userHandler.DeleteUser)
            adminGroup.POST("/list", userHandler.GetUserList)
        }
    }
}
```

### IP白名单使用

```go
// 仅允许特定IP访问的管理接口
adminIPs := []string{"192.168.1.100", "10.0.0.50"}
adminGroup.Use(middleware.IPWhitelistMiddleware(adminIPs))
```

## 🔧 自定义配置

### 自定义安全配置

```go
// 创建自定义安全配置
customConfig := &middleware.SecurityConfig{
    RateLimit: struct {
        Enabled        bool          `json:"enabled"`
        MaxRequests    int           `json:"max_requests"`
        Window         time.Duration `json:"window"`
        UserMaxRequest int           `json:"user_max_requests"`
        UserWindow     time.Duration `json:"user_window"`
    }{
        Enabled:        true,
        MaxRequests:    50,           // 更严格的限制
        Window:         time.Minute,
        UserMaxRequest: 200,
        UserWindow:     time.Minute,
    },
    SecurityHeaders: struct {
        Enabled               bool   `json:"enabled"`
        ContentSecurityPolicy string `json:"content_security_policy"`
        XFrameOptions         string `json:"x_frame_options"`
        XContentTypeOptions   string `json:"x_content_type_options"`
        ReferrerPolicy        string `json:"referrer_policy"`
        StrictTransportSec    string `json:"strict_transport_security"`
    }{
        Enabled: true,
        ContentSecurityPolicy: "default-src 'none'", // 更严格的CSP
        XFrameOptions:         "DENY",
        XContentTypeOptions:   "nosniff",
        ReferrerPolicy:        "no-referrer",
        StrictTransportSec:    "max-age=63072000; includeSubDomains; preload",
    },
    InputValidation: struct {
        Enabled           bool     `json:"enabled"`
        MaxRequestSize    int64    `json:"max_request_size"`
        BlockedPatterns   []string `json:"blocked_patterns"`
        AllowedUserAgents []string `json:"allowed_user_agents"`
        BlockedUserAgents []string `json:"blocked_user_agents"`
    }{
        Enabled:           true,
        MaxRequestSize:    1024 * 1024, // 1MB限制
        BlockedPatterns:   middleware.GetDefaultBlockedPatterns(),
        BlockedUserAgents: []string{"*bot*", "*crawler*", "*scraper*"},
    },
}

engine.Use(middleware.SecurityMiddleware(customConfig))
```

## 📊 监控和调试

### 请求日志格式

```json
{
  "request_id": "req_abc123def456789",
  "timestamp": "2025-07-28T10:30:00Z",
  "method": "POST",
  "path": "/api/users/login",
  "status": 200,
  "latency": "45.234ms",
  "client_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "response_size": 1024,
  "error": null
}
```

### 安全事件日志

中间件会记录以下安全事件：
- 频率限制触发
- 恶意输入检测
- User-Agent过滤
- IP白名单拒绝
- 认证失败

### 性能监控

```go
// 获取频率限制器状态（开发/调试用）
limiter := middleware.NewRateLimiter(100, time.Minute)
// limiter 提供内部状态访问方法（如需要）
```

## ⚠️ 安全注意事项

### 1. 配置安全

- **密钥管理**：JWT密钥应通过环境变量设置，不要硬编码
- **HTTPS强制**：生产环境必须启用HTTPS
- **头部验证**：不要信任客户端提供的IP头（X-Forwarded-For等）

### 2. 频率限制

- **合理设置**：过于严格可能影响正常用户体验
- **分层限制**：区分不同接口的限制策略
- **白名单**：为可信IP提供白名单机制

### 3. 输入验证

- **多层防护**：中间件验证 + 业务层验证
- **误报处理**：定期审查被阻止的请求，调整规则
- **更新模式**：及时更新恶意模式库

### 4. 日志安全

- **敏感信息**：避免记录密码、令牌等敏感数据
- **日志轮转**：配置适当的日志轮转策略
- **访问控制**：限制日志文件的访问权限

## 🔄 升级和迁移

### 从旧版本迁移

如果你在使用旧的 `ratelimit.go`：

```go
// 旧版本 ❌
engine.Use(middleware.RateLimit(100, time.Minute))

// 新版本 ✅
engine.Use(middleware.SecurityMiddleware(middleware.DefaultSecurityConfig()))
```

### 版本兼容性

- **v1.0.0+**：支持基础认证和CORS
- **v1.1.0+**：增加日志和请求ID
- **v1.2.0+**：增加频率限制（已废弃）
- **v2.0.0+**：综合安全中间件（当前版本）

## 📚 相关文档

- [JWT认证服务文档](../service/README.md#JWT服务)
- [配置管理文档](../config/README.md)
- [API接口文档](../../docs/api/)
- [安全最佳实践](../../docs/security.md)

## 🐛 故障排除

### 常见问题

**Q: 频率限制过于严格，如何调整？**

A: 修改 `configs/config.yaml` 中的 `security.rate_limit` 配置，或使用自定义配置。

**Q: 某些合法请求被误判为恶意内容？**

A: 检查 `getDefaultBlockedPatterns()` 中的正则表达式，根据需要调整模式。

**Q: CORS问题导致前端无法访问API？** 

A: 确认 `cors.go` 中的 `AllowOrigins` 包含你的前端域名。

**Q: JWT认证失败？**

A: 检查令牌格式、过期时间、签名密钥配置。

### 调试模式

开发环境下可以启用详细日志：

```go
// 开发环境启用调试日志
gin.SetMode(gin.DebugMode)
```

## 🔐 RBAC权限控制中间件 (`rbac.go`) ⭐ **新增**

**功能**：基于角色的访问控制(RBAC)，提供细粒度的权限管理

### 角色定义

系统定义了四种用户角色，按权限从低到高排列：

1. **user (用户)** - 普通用户，只读权限
2. **editor (编辑者)** - 内容编辑者，可发布和管理文章
3. **admin (管理员)** - 系统管理员，可管理用户和内容
4. **superadmin (超级管理员)** - 系统最高权限

### 中间件列表

```go
// 权限验证中间件
func RequirePermission(jwtService, userRepo, rbacService, permissions...) gin.HandlerFunc
func RequireAllPermissions(jwtService, userRepo, rbacService, permissions...) gin.HandlerFunc

// 角色级别验证中间件
func RequireRoleLevel(jwtService, userRepo, rbacService, minRole) gin.HandlerFunc
func RequireSuperAdmin(jwtService, userRepo) gin.HandlerFunc
func RequireAdminOrAbove(jwtService, userRepo) gin.HandlerFunc
func RequireEditorOrAbove(jwtService, userRepo) gin.HandlerFunc

// 资源所有权验证中间件
func RequireOwnershipOrAdmin(jwtService, userRepo, getResourceOwnerID) gin.HandlerFunc
func CanManageUserRole(jwtService, userRepo, rbacService, getTargetRole) gin.HandlerFunc
```

### 使用示例

```go
// 用户管理路由
userGroup := api.Group("/users")
{
    // 需要用户创建权限
    userGroup.POST("/create", 
        middleware.RequirePermission(jwtService, userRepo, rbacService, 
            service.PermissionUserCreate),
        handler.CreateUser)

    // 需要管理员或更高权限
    userGroup.POST("/list",
        middleware.RequireAdminOrAbove(jwtService, userRepo),
        handler.GetUserList)

    // 需要超级管理员权限
    userGroup.POST("/system-config",
        middleware.RequireSuperAdmin(jwtService, userRepo),
        handler.UpdateSystemConfig)
}
```

### 权限列表

#### 系统管理权限
- `system:config` - 系统配置管理（超级管理员）
- `system:logs` - 系统日志查看（管理员及以上）
- `system:stats` - 系统统计信息（管理员及以上）

#### 用户管理权限
- `user:create` - 创建用户（管理员及以上）
- `user:read` - 查看用户信息（编辑者及以上）
- `user:update` - 更新用户信息（管理员及以上）
- `user:delete` - 删除用户（管理员及以上）
- `user:list` - 用户列表（管理员及以上）

#### 文章管理权限
- `article:create` - 创建文章（编辑者及以上）
- `article:read` - 查看文章（所有用户）
- `article:update` - 更新文章（编辑者及以上）
- `article:delete` - 删除文章（编辑者及以上）
- `article:publish` - 发布文章（编辑者及以上）

#### 评论管理权限
- `comment:create` - 发表评论（所有用户）
- `comment:read` - 查看评论（所有用户）
- `comment:update` - 更新评论（仅本人或管理员）
- `comment:delete` - 删除评论（编辑者及以上）
- `comment:moderate` - 评论审核（管理员及以上）

### 获取用户信息

```go
func (h *Handler) SomeHandler(c *gin.Context) {
    // 获取当前用户完整信息
    user, exists := middleware.GetCurrentUser(c)
    if !exists {
        response.Unauthorized(c, "未找到用户信息")
        return
    }

    // 获取当前用户ID和角色
    userID, _ := middleware.GetCurrentUserID(c)
    userRole, _ := middleware.GetCurrentUserRole(c)
    
    // 业务逻辑处理...
}
```

### 安全特性

1. **角色层级控制**: 高级别角色自动拥有低级别权限
2. **超级管理员保护**: 超级管理员角色不可被降级
3. **用户状态验证**: 自动检查用户是否被禁用
4. **资源所有权**: 支持用户只能操作自己资源的限制
5. **角色转换验证**: 严格验证角色变更的合法性

### 错误响应

- `401 Unauthorized` - 未提供认证令牌或令牌无效
- `403 Forbidden` - 权限不足或用户被禁用

详细的RBAC使用文档请参考 [RBAC使用指南](./rbac_guide.md)
