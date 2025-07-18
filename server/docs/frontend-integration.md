# 前端项目整合指南

本文档说明如何将 Vite + SvelteKit 前端项目与当前的 Go 后端项目整合。

## 项目结构设计

### 推荐的 Monorepo 结构

```
MyBlog/
├── backend/                    # Go 后端项目
│   ├── cmd/
│   ├── internal/
│   ├── pkg/
│   ├── configs/
│   ├── docs/
│   ├── scripts/
│   ├── go.mod
│   └── go.sum
├── frontend/                   # SvelteKit 前端项目
│   ├── src/
│   ├── static/
│   ├── package.json
│   ├── vite.config.js
│   └── svelte.config.js
├── docs/                       # 共享文档
├── scripts/                    # 统一启动脚本
├── docker-compose.yml          # Docker 编排文件
├── README.md                   # 项目总览
└── .gitignore                  # 统一忽略规则
```

## 整合步骤

### 第一步：重构当前项目结构

1. **创建 backend 目录**

```bash
mkdir backend
```

2. **移动后端文件到 backend 目录**

```bash
# 移动所有Go后端相关文件
mv cmd backend/
mv internal backend/
mv pkg backend/
mv configs backend/
mv scripts backend/
mv go.mod backend/
mv go.sum backend/
mv main.go backend/ 2>/dev/null || echo "main.go already moved or doesn't exist"
```

3. **保留共享文档在根目录**

```bash
# docs 目录保留在根目录，作为项目整体文档
# CLAUDE.md 保留在根目录
```

### 第二步：添加前端项目

1. **在根目录创建前端项目**

```bash
# 使用 create-svelte 创建新项目
npm create svelte@latest frontend

# 或者如果你已有前端项目，直接复制到 frontend 目录
cp -r /path/to/your/frontend/project frontend/
```

2. **安装前端依赖**

```bash
cd frontend
npm install
```

### 第三步：配置开发环境

#### 1. 前端代理配置

在 `frontend/vite.config.js` 中配置API代理：

```javascript
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
        secure: false,
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, _res) => {
            console.log('proxy error', err);
          });
          proxy.on('proxyReq', (proxyReq, req, _res) => {
            console.log('Sending Request to the Target:', req.method, req.url);
          });
          proxy.on('proxyRes', (proxyRes, req, _res) => {
            console.log('Received Response from the Target:', proxyRes.statusCode, req.url);
          });
        },
      }
    }
  }
});
```

#### 2. 后端CORS配置

在 `backend/internal/` 中添加CORS中间件支持：

```bash
# 在backend目录下添加CORS支持
cd backend
go get github.com/gin-contrib/cors
```

更新 `backend/cmd/myblog/main.go`：

```go
import (
    "github.com/gin-contrib/cors"
    // ... 其他导入
)

func setupRoutes(r *gin.Engine, userHandler *handler.UserHandler) {
    // CORS配置
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:5173"} // 前端开发服务器
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
    r.Use(cors.New(config))

    // 其他路由配置...
}
```

### 第四步：统一启动脚本

#### 1. 创建根目录启动脚本

**Linux/Mac (`scripts/dev-full.sh`)**:

```bash
#!/bin/bash

echo "🚀 启动 MyBlog 全栈开发环境"

# 检查依赖
command -v go >/dev/null 2>&1 || { echo "❌ Go 未安装"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "❌ Node.js 未安装"; exit 1; }

echo "✅ 环境检查通过"

# 启动后端 (后台运行)
echo "🔧 启动后端服务..."
cd backend
go mod tidy
go run cmd/myblog/main.go &
BACKEND_PID=$!
cd ..

# 等待后端启动
sleep 3

# 检查后端是否启动成功
if curl -s http://localhost:3000/api/health > /dev/null; then
    echo "✅ 后端服务启动成功"
else
    echo "❌ 后端服务启动失败"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

# 启动前端
echo "🎨 启动前端服务..."
cd frontend
npm install
npm run dev &
FRONTEND_PID=$!
cd ..

echo "🎉 全栈开发环境启动完成!"
echo "📍 前端地址: http://localhost:5173"
echo "📍 后端地址: http://localhost:3000"
echo "💡 按 Ctrl+C 停止所有服务"

# 等待用户中断
trap "echo '🛑 正在停止服务...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT
wait
```

**Windows (`scripts/dev-full.bat`)**:

```batch
@echo off
cls

echo.
echo ==========================================
echo   MyBlog Full Stack Development
echo ==========================================
echo.

REM Check dependencies
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Go not installed
    pause
    exit /b 1
)

node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Node.js not installed
    pause
    exit /b 1
)

echo Dependencies OK
echo.

REM Start backend
echo Starting backend...
cd backend
start "Backend" cmd /c "go mod tidy && go run cmd/myblog/main.go"
timeout /t 3 /nobreak >nul
cd ..

REM Start frontend
echo Starting frontend...
cd frontend
start "Frontend" cmd /c "npm install && npm run dev"
cd ..

echo.
echo Full stack environment started!
echo Frontend: http://localhost:5173
echo Backend:  http://localhost:3000
echo.
echo Press any key to continue...
pause >nul
```

#### 2. 创建生产构建脚本

**构建脚本 (`scripts/build.sh`)**:

