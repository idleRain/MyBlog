# 备份与恢复操作手册

> 对应体检项 OPS-18：全仓此前零备份，单实例 MySQL 与本地媒体目录构成三重单点。
> 本手册给出每日全量备份的调度方式、恢复步骤与演练流程。

## 1. 备份范围与频率

| 对象 | 方式 | 建议频率 | 保留期 |
|---|---|---|---|
| MySQL 数据库 | `mysqldump --single-transaction` 一致性快照，gzip 压缩 | 每日一次（建议凌晨低峰） | 14 天，可按需调整 |
| uploads 媒体文件 | tar.gz 全量打包 | 随每日备份执行 | 与数据库一致 |

增量与 PITR：如需恢复到任意时间点，需在 MySQL 服务端开启 `log_bin` 并将 binlog 随备份归档；
当前脚本仅做每日全量，RPO 为 24 小时，开启 binlog 后可缩短至分钟级（方案见第 5 节）。

## 2. 每日备份（Linux 服务器）

脚本：`scripts/backup/backup.sh`，行为：

1. `mysqldump --single-transaction` 导出数据库（含触发器、存储过程、事件），压缩为 `database.sql.gz`；
2. 打包 `media.upload_dir` 指向的 uploads 目录为 `uploads.tar.gz`；
3. 生成 `checksums.sha256` 校验清单；
4. 清理超过保留期的过期备份。

cron 配置示例（每日 03:00 执行）：

```cron
0 3 * * * cd /srv/myblog && MYSQL_PWD='<数据库口令>' BACKUP_DIR=/var/backups/myblog UPLOAD_DIR=/srv/myblog/uploads ./scripts/backup/backup.sh >> /var/log/myblog-backup.log 2>&1
```

环境变量说明：`BACKUP_DIR`、`RETENTION_DAYS`、`DB_HOST/DB_PORT/DB_USER/DB_NAME`、`UPLOAD_DIR`。
口令一律经 `MYSQL_PWD` 传递，避免出现在进程参数列表。

## 3. 恢复步骤

脚本：`scripts/backup/restore.sh`。

```bash
# 校验完整性并恢复到指定库（库不存在时自动创建）
DB_USER=root MYSQL_PWD='<数据库口令>' ./scripts/backup/restore.sh /var/backups/myblog/20260915-030000 blog

# 仅验证备份可用性时，恢复到演练库，不触碰生产库
DB_USER=root MYSQL_PWD='<数据库口令>' ./scripts/backup/restore.sh /var/backups/myblog/20260915-030000 blog_restore_drill
```

恢复 uploads 时通过 `UPLOAD_DIR` 指定目标目录，脚本会将归档解包到其父目录下。

## 4. 恢复演练（上线前必做一次，此后每季度一次）

1. 选取最近一次备份目录；
2. 恢复到演练库 `blog_restore_drill`（见第 3 节第二条命令）；
3. 比对源库与演练库：表数量一致、核心表（articles/users/comments）行数一致；
4. 抽查演练库中文章与评论内容可读；
5. 演练完成后删除演练库：`DROP DATABASE blog_restore_drill;`；
6. 将演练日期与结果登记到运维值班记录。

> 2026-09-15 已在开发库完成首次演练：备份 blog 库 23 张表，恢复至演练库后表数量与
> users 行数一致，随后清理演练库。生产环境部署后需按本节重新演练一次。

## 5. binlog 与 PITR（可选增强）

当前脚本提供每日全量备份。需要分钟级 RPO 时：

1. MySQL 服务端开启 `log_bin`（`server_id` 设为唯一值）；
2. 备份脚本扩展：在全量备份后记录当前 binlog 位点（`SHOW MASTER STATUS`），
   并将自上次备份以来的 binlog 文件复制到备份目录；
3. 恢复流程变为"先全量、再按位点回放 binlog"（`mysqlbinlog --start-position ... | mysql`）；
4. binlog 保留期建议 `binlog_expire_logs_seconds >= 3 * 86400`。

## 6. Windows 开发环境

脚本面向 Linux 服务器编写。Windows 本机验证可直接调用 `D:\mysql\bin` 下的
`mysqldump.exe` / `mysql.exe`，步骤与第 4 节相同。
