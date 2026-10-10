# Handler 模块

HTTP 协议层：参数绑定与校验、调用 service、以统一格式输出响应，业务规则不落在本层。

## 本层约定

- 各模块的处理器接口声明在 handler 包内，命名统一为 `XxxHandlerInterface`（用户模块为 `UserHandlerInterface`），经 `router.Dependencies` 由组合根注入；禁止包内私自实例化 service。
- 请求参数经 `c.ShouldBindJSON` 与 DTO 的 `binding` tag 完成必填、长度、格式、枚举校验，校验失败返回 400；业务规则由 service 校验。
- service 返回的错误统一交 `HandleServiceError`（`error.go`）分档：哨兵错误映射 404/403/400，未分类错误脱敏为「服务器内部错误」。禁止在 handler 内直接 `response.InternalError(c, err.Error())`，也禁止另起私有映射分支；新增 not-found 哨兵须同步扩展该映射集合与 `contracts/errors.yaml`。
- DTO 与领域类型取自 `internal/domain`；**禁止直接序列化 `internal/model` 实体**（公开端点直出实体是已登记债务，见 `docs/debt/security-runtime.md` 第 1 条，触碰时应改为窄化 DTO 或字段白名单）。
- 路由注册在 `internal/router/<module>.go`，业务接口一律 `POST`；`GET` 仅用于健康探针（`internal/router/health.go`）。
- 请求语言经 `getRequestLanguage`（`common.go`）读取 language 中间件写入上下文的语言标识。

## 端点清单

各模块的路由、请求体与响应字段以 `server/docs/api/README.md` 为唯一索引，本文件不重复罗列。

相关约束见 `AGENTS.md` 第 5 节。