```bash
#!/bin/bash

echo "🏗️ 构建 MyBlog 项目"

# 构建前端
echo "📦 构建前端..."
cd frontend
npm install
npm run build
cd ..

# 构建后端
echo "🔧 构建后端..."
cd backend
go mod tidy
go build -o ../bin/myblog ./cmd/myblog
cd ..

echo "✅ 构建完成!"
echo "📁 前端构建文件: frontend/build/"
echo "📁 后端可执行文件: bin/myblog"
```

### 第五步：环境变量配置

#### 1. 前端环境变量

创建 `frontend/.env.development`:

```env
# 开发环境
VITE_API_BASE_URL=http://localhost:3000/api
VITE_APP_TITLE=MyBlog
```

创建 `frontend/.env.production`:

```env
# 生产环境
VITE_API_BASE_URL=/api
VITE_APP_TITLE=MyBlog
```

#### 2. 后端环境变量

更新 `backend/configs/config.yaml` 支持环境变量：

```yaml
server:
  host: ${SERVER_HOST:localhost}
  port: ${SERVER_PORT:3000}
  mode: ${GIN_MODE:debug}
```

### 第六步：Docker 支持

#### 1. 后端 Dockerfile

创建 `backend/Dockerfile`:

```dockerfile
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd/myblog

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
EXPOSE 3000
CMD ["./main"]
```

#### 2. 前端 Dockerfile

创建 `frontend/Dockerfile`:

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### 3. Docker Compose

创建根目录 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: 123456
      MYSQL_DATABASE: blog
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  backend:
    build: ./backend
    ports:
      - "3000:3000"
    depends_on:
      - mysql
    environment:
      - GIN_MODE=release
    volumes:
      - ./backend/configs:/root/configs

  frontend:
    build: ./frontend
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  mysql_data:
```

### 第七步：开发工作流

#### 1. 开发模式

```bash
# 启动完整开发环境
./scripts/dev-full.sh

# 或分别启动
cd backend && go run cmd/myblog/main.go &
cd frontend && npm run dev
```

#### 2. 生产部署

```bash
# Docker方式
docker-compose up -d

# 手动构建方式
./scripts/build.sh
```

### 第八步：API 调用示例

#### 前端API调用 (`frontend/src/lib/api.js`):

```javascript
// API基础配置
const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api';

// 通用请求函数
async function apiRequest(endpoint, options = {}) {
  const url = `${API_BASE}${endpoint}`;
  const config = {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  };

  try {
    const response = await fetch(url, config);
    const data = await response.json();
    
    if (data.code === 200) {
      return { success: true, data: data.data };
    } else {
      return { success: false, error: data.message || data.error };
    }
  } catch (error) {
    return { success: false, error: error.message };
  }
}

// 用户相关API
export const userAPI = {
  // 创建用户
  async create(userData) {
    return apiRequest('/users/create', {
      body: JSON.stringify(userData)
    });
  },

  // 获取用户列表
  async list(page = 1, pageSize = 10) {
    return apiRequest('/users/list', {
      body: JSON.stringify({ page, pageSize })
    });
  },

  // 获取用户详情
  async get(id) {
    return apiRequest('/users/get', {
      body: JSON.stringify({ id })
    });
  }
};
```

#### SvelteKit页面示例 (`frontend/src/routes/users/+page.svelte`):

```svelte
<script>
  import { onMount } from 'svelte';
  import { userAPI } from '$lib/api.js';

  let users = [];
  let loading = false;
  let error = null;

  onMount(async () => {
    await loadUsers();
  });

  async function loadUsers() {
    loading = true;
    error = null;
    
    const result = await userAPI.list();
    if (result.success) {
      users = result.data.users || [];
    } else {
      error = result.error;
    }
    
    loading = false;
  }

  async function createUser() {
    const userData = {
      username: 'test_user',
      email: 'test@example.com',
      password: '123456',
      nickname: 'Test User'
    };

    const result = await userAPI.create(userData);
    if (result.success) {
      await loadUsers(); // 重新加载列表
    } else {
      error = result.error;
    }
  }
</script>

<div class="container">
  <h1>用户管理</h1>
  
  <button on:click={createUser}>创建测试用户</button>
  
  {#if loading}
    <p>加载中...</p>
  {:else if error}
    <p class="error">错误: {error}</p>
  {:else if users.length > 0}
    <div class="users">
      {#each users as user}
        <div class="user-card">
          <h3>{user.username}</h3>
          <p>邮箱: {user.email}</p>
          <p>昵称: {user.nickname}</p>
          <p>创建时间: {user.createdAt}</p>
        </div>
      {/each}
    </div>
  {:else}
    <p>暂无用户数据</p>
  {/if}
</div>

<style>
  .container {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
  }
  
  .user-card {
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 16px;
  }
  
  .error {
    color: red;
  }
</style>
```

## 常见问题

### 1. 端口冲突

- 后端默认端口: 3000
- 前端默认端口: 5173
- 确保这些端口没有被其他服务占用

### 2. CORS 问题

- 确保后端配置了正确的CORS设置
- 开发环境允许 http://localhost:5173

### 3. API 路径问题

- 开发环境: 前端代理到 http://localhost:3000/api
- 生产环境: 前端直接请求 /api

### 4. 构建问题

- 确保前后端都能独立构建成功
- 检查依赖版本兼容性

## 总结

通过以上步骤，你将拥有一个完整的全栈开发环境，包括：

1. **开发体验**: 前后端热更新，API代理
2. **项目结构**: 清晰的Monorepo结构
3. **部署支持**: Docker化部署
4. **工作流**: 统一的开发和构建脚本

这种结构既保持了前后端的独立性，又提供了便捷的开发体验。
