# scripts/ 脚本说明

本目录集中存放 MyBlog 的跨项目构建与开发脚本，统一由 **Node.js + tsx** 执行。
根目录 `package.json` 中的 scripts 大多是对这些脚本的**薄封装**，实际命令即为
`tsx scripts/<脚本名>.ts [参数]`。

## 参数透传

pnpm 会把 `pnpm run <script>` 中脚本名之后的参数**原样透传**给脚本，因此多数脚本
无需单独封装即可按需传参：

```bash
pnpm run dev --web                          # 等价于 tsx scripts/dev.ts --web
pnpm run seed:admin --username root         # 自定义种子账户参数
```

为规避与 pnpm 自身选项（如 `--filter`、`--parallel`）同名冲突，推荐用 `--` 显式分隔：

```bash
pnpm run migrate -- goto 5
```

## 脚本一览

| 脚本                       | 用途                                                                                                  | 参数                                                                                                         | pnpm 入口                                                                                                    |
| -------------------------- | ----------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------ |
| `dev.ts`                   | 智能启动开发服务，含环境检查、端口检查、就绪监控，Ctrl+C 全部停止；后端热更新基于 air，缺失时自动安装 | `--server` / `--web` / `--admin`，缺省则全部启动                                                             | `pnpm run dev` / `dev:server` / `dev:web` / `dev:admin`                                                      |
| `build.ts`                 | 统一构建后端与两个前端应用                                                                            | `--clean` / `--production` / `--server-only` / `--web-only` / `--skip-tests` / `--skip-lint`                 | `pnpm run build` 及 `build:clean` / `build:production` / `build:server` / `build:web` / `build:fast`         |
| `setup.ts`                 | 一键环境设置，含系统要求检查、依赖安装、环境文件生成                                                  | 无                                                                                                           | `pnpm run setup`                                                                                             |
| `clean-web.ts`             | 跨平台清理前端各应用构建产物（`.svelte-kit` / `build` / `dist`）                                      | 无                                                                                                           | `pnpm run clean:web`                                                                                         |
| `go-tools.ts`              | Go 侧构建、测试、lint、格式化、清理的统一入口                                                         | `build` / `test` / `deps` / `lint-install` / `lint` / `format` / `fmt` / `vet` / `clean` / `quality`         | `pnpm run test:server` / `lint:server` / `format:server` / `clean:server` / `go:lint-install` / `go:quality` |
| `go-toolchain.ts`          | golangci-lint 与 goimports 的检测、安装与路径解析，供其他脚本复用                                     | `ensure` / `golangci` / `goimports`                                                                          | 经 `go-tools.ts` 间接使用                                                                                    |
| `lint-staged-goimports.ts` | lint-staged 专用 goimports 执行入口                                                                   | 由 lint-staged 传入文件列表                                                                                  | 由 `.husky/pre-commit` 触发                                                                                  |
| `migrate.ts`               | 基于 golang-migrate 的数据库迁移管理                                                                  | `create <name>` / `up` / `down [steps]` / `goto <version>` / `force <version>` / `version` / `drop` / `help` | `pnpm run migrate` 及 `migrate:create` / `migrate:up` / `migrate:down` / `migrate:version`                   |
| `seed.ts`                  | 初始化或提升超级管理员账户，命令幂等                                                                  | `--username` / `--password` / `--email` / `--help`                                                           | `pnpm run seed:admin`                                                                                        |
| `lib/is-main.ts`           | 判断当前模块是否为直接运行的入口，供各脚本复用                                                        | —                                                                                                            | 内部工具                                                                                                     |

## 备份与部署校验脚本

`scripts/backup/` 下为容器化备份链路的脚本与静态校验工具，均为纯 Node 执行，无 tsx 依赖。

| 脚本                              | 用途                                                        | 参数 | pnpm 入口 |
| --------------------------------- | ----------------------------------------------------------- | ---- | --------- |
| `backup/backup.sh`                | 容器内执行数据库导出与 uploads 打包，写入校验清单并清理过期 | 经环境变量配置，见 `docs/operations/backup-restore.md` | 无，由编排的 backup 服务调用 |
| `backup/restore.sh`               | 容器内执行备份还原，位置参数为备份目录与可选目标库名         | `<备份目录> [目标库名]` | 无，手工执行 |
| `backup/check-dependencies.mjs`   | 静态列出 `backup.sh`/`restore.sh` 的外部命令依赖，供容器镜像能力核对 | 无 | 无，直接 `node` 执行 |
| `backup/verify-static.mjs`        | 备份与编排交付物的静态校验套件，共 10 项，失败时退出码非零   | 无 | 无，直接 `node` 执行 |

## 脚本设计约定

- 脚本统一使用 `#!/usr/bin/env -S node --import tsx` shebang，可直接执行。
- 直接运行的脚本通过 `isMainModule(import.meta.url)` 判断入口，被其他模块 `import`
  时不触发副作用。
- 面向 Go 的工具链统一经 `go-tools.ts` / `go-toolchain.ts` 封装，确保跨平台，
  Windows 下的 `rmdir`、`.exe` 后缀等差异已在内部处理。
- 数据库相关操作（迁移、种子）依赖本机 `migrate` 命令行工具与 MySQL 服务，
  迁移前需确认 `server/migrations` 目录存在。
