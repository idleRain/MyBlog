# 债务登记：安全、隐私与运行时

> 本文件是 `docs/architecture-rules.md` 第 8 节债务登记表的「安全、隐私与运行时」分册。
> 通用机制、条目字段含义与维护流程见 [`README.md`](./README.md)。

## 1. 公开端点直出实体泄漏个人信息

- **状态**：已修复。
- **基线**：
  - 评论域与作者域：`model.Comment` 的游客邮箱、IP、UserAgent 改为 `json:"-"`；评论 `user` 与文章 `author` 经 `domain.AuthorPublic` 窄化视图输出；管理端审计走 `AdminCommentView`。
  - 用户域：`/users/get` 由仅挂基础认证改为挂 `user:list` 权限，仅限管理端访问；普通用户查看他人资料走窄化的 `publicProfile` 端点；回归测试在 `internal/router/user_rbac_test.go`。
  - 媒体域：上传者经 `domain.UploaderPublic` 窄化视图输出，`uploadIP` 审计字段改 `json:"-"`；媒体详情补齐水平越权校验（与列表归属规则对称）；回归测试在 `model/media_public_test.go` 与 `service/media_test.go`。
  - 友链域：公开列表与访客申请响应改经 `service.FriendlyLinkPublic` 窄化视图，只输出 `id`、`name`、`url`、`logo`、`description`、`sortOrder`、`isReciprocal`；站长 `contactEmail` 与管理员 `note` 不再随公开接口输出，管理端列表仍输出完整实体供审核；回归测试在 `internal/service/friendly_link_test.go` 的 `TestListVisibleLinksHidesContactFields` 与 `TestListLinksKeepsAdminAuditFields`。
- **验证命令**：读 `model/comment.go`、`model/article.go`、`model/media.go` 的 json tag 与 `domain/author.go`；service 层测试断言公开响应不含联系邮箱与备注（`go test ./internal/service/ -run "FriendlyLink|HidesContact" -count=1`）；`go test ./internal/router/ -run TestUsersGetRequiresUserListPermission -count=1`。
- **红线**：公开端点输出个人信息前必须经窄化 DTO 或字段白名单；实体新增隐私/审计字段默认 `json:"-"`。

## 2. 文章响应泄漏作者审计字段

- **状态**：已修复。
- **基线**：`lastLoginIP` 等审计字段改为 `json:"-"`。
- **验证命令**：读 `domain/user.go` 的 json tag。
- **红线**：新增审计字段默认 `json:"-"`。

## 3. WAF 内容级黑名单的固有误伤面

- **状态**：已修复（正文误伤面消除），残余风险为正文豁免依赖渲染层消毒。
- **基线**：JSON 请求体改按字段路径扫描（`richContentJSONPaths`），正文（`content` 与 `i18n.*.content`）不参与恶意模式扫描，标题、摘要、SEO 字段、URL 参数、请求头与表单参数保持全量扫描；结构无法解析时退化为整段扫描。按解码后的取值扫描，JSON 转义序列不再遮蔽模式中的词首边界。默认模式仍全部经词首边界或取值上下文锚定，攻击拦截与误伤回归两组用例在位。
- **验证命令**：`go test ./internal/middleware/ -run "TestDefaultBlockedPatterns\|TestSecurityMiddleware" -count=1`。
- **红线**：
  1. 新增或修改阻止模式必须先红后绿配「攻击拦截 + 误伤回归」两组用例，禁止回退宽匹配；
  2. 正文豁免以渲染层消毒为前提，扩大豁免清单前必须确认该字段的每一处输出都经消毒渲染或转义输出；
  3. 删除或放宽渲染层消毒（`pkg/markdown` 的 `Unsafe`、`apps/web` 的 `renderMarkdown`）时，必须同步收回对应的正文豁免。

## 4. 代理信任链未收敛导致 `ClientIP()` 可伪造

- **状态**：存量（真实缺陷，与部署副本数无关）。
- **基线**：
  - `internal/router/router.go` 以 `gin.New()` 建引擎，全仓无 `SetTrustedProxies` 调用；gin v1.10.1 的 `New()` 默认 `trustedProxies` 为 `0.0.0.0/0` 与 `::/0`，即信任全部代理。
  - 全信任判定下 `ClientIP()` 经 `validateHeader` 返回 `X-Forwarded-For` 的**最左值**（`i == 0` 分支），而 `deploy/nginx.conf` 的 `$proxy_add_x_forwarded_for` 把真实地址追加到最右，客户端自带的伪造值因此被原样采信。修复前的表述「经网关转发后 `ClientIP()` 恒为网关地址」与实测不符。
  - 受影响面：IP 级限流（`middleware/ratelimit.go` 与 `middleware/security.go` 的 ipLimiter，默认 100 次/分钟，换一个伪造值即换一个计数桶）；审计 IP 落库（`handler/article.go` → `service/article.go` 的 `IPAddress`、`handler/media.go` → `service/media.go` 的 `UploadIP`）；未登录访客去重键（`service/article.go` 在无登录态时以 ipAddress 作 visitorKey，伪造值即可重复计入浏览）；管理员接口 IP 白名单（`middleware/security.go`，`admin_security.ip_whitelist` 当前为空故未启用）。
  - 不受影响：登录失败锁定按 `user.ID` 落库（`service/user.go` 的 `IncrementLoginFailures` 与 `LockUserAfterFailures`），WAF 无 IP 黑名单，故早前登记的「影响限流、登录失败锁定、访问审计与 WAF 四项」表述不成立。
- **验证命令**：`git grep -n "SetTrustedProxies" -- server`（输出为空表示信任链仍未收敛，收敛后应命中 `internal/router`）；`git grep -n "ClientIP()" -- server` 为现有 7 处判定点清单。
- **红线**：收敛信任链前禁止新增以 `ClientIP()` 为唯一依据的安全判定（限流、白名单、封禁、审计）；不得以 `ClientIP()` 作为可信身份依据；修复须补「伪造 `X-Forwarded-For` 不影响 `ClientIP()`」的回归用例。
