# 债务登记：结构与类型

> 本文件是 `docs/architecture-rules.md` 第 8 节债务登记表的「结构与类型」分册。
> 通用机制、条目字段含义与维护流程见 [`README.md`](./README.md)。

## 1. service/middleware/router import repository

- **状态**：存量，只减不增。
- **基线**：service **12** + middleware **1** + router **0**，合计 13 个文件。router 已归零，middleware 仅 `identity.go`；`*_test.go` 测试替身文件不计入。dict 模块接入曾使 service 层由 11 增至 12。
- **验证命令**：见 `docs/architecture-rules.md` §2.3。必须在 `server/` 目录下执行（相对 pathspec 在仓库根执行会得到零命中的假绿），并带 `':(exclude)*_test.go'` 排除测试文件；不排除时命中数为 33。
- **红线**：只减不增。

## 2. router 重复定义 handler 接口与 `interface{}` 断言

- **状态**：已修复。
- **基线**：router 包重复接口 0、`interface{}` 运行时断言 0；接口统一声明在 `handler`、`service`、`repository` 各自实现包内。
- **验证命令**：`git grep -c "HandlerInterface interface" -- server/internal/router`（应为空）。
- **红线**：禁止在 `router` 包重新引入接口定义或运行时断言。

## 3. `RBACService` 生产实例化

- **状态**：已修复。
- **基线**：生产实例化点 1 处，即组合根 `cmd/myblog/deps.go`（`main.go` 管生命周期，`deps.go` 按域装配，二者同属组合根）。
- **验证命令**：见 `docs/architecture-rules.md` §5.3。
- **红线**：禁止新增实例化点；依赖一律经构造函数注入。

## 4. 双 User 模型同写 users 表

- **状态**：已修复。
- **基线**：合并为唯一 `domain.User` 实体，`model.User` 为类型别名。
- **验证命令**：`git grep -n "type User struct" -- server/internal '*.go'`（仅命中 `domain/user.go`）。
- **红线**：新字段只加 `domain.User`。

## 5. 应用层影子类型层

- **状态**：已修复。
- **基线**：`types/api.d.ts` 与 admin `lib/types`（admin/common/auth/index 共 535 行）已删除，admin `lib/types` 目录不存在；两应用 eslint 守门由 paths 改为 patterns，拦截 `$lib/types` 的全部引入形态。
- **验证命令**：`git grep -n "interface BaseApiResponse" -- apps`（应为空）；`git ls-files apps/admin/src/lib/types`（应为空）。
- **红线**：禁止重新引入；接口类型一律来自 `@myblog/api`，通用信封类型一律来自 `@myblog/shared`。

## 6. follow 模块仅后端

- **状态**：已修复。
- **基线**：`@myblog/api/modules/follow` 与两应用注册均已落地；web 作者页 `components/user/FollowButton.svelte` 已消费 follow 与 isFollowing 接口。
- **验证命令**：`git grep -ln "FollowAPI" -- apps/web/src`（非空即已消费）。
- **红线**：关注数据仅经 service 域端点读写，不得由前端拼装或旁路。
