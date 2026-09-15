#!/usr/bin/env bash
# MyBlog 数据库与上传文件备份脚本，供 cron 或编排器每日调度。
# 用法示例:
#   BACKUP_DIR=/var/backups/myblog DB_USER=root ./backup.sh
# 数据库口令经环境变量 MYSQL_PWD 传递，避免出现在进程参数列表中。
set -euo pipefail

# 备份保留天数，超过后自动清理当日目录
RETENTION_DAYS="${RETENTION_DAYS:-14}"
# 备份输出根目录，每日一个以时间戳命名的子目录
BACKUP_DIR="${BACKUP_DIR:-/var/backups/myblog}"
# 数据库连接参数
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_NAME="${DB_NAME:-blog}"
# 上传文件目录，与服务运行目录下的 media.upload_dir 一致
UPLOAD_DIR="${UPLOAD_DIR:-uploads}"

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
TARGET_DIR="${BACKUP_DIR}/${TIMESTAMP}"
mkdir -p "${TARGET_DIR}"

# 导出数据库：单事务保证一致性快照，包含触发器、存储过程与事件。
# 未使用 --databases 选项，恢复时可自由选择目标库名。
mysqldump --host="${DB_HOST}" --port="${DB_PORT}" --user="${DB_USER}" \
  --single-transaction --routines --triggers --events \
  --default-character-set=utf8mb4 \
  "${DB_NAME}" | gzip > "${TARGET_DIR}/database.sql.gz"

# 打包上传文件目录，目录不存在时跳过并给出提示
if [ -d "${UPLOAD_DIR}" ]; then
  tar -czf "${TARGET_DIR}/uploads.tar.gz" -C "$(dirname "${UPLOAD_DIR}")" "$(basename "${UPLOAD_DIR}")"
else
  echo "警告: 上传目录 ${UPLOAD_DIR} 不存在，本次备份未包含 uploads" >&2
fi

# 写入校验清单，恢复前用于核对备份完整性
( cd "${TARGET_DIR}" && sha256sum ./* > checksums.sha256 )

# 清理超过保留期的过期备份
find "${BACKUP_DIR}" -mindepth 1 -maxdepth 1 -type d -mtime "+${RETENTION_DAYS}" -exec rm -rf {} +

echo "备份完成: ${TARGET_DIR}"
