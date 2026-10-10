# 债务登记：安全、隐私与运行时

> 本文件是 `docs/architecture-rules.md` 第 8 节债务登记表的「安全、隐私与运行时」分册。
> 通用机制、条目字段含义与维护流程见 [`README.md`](./README.md)。

## 1. 公开端点直出实体泄漏个人信息

- **状态**：存量，多个域已收紧，红线持续适用。
- **基线**：
  - 评论域与作者域：`model.Comment` 的游客邮箱、IP、UserAgent 改为 `json:"-"`；评论 `user` 与文章 `author` 经 `domain.AuthorPublic` 窄化视图输出；管理端审计走 `AdminCommentView`。
  - 用户域：`/users/get` 由仅挂基础认证改为挂 `user:list` 权限，仅限管理端访问；普通用户查看他人资料走窄化的 `publicProfile` 端点；回归测试在 `internal/router/user_rbac_test.go`。
  - 媒体域：上传者经 `domain.UploaderPublic` 窄化视图输出，`uploadIP` 审计字段改 `json:"-"`；媒体详情补齐水平越权校验（与列表归属规则对称）；回归测试在 `model/media_public_test.go` 与 `service/media_test.go`。
  - 尚未收敛：友链公开列表仍直出 `model.FriendlyLink`，包含站长 `contactEmail` 与管理员 `note`。
- **验证命令**：读 `model/comment.go`、`model/article.go`、`model/media.go` 的 json tag 与 `domain/author.go`；service 层测试断言公开响应无审计字段；`go test ./internal/router/ -run TestUsersGetRequiresUserListPermission -count=1`。
- **红线**：公开端点输出个人信息前必须经窄化 DTO 或字段白名单；实体新增隐私/审计字段默认 `json:"-"`。

## 2. 文章响应泄漏作者审计字段

- **状态**：已修复。
- **基线**：`lastLoginIP` 等审计字段改为 `json:"-"`。
- **验证命令**：读 `domain/user.go` 的 json tag。
- **红线**：新增审计字段默认 `json:"-"`。

## 3. WAF 内容级黑名单的固有误伤面

- **状态**：存量，受接受的残余风险。
- **基线**：误伤回归用例已就位，默认模式全部经词首边界或取值上下文锚定，攻击拦截与误伤回归两组用例在位；残余风险为讲解 SQL/XSS 的技术文章正文命中关键词模式仍会被拦，根治需内容感知解析或按路由豁免。
- **验证命令**：`go test ./internal/middleware/ -run "TestDefaultBlockedPatterns\|TestSecurityMiddleware" -count=1`。
- **红线**：新增或修改阻止模式必须先红后绿配「攻击拦截 + 误伤回归」两组用例，禁止回退宽匹配。

## 4. 令牌表为内存 map

- **状态**：存量，单实例部署前提。
- **基线**：已加锁并支持过期惰性清理（`sync.RWMutex` 保护，过期记录随签发清理）；令牌为服务端签发的不透明随机串，身份唯一权威在服务端令牌表。
- **验证命令**：`git grep -n "tokensByUser" -- server`。
- **红线**：持久化前保持锁；服务重启即全部会话失效；多副本部署前必须先完成持久化改造（见 `docs/architecture-rules.md` §7.2）。
