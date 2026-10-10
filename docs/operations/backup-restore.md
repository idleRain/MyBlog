# 备份与恢复操作手册

> 全仓此前零备份，单实例 MySQL 与本地媒体目录构成三重单点。
> 本手册给出容器化备份的调度方式、恢复步骤与演练流程。

## 1. 备份范围与频率

| 对象 | 方式 | 建议频率 | 保留期 |
|---|---|---|---|
| MySQL 数据库 | `mysqldump --single-transaction` 一致性快照，gzip 压缩 | 每日一次，建议安排凌晨低峰 | 14 天，可按需调整 |
| uploads 媒体文件 | tar.gz 全量打包 | 随每日备份执行 | 与数据库一致 |

增量与 PITR：如需恢复到任意时间点，需在 MySQL 服务端开启 `log_bin` 并将 binlog 随备份归档。
当前脚本仅做每日全量，RPO 为 24 小时，开启 binlog 后可缩短至分钟级，方案见第 7 节。
部署形态与可用性约束（单实例、无自动故障转移、故障即停站、恢复依赖人工介入与重启、无 RTO 承诺）
以 `docs/architecture-rules.md` 第 7 节为唯一口径，本手册不重复。

## 2. 备份的落盘位置与数据流

理解备份方案前需先明确三个存储位置的区别，这是容器化编排与传统宿主机部署最大的差异：

| 存储 | 类型 | 宿主机可见 | 存放内容 |
|---|---|---|---|
| `mysql_data` | Docker 命名卷 | 否 | MySQL 数据文件 |
| `uploads` | Docker 命名卷 | 否 | 用户上传的媒体文件 |
| `BACKUP_HOST_DIR` | 宿主机绑定挂载 | 是 | 备份产物，必须可见以便取出机器 |

关键约束：**备份必须落到宿主机可见的目录**。`mysql_data` 与 `uploads` 都是 Docker 命名卷，
其内容位于 Docker 的数据根目录内，若不显式绑定挂载到宿主机路径，备份产物将无法被复制到
异地存储，也就无法在宿主机或整机故障时发挥作用。

`BACKUP_HOST_DIR` 可在 `.env` 中设置，编排文件为该变量提供了默认值，未设置时按默认值落盘。
生产环境应显式设置该变量并指向独立于 Docker 数据目录的磁盘或挂载点。
该变量与 `BACKUP_RETENTION_DAYS` 均已登记在 `deploy/.env.example` 模板中，默认值分别为 `./backups` 与 14 天。

## 3. 为什么备份不作为常驻服务运行

Docker Compose **没有内置的定时调度能力**，这一点决定了 backup 服务的形态。可选方案与取舍：

| 方案 | 说明 | 取舍 |
|---|---|---|
| 常驻容器内运行 cron | 容器内启动 crond 每日触发 | 需额外维护进程监督、cron 日志与容器存活判断，且备份逻辑与调度耦合在一个容器内 |
| **一次性 job + 外部调度** | 服务声明 `profiles: [backup]`，由宿主机 cron 或外部调度器显式触发 | **本项目采用**。备份进程只在需要时存在，退出码即结果，无需常驻资源 |
| 独立 sidecar | 单独一套调度组件 | 对单实例个人博客而言运维成本过高 |

本项目采用**一次性 job 加外部调度**。原因有三：

1. 备份是**周期性批处理**而非长期驻留的服务，让一个每天工作数十秒的进程 24 小时占用资源并不合理；
2. 一次性进程的**退出码就是执行结果**，调度器可直接据此判定成败，无需额外实现健康检查；
3. 把调度交还给宿主机 cron 或外部调度器后，**告警链路可复用现有的值班机制**，不必在容器内重建一套。

`profiles: [backup]` 的作用是让 `docker compose up -d` **默认不启动** backup 服务，
避免它被当作常驻服务拉起；只有显式指定 profile 或使用 `docker compose run` 时才会创建。

## 4. 执行备份

### 4.1 手动执行一次备份

```bash
# 在仓库根目录执行，--rm 表示退出后自动移除容器
docker compose run --rm backup
```

`docker compose run` 会等待容器退出并把**容器的退出码作为命令的退出码**返回：

- `0` 表示备份成功，产物位于 `${BACKUP_HOST_DIR}/<时间戳>/`；
- 非 `0` 表示备份失败，具体原因见 stdout 与 stderr。

### 4.2 容器启动依赖的说明

