# 债务登记：结构与类型

> 本文件是 `docs/architecture-rules.md` 第 8 节债务登记表的「结构与类型」分册。
> 通用机制、条目字段含义与维护流程见 [`README.md`](./README.md)。

## 1. service/middleware/router import repository

- **状态**：已修复。
- **基线**：违例边 **0** 个文件。`middleware → repository` 经消费方窄接口收敛——`identity.go` 的 `userLookup` 只声明 `GetByID` 一项能力，仓储实现由组合根注入后结构化满足；`router → repository` 早已归零。不排除 `*_test.go` 时命中 2 个测试替身文件（middleware 与 router 各 1）。
- **允许边**：`service → repository` 属 `docs/architecture-rules.md` §2.1 允许图。service 消费 `repository.XxxRepositoryInterface` 时必然 import 该包，故 service 的 12 个文件不计入违例。首版基线将三者合并计为 13 个文件，实为 12 个允许边加 1 个违例边，口径随本次修复订正。
- **已接受口径**：查询参数与结果结构（`ArticleListParams`、`ArticleArchiveGroup` 等 11 个）随接口声明在 repository 包，属接口签名的组成部分，不迁移至 `internal/domain`。
- **验证命令**：见 `docs/architecture-rules.md` §2.3。必须在 `server/` 目录下执行（相对 pathspec 在仓库根执行会得到零命中的假绿），并带 `':(exclude)*_test.go'` 排除测试文件。
- **红线**：`middleware` 与 `router` 不得 import repository；中间件消费存储能力一律在包内声明窄接口，实现由组合根注入。仓储的具体实现 struct 与 `New*Repository` 构造函数只允许出现在 repository 包与组合根，service 只可消费接口与查询参数。

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

## 7. 业务错误未归哨兵导致分档退化为 500

- **状态**：已修复。
- **基线**：以裸 `errors.New` 或 `fmt.Errorf`（不含 `%w`）表达的业务错误共 **0 处**。修复前基线为 30 处、分布于 7 个文件（`internal/service/user.go` 16、`token.go` 4、`category.go` 3、`friendly_link.go` 2、`media.go` 2、`tag.go` 2、`article.go` 1）；现全部改为包装哨兵：认证失败包装新增的 `service.ErrUnauthenticated`（映射 401）、业务规则校验失败包装 `service.ErrInvalidRequest`（400）、资源未找到包装 `domain.ErrUserNotFound` 与 `domain.ErrCategoryNotFound`（404）。`contracts/errors.yaml` 增补 `'401'` 分类，其「实现偏差」节随之清空删除；`handler/user.go` 的 Login 与 RefreshToken 私有 401 分支、`handler/session.go` 的 CreateSession 私有 401 分支、`handler/media.go` 的上传私有 400 分支与 `handler/user.go` 的角色转换私有 400 分支均改走 `HandleServiceError` 唯一映射点。
- **验证命令**：在 `server/` 目录下执行
  `{ git grep -n 'fmt.Errorf(' -- internal/service ':(exclude)*_test.go' | grep -v '%w'; git grep -n 'errors.New(' -- internal/service ':(exclude)*_test.go' ':(exclude)internal/service/errors.go'; } | wc -l`
  应输出 0；分档一致性由 `go test ./internal/handler/ -run TestErrorMappingMatchesContract -count=1` 与 `pnpm run contract:check` 锚定。
- **红线**：新增业务校验失败、认证失败与资源未找到必须包装对应哨兵（`service.ErrUnauthenticated`、`service.ErrInvalidRequest`、`domain.ErrXxxNotFound` 等），并同步登记 `contracts/errors.yaml` 的 mapping 节与 `handler` 分档注册表；禁止以裸错误表达同类语义，禁止在模块内新增私有映射分支。
