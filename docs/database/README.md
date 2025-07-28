# 数据库文档

本目录包含 MyBlog 项目的数据库相关文档。

## 文件说明

### schema.sql
**用途**: 数据库架构设计参考文档
**状态**: 仅供参考，不用于实际迁移

包含完整的数据库表结构定义，包括：
- 12张核心数据表的完整 DDL
- 所有索引、外键约束和关联关系
- 默认数据插入脚本
- MongoDB 缓存配置预设

### 重要说明

**⚠️ 实际数据库迁移方式**

项目使用 **GORM 自动迁移** 作为数据库管理方式：

1. **模型定义位置**: `server/internal/model/*.go`
2. **迁移入口**: `server/internal/database/migrate.go`
3. **执行方式**: 应用启动时自动执行

**如何添加新表**:

1. 在 `server/internal/model/` 中创建 Go 模型
2. 在 `server/internal/model/models.go` 的 `Models()` 函数中注册
3. 重启应用，GORM 会自动创建表结构

**如何修改表结构**:

1. 修改对应的 Go 模型结构
2. 重启应用，GORM 会自动同步变更
3. 注意：删除字段需要手动处理

### 数据库连接

- **主数据库**: MySQL 8.0 (通过 GORM)
- **缓存数据库**: MongoDB (通过官方驱动)
- **配置文件**: `server/configs/config.yaml`

### 参考资源

- [完整的数据库架构设计](./database-architecture.md)
- [GORM 模型定义](../server/internal/model/)
- [数据库配置说明](../server/configs/config.yaml)