`backup` 服务声明了 `depends_on: mysql: condition: service_healthy`，但需注意
**`docker compose run` 与 `docker compose up` 对 `depends_on` 的处理不同**：
`run` 默认**只启动目标服务本身**，不会等待也不会拉起所依赖的服务。

因此存在两种使用姿势，请按场景选择：

```bash
# 姿势一：MySQL 已在运行，直接执行备份即可
docker compose up -d
docker compose run --rm backup

# 姿势二：MySQL 尚未启动，需显式一并启动依赖服务
docker compose run --rm --dependencies backup
```

若在 MySQL 未运行时直接执行姿势一的命令，备份会因连接失败而返回非零退出码并打印错误。
这是刻意的设计：让失败显式暴露，而非静默产出一个空备份。

> **待实机确认**：`--dependencies` 选项的作用是让 `run` 一并启动目标服务的依赖，
> 该行为**尚未在真实 Docker 环境中验证**。若本机 Docker Compose 版本不支持该选项，
> 命令会提示未知参数；此时改用姿势一，先执行 `docker compose up -d` 再执行备份。
> 无论采用哪种姿势，`depends_on` 中的 `service_healthy` 条件都只对 `up` 生效，
> 备份脚本自身仍会对连接失败返回非零退出码，因此不会出现静默失败。

### 4.3 配置外部定时调度

宿主机 cron 配置示例，每日 03:00 执行备份并将日志留档：

```cron
0 3 * * * cd /srv/myblog && docker compose run --rm backup >> /var/log/myblog-backup.log 2>&1
```

要点说明：

- **必须 `cd` 到仓库根目录**，`docker compose` 依赖该目录下的编排文件与 `.env`；
- 日志重定向到文件后，可用 `grep` 检索失败记录，也可接入现有的日志采集；
- 输出了非零退出码即代表当日备份失败，可据此建立告警。

### 4.4 备份结果的监控

备份失败必须可被发现，本方案提供三个层次的信号：

1. **退出码**：`docker compose run` 返回非零即失败。这是最可靠的判定依据，
   可直接被 cron 的 `MAILTO`、systemd timer 的 `OnFailure` 或外部调度器的告警捕获。
2. **明确日志**：脚本失败时向 stderr 输出以 `错误:` 开头的完整技术陈述句，
   包含失败对象与判定依据；成功时输出 `备份完成: <目录>` 并列出产物体积。
3. **产物自检**：脚本在导出后校验归档非空，在校验清单写入前确认存在归档文件，
   避免「退出码为 0 但产物不可用」的静默失效。

若需接入主动告警，可将 cron 改为调用一个包装脚本，在 `docker compose run` 返回非零时
触发通知。本仓库未内置该包装脚本，因为通知渠道属于部署环境相关配置。

### 4.5 管道退出码的陷阱与修复

备份脚本原先使用如下形式导出数据库：

```bash
mysqldump ... "${DB_NAME}" | gzip > "${TARGET_DIR}/database.sql.gz"
```

备份脚本原先只声明 `set -e`，而 **shell 只取管道中最后一条命令的退出码**。因此当 `mysqldump` 因连接失败、
权限不足或库不存在而报错退出时，`gzip` 仍会正常读完（空的）输入并以 `0` 退出，
整个管道被判为成功。后果是脚本打印「备份完成」，实际产出一个仅含 gzip 头的**残缺归档**，
而这类故障通常直到需要恢复时才会暴露。

修复方式是对该管道显式判定并中止：

```bash
if ! mysqldump ... | gzip > "${TARGET_DIR}/database.sql.gz"; then
  echo "错误: 数据库 ${DB_NAME} 导出失败，mysqldump 或 gzip 返回非零退出码" >&2
  exit 1
fi
```

`restore.sh` 中的 `gzip -dc ... | mysql ...` 存在完全同型的缺陷：`mysql` 正常退出会掩盖
`gzip` 的解压失败，从而留下一个导入不完整的库。两处均已在同一变更中修复。
两个脚本现已在文件头全局声明 `set -euo pipefail`，并在导出与导入管道就近重复声明
`set -o pipefail` 作为意图标注；上述 `if !` 显式判定仍然保留，作为第一道拦截。

> **同类风险提示**：本仓库其他位置若存在「生产者 | 消费者」形式的管道且依赖 `set -e`
> 判定失败，都需按上述方式显式判定。仅在有 `set -o pipefail` 的前提下，管道才会
> 反映任一环节的非零退出码。

