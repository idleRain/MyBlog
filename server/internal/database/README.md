# Database 模块

MySQL 连接、结构迁移与种子数据的实现包。GORM 实体在 `internal/model`，其中用户实体为 `internal/domain.User`（`model.User` 是兼容别名）；连接参数来自 `configs/config.yaml` 的 `database` 节。

## 连接管理（mysql.go）

- `InitMySQL(cfg *config.Config)` 经 `sync.Once` 单例初始化：先由 `createDatabaseIfNotExists` 建库，再建立连接、设置连接池（`MaxIdleConns`、`MaxOpenConns`、`ConnMaxLifetime` 1 小时）并 Ping 验证。
- `GetDB()` 在未初始化时 panic；`Close()` 关闭底层连接；`HealthCheck()` 供就绪探针调用。
- `AutoMigrate(models ...interface{})` 会忽略入参并固定委托 `AutoMigrateWithFix`，不要依赖它迁移指定模型；迁移对象由 `model.Models()` 注册决定。

## 结构迁移（migrate.go）

- `RunMigrations`、`MigrateDown`、`MigrateToVersion`、`GetMigrationVersion`、`ForceMigrationVersion` 基于 golang-migrate 执行 `server/migrations/` 下的增量迁移，迁移链路自行追加 `multiStatements=true`，不污染 GORM 使用的 DSN。
- `AutoMigrateWithFix` 委托 `model.AutoMigrate`，仅供 debug 模式快速同步。两条轨道的分工与演进纪律见 `docs/database-architecture.md`「迁移双轨」。
- 组合根按 `server.mode` 二选一调度，见 `cmd/myblog/main.go` 的 `initDatabase`。

## 全文索引与种子（fulltext.go、seed.go、seed_dict.go）

- `EnsureArticleFulltextIndex`、`EnsureArticleTranslationFulltextIndex` 幂等补建 ngram 全文索引，GORM AutoMigrate 无法声明 FULLTEXT 索引。
- `EnsureSuperAdmin` 写入超级管理员，仅由 `cmd/seed` 调用；`SeedDicts` 幂等写入字典种子，在组合根启动链路调用。
- `seed:admin` 命令的默认账户、自定义参数与幂等语义见 `docs/development.md`「初始化管理员账户」。
