# contracts/ —— 全栈契约面（K）

> 本目录是前后端共享事实的物理锚点，所有跨端争论最终指向本目录下的同一份文件。
> 详细方案见 `temp/health-check/unified-refactoring-plan.md`。

## 契约面清单

| 契约 | 文件 | 权威方 | 状态 |
|---|---|---|---|
| C1 API 形状 | 后端 `pkg/response` + `@myblog/shared` 的 `ApiResponse` | 后端 | ✅ 已固化 |
| C2 错误码 | `errors.yaml` | 本文件（唯一编辑点） | 🟡 初版，业务码细分待落地 |
| C3 认证协议 | `auth-protocol.md` | 本文档 | ✅ 已文档化（含 payload-only 怪癖） |
| C4 权限下发 | 后端 RBAC（终态后端下发） | 后端 | 🟡 阶段 D 落地 |
| C5 类型生成 | 三把锁方案（见下） | 后端 DTO / fixtures | 🟡 阶段 C 落地 |

## fixtures —— 金样本响应

`fixtures/` 存放双端共用的契约金样本。**当前为结构样例**，用于锁定响应形状；
阶段 C 落地三把锁后，将基于真实响应生成字节级金样本，并配套：

1. **Go handler 测试**：断言响应输出与 fixture 逐字节一致。
2. **TS 侧 vitest + expect-type**：`expectTypeOf(fixture).toEqualTypeOf<手写类型>` 双向精确相等。
3. **eslint no-restricted-imports**：禁止应用层定义同构类型。

任一端漂移即测试变红，静默漂移在结构上不可能。

## 编辑纪律

- 改 wire 格式（响应结构、错误码、认证协议）必须先改本目录对应文件。
- 新增业务模块必须先登记模块对齐表（见 `docs/architecture-rules.md` §4.2）。
- 前端不新增权限/类型/刷新的手写副本，一律消费生成物或下发值。
