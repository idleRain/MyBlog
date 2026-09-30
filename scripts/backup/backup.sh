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

# 导出数据库：单事务确保一致性快照，包含触发器、存储过程与事件。
# 未使用 --databases 选项，恢复时可自由选择目标库名。
# 本次导出的失败必须以显式分支判定并给出可读日志，避免 gzip 正常退出掩盖 mysqldump 的失败。
# 此处重复声明 pipefail 的作用是让该意图在管道就近可见，不依赖回溯文件头；
# 文件头部已全局声明 pipefail，因此该行是幂等的意图标注，不构成额外防护。
set -o pipefail
if ! mysqldump --host="${DB_HOST}" --port="${DB_PORT}" --user="${DB_USER}" \
  --single-transaction --routines --triggers --events \
  --default-character-set=utf8mb4 \
  "${DB_NAME}" | gzip > "${TARGET_DIR}/database.sql.gz"; then
  echo "错误: 数据库 ${DB_NAME} 导出失败，mysqldump 或 gzip 返回非零退出码" >&2
  exit 1
fi

# 导出后校验归档非空：数据库存在但表数量为零时 mysqldump 仍返回 0 并产出极小的归档，
# 这类归档恢复后无法得到可用数据，属于静默失效，故在此显式拦截。
if [ ! -s "${TARGET_DIR}/database.sql.gz" ]; then
  echo "错误: ${TARGET_DIR}/database.sql.gz 为空，备份内容不可用" >&2
  exit 1
fi

# 打包上传文件目录，目录不存在时跳过并给出提示
if [ -d "${UPLOAD_DIR}" ]; then
  if ! tar -czf "${TARGET_DIR}/uploads.tar.gz" \
    -C "$(dirname "${UPLOAD_DIR}")" "$(basename "${UPLOAD_DIR}")"; then
    echo "错误: 上传目录 ${UPLOAD_DIR} 打包失败，tar 返回非零退出码" >&2
    exit 1
  fi
else
  echo "警告: 上传目录 ${UPLOAD_DIR} 不存在，本次备份未包含 uploads" >&2
fi

# 写入校验清单，恢复前用于核对备份完整性。
# 先枚举归档文件再写入清单，避免 sha256sum 的输出文件自身被纳入校验范围，
# 否则恢复时会产生一条永远无法通过的校验项。
CHECKSUM_FILE="checksums.sha256"
ARCHIVE_LIST="archives.list"
# 使用 -printf 之外的写法以兼容 GNU 与 BusyBox 两套 find 实现，
# 只取文件名部分，避免清单中出现绝对路径而使备份目录无法整体迁移。
find "${TARGET_DIR}" -mindepth 1 -maxdepth 1 -type f ! -name "${CHECKSUM_FILE}" \
  ! -name "${ARCHIVE_LIST}" -exec basename {} \; | sort > "${TARGET_DIR}/${ARCHIVE_LIST}"
if [ ! -s "${TARGET_DIR}/${ARCHIVE_LIST}" ]; then
  echo "错误: ${TARGET_DIR} 中没有任何归档文件，备份未产出可用内容" >&2
  exit 1
fi
# 显式判定校验清单的生成结果，使该步骤的失败原因不依赖 pipefail 的间接传递。
# 归档文件名均以受控的时间戳与固定后缀构成，不含前导连字符，故无需 -- 分隔符。
if ! ( cd "${TARGET_DIR}" && sha256sum $(cat "${ARCHIVE_LIST}") > "${CHECKSUM_FILE}" ); then
  echo "错误: ${TARGET_DIR}/${CHECKSUM_FILE} 生成失败，备份完整性无法在恢复前核对" >&2
  exit 1
fi

# 清理超过保留期的过期备份
find "${BACKUP_DIR}" -mindepth 1 -maxdepth 1 -type d -mtime "+${RETENTION_DAYS}" -exec rm -rf {} +

# 汇总输出本次产出的归档与体积，便于调度器与值班人员从日志直接核对备份结果
echo "备份完成: ${TARGET_DIR}"
( cd "${TARGET_DIR}" && ls -lh $(cat "${ARCHIVE_LIST}") "${CHECKSUM_FILE}" )