## 5. 恢复步骤

### 5.1 恢复数据库

```bash
# 校验完整性并恢复到指定库，库不存在时脚本会自动创建
docker compose run --rm backup /opt/myblog/scripts/backup/restore.sh /backups/20260915-030000 myblog
```

脚本在容器内的路径为 `/opt/myblog/scripts/backup/`，与编排中 `./scripts/backup` 的只读挂载点一致；
`restore.sh` 与 `backup.sh` 同在该目录下。

脚本会依次执行：检查归档存在且非空、按 `checksums.sha256` 核对完整性、创建目标库并导入。

### 5.2 恢复演练：使用演练库而非生产库

**验证备份可用性时必须恢复到演练库**，避免覆盖生产数据：

```bash
# 恢复到演练库 myblog_restore_drill，不触碰生产库
docker compose run --rm backup /opt/myblog/scripts/backup/restore.sh /backups/20260915-030000 myblog_restore_drill
```

### 5.3 恢复 uploads 媒体文件

uploads 归档的解包目标由 `UPLOAD_DIR` 决定，脚本会解包到该目录的父目录下。
容器内 `UPLOAD_DIR` 已指向 `uploads` 命名卷的挂载点，故恢复操作会直接写回该卷。

> **注意**：向已存在文件的 uploads 卷恢复会**覆盖同名文件**。生产环境执行前应先对现有
> 卷做一次快照，或先恢复到演练卷核对内容。

## 6. 恢复演练

### 6.1 历史演练记录

> **2026-09-15 开发库演练，非容器环境**：在开发机直接调用 MySQL 客户端完成。
> 备份 `blog` 库全部业务表（表清单见 `docs/database-architecture.md`，当前 30 张：用户 4、
> 内容 10、评论 2、互动 4、媒体 1、站点运营 2、字典 4、统计日志 3），
> 恢复至演练库后表数量与 `users` 行数一致，随后清理演练库。
> 该记录**仅覆盖脚本逻辑在传统环境下的正确性**，未覆盖容器化编排、命名卷挂载与
> 跨机器恢复，不能作为生产环境的演练结论。

### 6.2 待执行的生产环境容器化演练

> **状态：尚未执行。** 编制本手册的机器未安装 Docker，无法实机执行以下步骤。
> 以下步骤需在**具备 Docker 的环境**中执行，并须至少执行一次完整流程。
> 未执行前，本手册的容器化备份链路属于**未经运行验证的静态设计**。

演练目标：证明备份产物可用于恢复，且恢复后的站点能看到原有文章与图片。

#### 阶段一：在有数据的环境中产出备份

```bash
# 1) 确认四个常驻服务已启动且 mysql 与 server 为 healthy（backup 是一次性 job，不在常驻服务内）
docker compose up -d
docker compose ps

# 2) 记录恢复前的基线数据量，供阶段四比对
docker compose exec mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -N -B \
  -e "SELECT COUNT(*) FROM myblog.articles; SELECT COUNT(*) FROM myblog.users;"

# 3) 执行备份
docker compose run --rm backup
echo "备份退出码: $?"
```

期望结果：

- 退出码为 `0`；
- 日志中出现 `备份完成: /backups/<时间戳>`；
- 宿主机 `${BACKUP_HOST_DIR}/<时间戳>/` 下存在 `database.sql.gz`、`uploads.tar.gz`、
  `checksums.sha256` 与 `archives.list` 四个文件，且两个归档体积均大于 0。

```bash
# 4) 在宿主机核对产物
ls -lh "${BACKUP_HOST_DIR}"/*/
```

#### 阶段二：在干净环境中恢复

```bash
# 5) 另起一套干净的 Compose 环境，使用独立的项目名与数据卷，避免污染生产环境
BACKUP_HOST_DIR=/srv/myblog-drill-backups \
  docker compose -p myblog-drill up -d

# 6) 在干净环境中恢复数据库到独立库名
docker compose -p myblog-drill run --rm backup \
  /opt/myblog/scripts/backup/restore.sh /backups/<时间戳> myblog_restore_drill
```

期望结果：恢复脚本输出 `恢复完成: 数据库 myblog_restore_drill 与上传目录 uploads`，
退出码为 `0`；若任一环节失败，脚本会以非零退出码中止并说明失败对象。

#### 阶段三：比对数据一致性

