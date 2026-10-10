# Repository 模块

数据访问层：只承载持久化实现，不含业务规则。

## 本层约定

- 领域实体与请求/响应 DTO 统一来自 `internal/domain`，GORM 实体在 `internal/model`；**禁止在本包新增实体或 DTO 定义**。
- 各业务域接口声明在本包，命名统一为 `XxxRepositoryInterface`（存量例外 `UserRepository`），service 只依赖接口。
- 查询未命中返回 domain 哨兵错误（如 `domain.ErrUserNotFound`），由 service 与 handler 决定响应语义；数据库错误一律经 `fmt.Errorf` 包装上下文后向上传递。
- 删除使用 GORM 软删除；唯一性与索引约束声明在实体的 `gorm` tag 上（用户名、邮箱为唯一索引，`domain.User` 见 `internal/domain/user.go`）。
- 多段写库在同一事务内提交，沿用既有的 `xxxTx` 事务方法先例。

方法签名（含分页与过滤参数，如用户仓库的 `List(offset, limit int, keyword string)`）以各 `repository/*.go` 的接口声明为准。

相关约束见 `AGENTS.md` 第 5 节「类型归属」「数据库」。
