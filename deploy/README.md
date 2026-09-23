# MyBlog Docker 部署手册

> 关联体检项：OPS-01（部署载体）/ OPS-03（adapter 选型）。
> 本机已验证镜像构建逻辑与产物结构，真实环境镜像构建与编排运行待部署环境验证。

## 架构

| 容器 | 镜像 | 职责 |
|---|---|---|
| `nginx` | `deploy/Dockerfile.nginx` | 网关：托管 admin 静态、反代前台 SSR、后端 API 与 /uploads |
| `web` | `deploy/Dockerfile.web` | 前台 SvelteKit SSR，adapter-node 产物，`node build` 运行 |
| `server` | `deploy/Dockerfile.server` | Go 后端，release 模式启动时执行 golang-migrate 迁移 |
| `mysql` | `mysql:8.0` | 业务数据库，数据与上传文件分别落命名卷 |

访问路径：`http://<主机>/` 前台，`http://<主机>/admin/` 后台，`/api/*` 与 `/uploads/*` 由 nginx 反代到后端。

## 前置条件

1. 安装 Docker 与 Docker Compose 插件。
2. `server/configs/config.yaml` 必须存在于工作区。该文件已脱离 git 追踪（OPS-08），
   新克隆需按 OPS-08 人工操作清单重建或从部署机同步。
3. 本地镜像构建会执行前端生产构建与 Go 编译，耗时约 5 至 10 分钟。

## 部署步骤

```bash
# 1) 创建环境变量文件并填写数据库强口令
cp deploy/.env.example .env

# 2) 编辑 .env：MYSQL_ROOT_PASSWORD 必须替换为随机强值

# 3) 构建并启动
docker compose up -d --build

# 4) 验证
curl http://localhost/api/health/ready     # 期望 200 与 status=ready
curl -I http://localhost/                  # 前台 SSR 首页
curl -I http://localhost/admin/            # 后台 SPA 入口
```

## 运维要点

- 后端迁移：server 容器以 release 模式启动，首次启动自动执行 `server/migrations` 基线；
  后续 schema 变更按迁移双轨手写增量迁移后重新部署。
- 数据卷：`mysql_data` 与 `uploads` 为命名卷，`docker compose down` 不会删除；
  彻底清理需 `docker compose down -v`。
- 升级：`git pull` 后 `docker compose up -d --build` 重建受影响镜像。
- 备份：数据库与 uploads 的每日备份沿用 `scripts/backup/`，手册见 `docs/operations/backup-restore.md`。

## 已知边界

- 本机未安装 Docker，`docker compose config` 与镜像构建未在本机执行，配置经静态审查。
- 镜像内 config.yaml 使用环境变量覆盖数据库口令，其余配置项沿用仓库默认。
- TLS 终止未包含在本编排内，对外暴露建议置于前置反代或负载均衡之后。
