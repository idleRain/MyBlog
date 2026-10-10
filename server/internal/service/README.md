# Service 模块

业务逻辑层。业务规则（状态流转、slug 生成、密码强度、权限判定）的唯一权威在本层，repository 只做持久化，handler 只做协议转换。

## 本层约定

- 各业务域接口声明在 service 包内，命名统一为 `XxxServiceInterface`（存量例外 `UserService`、`RBACService`），各层只依赖接口。
- 依赖一律经构造函数注入，组合根为 `cmd/myblog/`（`main.go` 管生命周期、`deps.go` 按域装配）；可选依赖用选项函数注入，如 `NewUserService(userRepo, tokenService, rbacService, WithLoginLockoutPolicy(...))`。
- 业务错误集中在 `errors.go` 的哨兵错误（`ErrPermissionDenied`、`ErrInvalidRequest`、`ErrUsernameTaken`、`ErrEmailTaken`），handler 经 `errors.Is` 分档映射，不得另起裸 `errors.New` 表达同类语义。
- 多段写库在 repository 的事务方法内完成，service 不自行拼装事务。
- 权限判定经 `RBACService`，权限映射的生产环境唯一权威为 `configs/config.yaml` 的 `rbac` 节，由 `LoadRBACConfig` 加载（见 `cmd/myblog/deps.go`）。

## 认证与密码

- 登录先按用户名查找，未命中回退邮箱；锁定检查先于密码校验；失败计数原子自增，达到 `security.login_lockout` 阈值写入 `locked_until`，登录成功清零。
- 密码使用 bcrypt，成本常量 `BcryptCost = 12`；强度校验 `validatePasswordStrength` 要求 6 至 100 位、同时含字母与数字、拒绝常见弱密码。
- 刷新令牌换取新令牌对前查库校验用户存在且状态正常；改密成功后经 `RevokeUserTokens` 撤销该用户全部既有令牌。

方法与字段清单以各 `service/*.go` 的接口声明为准。相关约束见 `AGENTS.md` 第 5 节与 `docs/architecture-rules.md` 第 5 节。
