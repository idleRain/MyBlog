# 开发环境配置指南

## 热更新开发环境

为了提升开发效率，项目集成了Air热更新工具，让你在修改代码后自动重新编译和重启应用。

### 快速开始

#### 方式1: 使用Makefile (推荐)

```bash
# 启动热更新开发环境
make dev

# 查看所有可用命令
make help
```

#### 方式2: 使用脚本

**Linux/Mac:**

```bash
./scripts/dev.sh
```

**Windows:**

```bash
scripts\dev.bat
```

#### 方式3: 直接使用Air

```bash
# 首次安装Air
go install github.com/cosmtrek/air@latest

# 启动热更新
air -c .air.toml
```

### Air配置说明

项目包含两个Air配置文件：

#### `.air.toml` (完整配置)

- 详细的配置选项
- 完整的文件监听和排除规则
- 适合生产级开发环境

#### `.air.simple.toml` (简化配置)

- 精简的配置选项
- 基本的热更新功能
- 适合快速开发和测试

### 监听规则

**监听的文件类型:**

- `.go` - Go源代码文件
- `.yaml`, `.yml` - 配置文件
- `.json` - JSON配置文件

**忽略的目录:**

- `tmp/` - 临时构建文件
- `vendor/` - 依赖包目录
- `logs/` - 日志文件目录
- `.git/` - Git版本控制目录
- `.vscode/`, `.idea/` - IDE配置目录

**忽略的文件:**

- `*_test.go` - 测试文件
- `*.md` - 文档文件
- `*.txt` - 文本文件

### 开发工作流程

1. **启动热更新环境**
   ```bash
   make dev
   ```

2. **修改代码**
  - 编辑任何Go源文件
  - 修改配置文件 (yaml/json)
  - 保存文件

3. **自动重新编译**
  - Air检测到文件变化
  - 自动重新编译项目
  - 重启应用服务器

4. **测试API**
   ```bash
   # 健康检查
   curl -X POST http://localhost:3000/api/health
   ```

### 故障排除

#### Air未安装

```bash
go install github.com/cosmtrek/air@latest
```

#### 权限问题 (Linux/Mac)

```bash
chmod +x scripts/dev.sh
```

#### 端口占用

- 检查端口3000是否被占用
- 修改 `configs/config.yaml` 中的端口配置

#### 编译失败

- 检查Go代码语法错误
- 运行 `go mod tidy` 更新依赖
- 查看Air输出的错误信息

### 性能优化

#### 减少重新编译时间

- 使用Go模块缓存：`go env GOCACHE`
- 优化导入语句
- 避免不必要的依赖

#### 自定义监听规则

编辑 `.air.toml` 文件：

```toml
[build]
  # 自定义包含的文件扩展名
  include_ext = ["go", "yaml", "json"]
  
  # 自定义排除的目录
  exclude_dir = ["tmp", "vendor", "logs"]
  
  # 调整重新构建延迟 (毫秒)
  rerun_delay = 500
```

### 调试支持

#### 使用Delve调试器

1. 安装Delve：
   ```bash
   go install github.com/go-delve/delve/cmd/dlv@latest
   ```

2. 修改Air配置启用调试：
   ```toml
   [build]
     cmd = "go build -gcflags='-N -l' -o ./tmp/main ./cmd/myblog"
     full_bin = "dlv exec ./tmp/main --listen=:2345 --headless=true --api-version=2 --accept-multiclient"
   ```

3. 使用IDE连接调试器 (端口2345)

### 日志管理

#### 查看实时日志

热更新模式下，应用日志会直接显示在终端中。

#### 日志级别调整

修改 `configs/config.yaml`：

```yaml
logger:
  level: "debug"  # debug, info, warn, error
```

### 环境变量

#### 常用环境变量

```bash
# 设置Go代理 (国内用户)
export GOPROXY=https://goproxy.cn,direct

# 设置Go模块模式
export GO111MODULE=on

# 设置构建缓存目录
export GOCACHE=/path/to/cache
```

### 团队开发

#### .gitignore 配置

项目已配置忽略以下文件：

```
tmp/
bin/
logs/
*.log
.env
.DS_Store
```

#### 依赖管理

```bash
# 更新依赖
go mod tidy

# 查看依赖
go list -m all

# 清理模块缓存
go clean -modcache
```

### 最佳实践

1. **使用热更新进行开发**：日常开发使用 `make dev`
2. **定期测试生产构建**：使用 `make build && make run` 测试
3. **代码格式化**：保存前运行 `make fmt`
4. **代码检查**：提交前运行 `make check`
5. **清理临时文件**：定期运行 `make clean`

### 更多资源

- [Air官方文档](https://github.com/cosmtrek/air)
- [Go官方文档](https://golang.org/doc/)
- [Gin框架文档](https://gin-gonic.com/docs/)