```bash
# 7) 比对表数量
docker compose -p myblog-drill exec mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -N -B \
  -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='myblog_restore_drill';"

# 8) 比对核心表行数，与阶段一第 2 步的基线逐项核对
docker compose -p myblog-drill exec mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -N -B \
  -e "SELECT COUNT(*) FROM myblog_restore_drill.articles; SELECT COUNT(*) FROM myblog_restore_drill.users;"
```

期望结果：表数量与阶段一基线一致；`articles` 与 `users` 的行数与备份时刻的基线**逐项相等**。

```bash
# 9) 抽查内容可读，确认不是空表或乱码
docker compose -p myblog-drill exec mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" \
  -e "SELECT id, title FROM myblog_restore_drill.articles LIMIT 5;"
```

期望结果：输出真实文章标题，中文正常显示无乱码。

#### 阶段四：验证 uploads 与端到端可访问性

```bash
# 10) 核对 uploads 卷内文件数量与备份前一致
docker compose exec server sh -c 'find /app/uploads -type f | wc -l'
docker compose -p myblog-drill exec server sh -c 'find /app/uploads -type f | wc -l'

# 11) 端到端验证：经网关访问前台首页与一张真实图片
curl -I http://localhost/
curl -I http://localhost/uploads/<已知存在的一张图片路径>
```

期望结果：两条命令均返回 `200`；第 11 步证明恢复后的站点能看到原有文章与图片。

#### 阶段五：清理与留档

```bash
# 12) 删除演练库与演练环境，注意 -v 会删除演练卷
docker compose -p myblog-drill exec mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" \
  -e "DROP DATABASE myblog_restore_drill;"
docker compose -p myblog-drill down -v
```

期望结果：演练库与环境被清理，生产环境不受影响。

**留档要求**：演练完成后须将**执行日期、总耗时、实际遇到的问题与处理方式**追加到本节，
并把本节状态从「尚未执行」改为实测结论。在此之前，请勿将本节的期望结果表述为已验证。

## 7. binlog 与 PITR（可选增强）

当前脚本提供每日全量备份。需要分钟级 RPO 时：

1. MySQL 服务端开启 `log_bin`，并将 `server_id` 设为唯一值；
2. 备份脚本扩展：在全量备份后记录当前 binlog 位点，即执行 `SHOW MASTER STATUS`，
   并将自上次备份以来的 binlog 文件复制到备份目录；
3. 恢复流程变为先全量、再按位点回放 binlog，回放命令为
   `mysqlbinlog --start-position ... | mysql`；
4. binlog 保留期建议满足 `binlog_expire_logs_seconds >= 3 * 86400`。

## 8. 脚本环境变量契约

`backup.sh` 与 `restore.sh` 的环境变量是编排与脚本之间的接口，改动任一侧都需同步另一侧。

| 变量 | 脚本默认值 | 用途 |
|---|---|---|
| `BACKUP_DIR` | `/var/backups/myblog` | 备份输出根目录，每次备份创建以时间戳命名的子目录 |
| `RETENTION_DAYS` | `14` | 备份保留天数，超期目录自动清理 |
| `DB_HOST` / `DB_PORT` | `127.0.0.1` / `3306` | 数据库连接地址 |
| `DB_USER` / `DB_NAME` | `root` / `blog` | 数据库账户与库名；脚本默认库名为 `blog`，经编排运行时由 `${MYSQL_DATABASE:-myblog}` 覆盖为 `myblog` |
| `UPLOAD_DIR` | `uploads` | 上传目录，须与 `media.upload_dir` 一致 |
| `MYSQL_PWD` | 无 | 数据库口令，经环境变量传递以避免出现在进程参数列表中 |

`restore.sh` 的调用形式为 `restore.sh <备份目录> [目标数据库名]`，
目标库名缺省时回落 `DB_NAME`，再缺省为 `blog`。

### 8.1 编排层向脚本注入的取值

编排文件中的 `backup` 服务会覆盖上表的脚本默认值，使容器内无需额外配置即可正确运行：

| 脚本变量 | 编排内取值 | 说明 |
|---|---|---|
| `BACKUP_DIR` | `/backups` | 指向绑定挂载的宿主机目录，产物因此可在宿主机直接取用 |
| `RETENTION_DAYS` | `${BACKUP_RETENTION_DAYS:-14}` | 由 `.env` 的 `BACKUP_RETENTION_DAYS` 控制，未设置时取 14 天 |
| `DB_HOST` | `mysql` | 取 Compose 服务名，经服务发现访问数据库容器 |
| `DB_USER` / `DB_NAME` | `root` / `${MYSQL_DATABASE:-myblog}` | 与数据库服务的账户及库名保持一致 |
| `UPLOAD_DIR` | `/app/uploads` | 指向 uploads 命名卷在备份容器内的挂载点 |
| `MYSQL_PWD` | `${MYSQL_ROOT_PASSWORD}` | 复用数据库口令变量，避免第二处口令来源 |

