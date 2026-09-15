#!/usr/bin/env bash
# MyBlog 备份恢复脚本：将备份目录中的数据库与上传文件恢复到指定位置。
# 用法示例:
#   DB_USER=root ./restore.sh /var/backups/myblog/20260915-030000 [目标数据库名]
# 目标数据库名缺省取 DB_NAME 环境变量，再缺省为 blog；可传演练库名做恢复验证。
set -euo pipefail

BACKUP_PATH="${1:?用法: restore.sh <备份目录> [目标数据库名]}"
DB_NAME="${2:-${DB_NAME:-blog}}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
# 上传文件的恢复目标目录，与服务运行目录下的 media.upload_dir 一致
UPLOAD_DIR="${UPLOAD_DIR:-uploads}"

if [ ! -f "${BACKUP_PATH}/database.sql.gz" ]; then
  echo "错误: ${BACKUP_PATH} 中未找到 database.sql.gz" >&2
  exit 1
fi

# 恢复前核对备份完整性，校验失败立即中止
if [ -f "${BACKUP_PATH}/checksums.sha256" ]; then
  ( cd "${BACKUP_PATH}" && sha256sum -c checksums.sha256 --quiet )
fi

# 恢复数据库到目标库，目标库需已存在
mysql --host="${DB_HOST}" --port="${DB_PORT}" --user="${DB_USER}" \
  --default-character-set=utf8mb4 \
  -e "CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
gzip -dc "${BACKUP_PATH}/database.sql.gz" | \
  mysql --host="${DB_HOST}" --port="${DB_PORT}" --user="${DB_USER}" \
    --default-character-set=utf8mb4 "${DB_NAME}"

# 恢复上传文件到目标目录
if [ -f "${BACKUP_PATH}/uploads.tar.gz" ]; then
  tar -xzf "${BACKUP_PATH}/uploads.tar.gz" -C "${UPLOAD_DIR}/.."
fi

echo "恢复完成: 数据库 ${DB_NAME} 与上传目录 ${UPLOAD_DIR}"
