# Config 模块

配置加载与校验。配置项清单以 `configs/config.yaml` 与 `config.go` 结构体的 `mapstructure` tag 为唯一来源，本文件不重复罗列字段。

## 加载与访问

- `Load(configPath)` 经 `sync.Once` 单例加载，进程内只生效一次；`Get()` 在未初始化时 panic；`GetDSN()`、`GetServerAddress()` 提供派生值。
- 代码默认值集中在 `setDefaults()`，YAML 缺项时按默认值生效。
- 标量配置项可经 `MYBLOG_<SECTION>_<FIELD>` 环境变量覆盖，例如 `server.port` 对应 `MYBLOG_SERVER_PORT`；列表与映射无法经单个变量无损表达，保持 YAML 原值。

## 校验规则

`validateConfig()` 在加载期拒绝以下配置：

- `server.port` 超出 1 至 65535。
- `database.host`、`database.username`、`database.dbname` 为空。
- `token.access_expire` 或 `token.refresh_expire` 不为正数。
- `rbac.role_hierarchy` 或 `rbac.role_permissions` 缺少 superadmin、admin、editor、user 任一类角色的定义。
- `i18n.supported_languages` 为空，或 `i18n.default_language` 不在该列表内。

## 约定

- 令牌为服务端签发的不透明随机串，不存在签名密钥与签发者配置；`token.cookie_secure` 控制会话 Cookie 是否仅经 HTTPS 传输，生产环境必须开启。
- 权限映射的生产环境唯一权威为 `rbac` 节，服务层经 `service.LoadRBACConfig` 读取，见 `cmd/myblog/deps.go`。
- 各子配置节的字段含义以 `configs/config.yaml` 注释为准。

相关约束见 `AGENTS.md` 第 5 节与 `docs/development.md`「核心模块」。