> **口令与账户来源**：`MYSQL_PWD` 取自编排文件中数据库服务已声明的 `${MYSQL_ROOT_PASSWORD}`；
> `DB_USER` 为该口令对应的 root 账户，`DB_NAME` 取自 `${MYSQL_DATABASE:-myblog}`。
> 后端容器经 MYBLOG_DATABASE_* 前缀的环境变量连接同一库与同一账户（见 `docker-compose.yml` 的 server 服务），
> 备份侧因此不引入第二处口令来源。

> **注意**：`RETENTION_DAYS` 与 `BACKUP_RETENTION_DAYS` 是两个不同层级的变量。
> 前者是脚本读取的变量，后者是 `.env` 中供编排取值使用的变量，编排负责把后者注入为前者。
> 设置保留期应修改 `.env` 中的 `BACKUP_RETENTION_DAYS`，而非直接设置 `RETENTION_DAYS`。

## 9. 脚本依赖的外部命令

两个脚本依赖以下外部命令，执行备份的容器镜像必须全部提供：

| 命令 | 提供方 |
|---|---|
| `mysqldump`、`mysql` | MySQL 客户端 |
| `sha256sum`、`date`、`mkdir`、`rm`、`ls`、`cat`、`basename`、`dirname`、`sort` | GNU coreutils |
| `find` | GNU findutils |
| `tar` | GNU tar |
| `gzip` | gzip |
| `bash` | 脚本解释器 |

> **待实机确认**：执行备份的容器基础镜像是否自带上述全部命令**尚未在真实镜像中验证**。
> 若镜像缺少 `gzip`、`tar`、`sha256sum` 或 GNU `find`，脚本会在相应步骤以非零退出码失败。
> 核对方式见本仓库 `scripts/backup/check-dependencies.mjs`，该工具可静态列出脚本的全部命令依赖。

### 9.1 镜像命令可用性的一次性判定

在**具备 Docker 的环境**中执行以下命令，可一次判定镜像是否满足脚本的全部依赖：

```bash
docker run --rm mysql:8.0 sh -c 'for c in bash mysqldump mysql gzip tar sha256sum find sort date basename dirname cat ls rm; do printf "%-12s" "$c"; command -v "$c" || echo MISSING; done'
```

期望结果：每一行都输出该命令的绝对路径，**不应出现任何 `MISSING`**。

### 9.2 缺少命令时的处理方案

若上述命令的输出包含 `MISSING`，说明 `mysql:8.0` 镜像不满足脚本依赖，
此时需改用**自定义镜像**：新增 `deploy/Dockerfile.backup`，基于 `mysql:8.0` 补装缺失的软件包，
并将编排中 backup 服务的 `image: mysql:8.0` 改为 `build` 指向该 Dockerfile。

补装内容取决于实际缺失项，通常为 `gzip`、`tar`、`findutils` 与 `coreutils`。
以 Oracle Linux 系基础镜像为例，安装命令形如：

```dockerfile
FROM mysql:8.0
RUN microdnf install -y gzip tar findutils coreutils && microdnf clean all
```

> **关于基础发行版的说明**：官方 `mysql:8.0` 镜像据其发行版信息推断基于 Oracle Linux，
> 该发行版的官方 MySQL 镜像为精简镜像，历史上存在缺少 `gunzip` 等工具的用户报告。
> 此项判断**据镜像发行版推断，未经本项目实机确认**，外部参考见
> [docker-library/mysql 议题 826](https://github.com/docker-library/mysql/issues/826)。
> 请以 9.1 节命令的实际输出为准，不要以本节推断作为结论。

## 10. Windows 开发环境

脚本面向 Linux 容器编写，依赖 GNU 工具集，**不能在 Windows 的原生 cmd 或 PowerShell 中直接运行**。
在 Windows 上验证脚本逻辑需经 WSL 或 Git Bash 提供 GNU 环境。若仅需手工导出一份数据，
可直接调用本机 MySQL 安装目录下的 `mysqldump.exe` 与 `mysql.exe`，步骤与第 5 节相同。
