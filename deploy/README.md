# MyBlog Docker 部署手册

> 本机已验证镜像构建逻辑与产物结构，真实环境镜像构建与编排运行待部署环境验证。

## 架构

| 容器 | 镜像 | 职责 |
|---|---|---|
| `nginx` | `deploy/Dockerfile.nginx` | 网关：托管 admin 静态、反代前台 SSR、后端 API 与 /uploads，对外暴露 `${HTTP_PORT:-80}` |
| `web` | `deploy/Dockerfile.web` | 前台 SvelteKit SSR，adapter-node 产物，`node build` 运行，仅容器网络内暴露 3000 |
| `server` | `deploy/Dockerfile.server` | Go 后端，release 模式启动时执行 golang-migrate 迁移，仅容器网络内暴露 3000 |
| `mysql` | `mysql:8.0` | 业务数据库，数据与上传文件分别落命名卷 |
| `backup` | `mysql:8.0` | 一次性备份任务，归 `backup` profile，`docker compose up` 默认不拉起 |

访问路径：`http://<主机>/` 前台，`http://<主机>/admin/` 后台，`/api/*` 与 `/uploads/*` 由 nginx 反代到后端。

### 部署形态：单实例

- 编排中无副本声明与故障转移（`git grep -n "replicas" -- docker-compose.yml` 应为空），应用进程、MySQL、网关各一份，**不支持按多副本假设做设计与承诺**。
- 会话令牌表与请求限流计数均为进程内存结构，`server` 容器重启或重新部署将导致**全体用户登出**，发布窗口需提前预告；任一组件故障即停站，恢复依赖人工介入，无 RTO 承诺，RPO 为 24 小时（`docs/operations/backup-restore.md`）。
- 扩容前置改造三项见 `docs/architecture-rules.md` §7.2。

### 健康探针

| 服务 | 探针命令 | 判定 |
|---|---|---|
| `server` | `wget -q -O /dev/null http://localhost:3000/api/health/ready` | 数据库连通且上传目录可写，否则返回 503 |
| `web` | `wget -q -O /dev/null http://localhost:3000/` | 前台首页可渲染 |
| `nginx` | `wget -q -O /dev/null http://localhost/` | 网关静态托管与反代链路可用 |
| `mysql` | `mysqladmin ping` | 数据库进程可达 |

启动顺序由 `depends_on` 约束：`server` 等 MySQL 健康，`nginx` 等 `server` 健康且 `web` 已启动。

### 环境变量注入

| 变量 | 必需 | 作用 | 默认值 |
|---|---|---|---|
| `MYSQL_ROOT_PASSWORD` | 是 | MySQL root 口令，同时注入 `server` 的 `MYBLOG_DATABASE_PASSWORD` 与备份任务的 `MYSQL_PWD` | 无，缺失时 compose 拒绝启动 |
| `MYSQL_DATABASE` | 否 | 业务库名，注入 `server` 的 `MYBLOG_DATABASE_DBNAME` | `myblog` |
| `HTTP_PORT` | 否 | nginx 对外暴露端口 | `80` |
| `BACKUP_HOST_DIR` | 否 | 备份产物在宿主机上的落盘目录（绑定挂载） | `./backups` |
| `BACKUP_RETENTION_DAYS` | 否 | 备份保留天数，超期目录被清理 | `14` |

`server` 容器的其余参数经 `MYBLOG_*` 前缀环境变量注入：`MYBLOG_SERVER_MODE=release`、
`MYBLOG_SERVER_HOST=0.0.0.0`、`MYBLOG_DATABASE_HOST=mysql`、`MYBLOG_DATABASE_PORT=3306`、
`MYBLOG_DATABASE_USERNAME=root`。变量名由 `MYBLOG_` 加配置键的大写下划线形式构成，
列表与映射类配置无法经环境变量表达，见 `server/internal/config/config.go`。

### 数据卷

| 卷 | 挂载点 | 说明 |
|---|---|---|
| `mysql_data`（命名卷） | `mysql` 的 `/var/lib/mysql` | 数据库数据 |
| `uploads`（命名卷） | `server` 的 `/app/uploads` | 上传目录，与后端 `media.upload_dir=uploads` 对应，经 `/uploads/*` 反代访问 |
| `uploads`（只读） | `backup` 的 `/app/uploads:ro` | 备份打包媒体文件，只读挂载避免改动原文件 |
| `${BACKUP_HOST_DIR:-./backups}`（绑定挂载） | `backup` 的 `/backups` | 备份产物落到宿主机可见目录 |

## 前置条件

1. 安装 Docker 与 Docker Compose 插件。
2. 无需人工补齐任何配置文件。镜像构建改用入库的无密钥模板
   `server/configs/config.example.yaml`，真实配置一律经环境变量在运行时注入。
   本地开发仍使用 `server/configs/config.yaml`，该文件已登记在 `.gitignore`，镜像构建不读取它。
3. 本地镜像构建会执行前端生产构建与 Go 编译，耗时约 5 至 10 分钟。

## 部署步骤

```bash
# 1) 创建环境变量文件并填写数据库强口令
cp deploy/.env.example .env

# 2) 编辑 .env：MYSQL_ROOT_PASSWORD 必须替换为随机强值

# 3) 构建并启动（backup 属 backup profile，不随本次启动拉起）
docker compose up -d --build

# 4) 验证
curl http://localhost/api/health/ready     # 期望 HTTP 200 与 status=ready
curl -I http://localhost/                  # 前台 SSR 首页
curl -I http://localhost/admin/            # 后台 SPA 入口
```

## 运维要点

- 后端迁移：server 容器以 `MYBLOG_SERVER_MODE=release` 启动，启动时对镜像内 `/app/migrations`
  执行 golang-migrate 的 `up`；后续 schema 变更按迁移双轨手写增量迁移后重新部署。
- 数据卷：`server` 的 `/app/uploads` 与 `mysql` 数据目录均为命名卷，`docker compose down` 不会删除；
  彻底清理需 `docker compose down -v`。
- 升级：`git pull` 后 `docker compose up -d --build` 重建受影响镜像；重启导致全体用户登出，需提前预告。
- 备份：经 `docker compose run --rm backup` 显式触发，退出码即结果，调度交回宿主机 cron；
  命令细节与恢复流程见 `docs/operations/backup-restore.md`。

## 已知边界

- 本机未安装 Docker，`docker compose config` 与镜像构建未在本机执行，配置经静态审查。
- 后端镜像内烧录的是无密钥模板 `config.example.yaml`，其中口令与库名均为占位值，
  真实值经 `MYBLOG_DATABASE_*` 环境变量在运行时注入，镜像层中不含真实口令。
- 前端运行阶段依赖经 `prod-deps` 阶段以扁平布局安装后随镜像分发。
  该布局已在无 Docker 环境下以 Node 直接运行 `build` 完成验证，
  镜像构建本身仍待部署环境确认。
- TLS 终止未包含在本编排内，对外暴露建议置于前置反代或负载均衡之后。
