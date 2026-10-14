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

- **状态**：已修复（信任链收敛为白名单，配置项 `security.trusted_proxies`）。
- **基线**：
  - `internal/router/router.go` 在构造引擎后立即调用 `engine.SetTrustedProxies(cfg.Security.TrustedProxies)`，取值非法时返回错误并由组合根在启动期终止；`NewRouter` 相应改为返回 `(*Router, error)`。
  - 受信来源唯一权威为 `security.trusted_proxies`，代码默认只信任 `127.0.0.1/32` 与 `::1/128`；容器编排经 `MYBLOG_SECURITY_TRUSTED_PROXIES` 按逗号分隔注入容器网段，compose 默认注入 `172.16.0.0/12`。
  - 配置校验拒绝空项、非法 IP 或 CIDR，以及 `0.0.0.0/0` 与 `::/0` 这类覆盖全部地址的取值，避免回退为信任全部来源。
  - 回归用例：`internal/router/router_trusted_proxy_test.go` 覆盖四态——非受信来源携带伪造 `X-Forwarded-For` 时取连接来源地址、经受信代理转发时取最右侧非受信地址、受信代理单值转发被采信、信任列表为空时一律取连接来源地址；`internal/config/config_test.go` 覆盖非法项与覆盖全部地址的拒绝、环境变量注入与空值语义。接线前实测三例失败，日志字段 `clientIp` 为伪造值。
  - 早前登记的表述已订正：既非「经网关转发后 `ClientIP()` 恒为网关地址」，也非「影响限流、登录失败锁定、访问审计与 WAF 四项」。实际受影响面为 IP 级限流、审计 IP 落库、未登录访客去重键与管理员接口 IP 白名单；登录失败锁定按用户 ID 落库，WAF 无 IP 黑名单，二者均不受影响。
- **验证命令**：`git grep -n "SetTrustedProxies" -- server` 应恰好命中 `internal/router/router.go` 一处实现（README 说明性命中不计）；`go test ./internal/router/ ./internal/config/ -count=1`。
- **红线**：仅 `security.trusted_proxies` 内的来源可参与 `ClientIP()` 解析；禁止回退为信任全部来源，禁止在配置校验中放行覆盖全部地址的取值；新增按 IP 的安全判定一律基于 `ClientIP()`，不得自行解析转发头。
