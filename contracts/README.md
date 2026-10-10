# contracts/ —— 全栈契约面

> 本目录是前后端共享事实的物理锚点，所有跨端争论最终指向本目录下的同一份文件。

## 契约面清单

| 契约 | 文件 | 权威方 | 状态 |
|---|---|---|---|
| API 形状 | 后端 `pkg/response` + `@myblog/shared` 的 `ApiResponse` | 后端 | ✅ 已固化 |
| 错误码 | `errors.yaml` | 本文件（唯一编辑点） | 🟡 传输层码已固化，业务码细分待落地 |
| 认证协议 | `auth-protocol.md` | 本文档 | ✅ 已文档化（不透明令牌 + Cookie 会话） |
| 权限下发 | 后端 RBAC，登录响应 `permissions[]` | `server/configs/config.yaml` 的 `rbac` 节 | ✅ 已落地 |
| 类型生成 | 三把锁方案（见 `docs/architecture-rules.md` §3.4） | 后端 DTO / fixtures | ✅ 已落地 |
| 内容语言协商 | `i18n-protocol.md` | 本文档 | ✅ 已落地（文章/分类/标签/字典） |

## fixtures —— 金样本响应

`fixtures/` 存放双端共用的契约金样本，基于真实响应生成，用于锁定响应形状。
配套的 Go 金样本测试、TS 类型锚定与 eslint 导入守门三把锁，机制与落地清单见
`docs/architecture-rules.md` §3.4；任一环漂移即测试变红。

## 编辑纪律

- wire 格式变更流程（影响面声明 → 双端同窗口变更 → 错误语义与认证协议同步）见
  `docs/architecture-rules.md` §4.1，改认证协议须先改 `auth-protocol.md`。
- 新增业务模块必须先登记模块对齐表（见 `docs/architecture-rules.md` §4.2）。
- 新增 not-found 哨兵错误须同步扩展 `handler/error.go` 的映射集合与 `errors.yaml`。
- 前端不新增权限/类型/刷新的手写副本，一律消费后端下发值或 `@myblog/api` 类型。
