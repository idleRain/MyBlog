# 数据库文档

本目录包含 MyBlog 项目的数据库相关文档。

## 文件说明

### schema.sql
**用途**: 数据库架构设计参考文档，共 30 张业务表的完整 DDL
**状态**: 仅供参考，不用于实际迁移

包含：
- 30 张业务表的完整 DDL，覆盖用户、内容（含文章/分类/标签翻译表）、评论、互动、媒体、站点运营、字典、统计日志 8 个模块
- 所有索引、复合唯一索引、全文索引、检查约束与外键级联策略
- 默认数据插入脚本

### database-architecture.md
**用途**: 数据库架构设计说明文档
**状态**: 设计原则、表结构与演进约定的权威说明

## 重要说明

**⚠️ 实际数据库迁移方式**

项目使用 **GORM AutoMigrate 与 golang-migrate 双轨** 管理数据库结构，由 `cmd/myblog/main.go` 的 `initDatabase` 按运行模式选择：

1. **模型定义位置**: `server/internal/model/*.go`
2. **开发轨**（`server.mode = debug`）: 启动时执行 `server/internal/database/migrate.go` 的 `AutoMigrateWithFix`，经 `model.AutoMigrate` 建表并同步表注释
3. **生产轨**（非 debug 模式）: 启动时执行同文件的 `RunMigrations`，经 golang-migrate 执行 `server/migrations/` 下的增量迁移，当前基线为 `000001_init_schema`
4. **表注释同步**: 开发轨迁移完成后由 `model/table_comments.go` 统一同步表级注释
5. **全文索引补建**: 迁移后由 `internal/database/fulltext.go` 幂等补建 articles 与 article_translations 的 ngram 全文索引，GORM 无法声明 FULLTEXT
6. **字典种子**: `internal/database/seed_dict.go` 的 `SeedDicts` 幂等写入初始字典，两条迁移路径共用

**如何添加新表**:

1. 在 `server/internal/model/` 中创建 Go 模型
2. 在 `server/internal/model/models.go` 的 `Models()` 函数中注册
3. 在 `server/internal/model/table_comments.go` 的表注释映射中补充
4. 重启应用，开发环境 GORM 会自动创建表结构
5. 同步在 `server/migrations/` 手写增量迁移，生产环境以迁移版本为准

**如何修改表结构**:

1. 修改对应的 Go 模型结构
2. 开发环境重启应用，GORM 自动同步变更；生产环境手写增量迁移执行
3. 注意：删除字段需要手动处理

## 数据库连接

- **数据库**: MySQL 8.0 (通过 GORM)
- **配置文件**: `server/configs/config.yaml`

## 参考资源

- [数据库架构设计](../database-architecture.md)
- [GORM 模型定义](../../server/internal/model/)
- [增量迁移脚本](../../server/migrations/)
- [数据库配置说明](../../server/configs/config.yaml)
