#!/bin/bash

# MyBlog 项目重构脚本 - 为前端整合准备
echo "🔄 开始重构 MyBlog 项目结构..."

# 检查当前是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ 错误: 请在项目根目录运行此脚本"
    exit 1
fi

echo "📁 创建 backend 目录..."
mkdir -p backend

echo "📦 移动后端文件到 backend 目录..."

# 移动Go相关文件和目录
[ -d "cmd" ] && mv cmd backend/
[ -d "internal" ] && mv internal backend/
[ -d "pkg" ] && mv pkg backend/
[ -d "configs" ] && mv configs backend/
[ -f "go.mod" ] && mv go.mod backend/
[ -f "go.sum" ] && mv go.sum backend/
[ -f "main.go" ] && mv main.go backend/

# 移动后端相关的脚本
echo "🔧 整理脚本文件..."
mkdir -p backend/scripts
if [ -d "scripts" ]; then
    # 移动Go相关脚本到backend
    [ -f "scripts/watcher.go" ] && mv scripts/watcher.go backend/scripts/
    [ -f "scripts/run.sh" ] && mv scripts/run.sh backend/scripts/
    
    # 保留Windows和开发脚本在根目录
    # 这些将成为全栈开发脚本
fi

# 创建前端目录占位符
echo "🎨 创建 frontend 目录..."
mkdir -p frontend
cat > frontend/README.md << 'EOF'
# Frontend Project

请将你的 Vite + SvelteKit 项目文件放在这个目录下。

## 快速开始

如果你还没有前端项目，可以创建一个新的:

```bash
cd frontend
npm create svelte@latest .
npm install
```

## 配置

确保配置了正确的API代理设置，参见 `../docs/frontend-integration.md`
EOF

# 更新.gitignore
echo "📝 更新 .gitignore..."
cat > .gitignore << 'EOF'
# Go 编译产物
*.exe
*.exe~
*.dll
*.so
*.dylib
/bin/
/backend/tmp/

# 测试产物
*.test
*.out
/coverage/

# 依赖包目录
/vendor/

# Go 工作空间文件
go.work

# 环境变量文件
.env
.env.local
.env.development
.env.test
.env.production

# 日志文件
*.log
/logs/

# 数据库文件
*.db
*.sqlite
*.sqlite3

# IDE 配置文件
.vscode/
.idea/
*.swp
*.swo
*~

# 操作系统产生的文件
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# 临时文件
*.tmp
*.temp
/temp/

# Air 热更新产生的临时文件
/tmp/
air_tmp/

# 配置文件备份
*.yaml.bak
*.yml.bak
*.json.bak

# 构建产物
/dist/
/build/

# 证书文件
*.pem
*.key
*.crt

# Node.js 相关
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
pnpm-debug.log*
lerna-debug.log*

# Frontend 构建产物
/frontend/build/
/frontend/dist/
/frontend/.svelte-kit/

# Frontend 环境文件
/frontend/.env
/frontend/.env.local
/frontend/.env.development.local
/frontend/.env.test.local
/frontend/.env.production.local

# Package Manager
.pnpm-store/
.yarn/
.npm/

# Docker
docker-compose.override.yml
EOF

# 创建新的根目录README
echo "📖 创建新的 README.md..."
cat > README.md << 'EOF'
# MyBlog - 全栈博客系统

基于 Go + SvelteKit 的现代博客系统。

## 项目结构

```
MyBlog/
├── backend/          # Go 后端服务
├── frontend/         # SvelteKit 前端应用
├── docs/            # 项目文档
├── scripts/         # 开发脚本
└── README.md        # 项目说明
```

## 快速开始

### 开发环境

1. **安装依赖**
   - Go 1.20+
   - Node.js 18+
   - MySQL 8.0+

2. **启动服务**
   ```bash
   # 启动完整开发环境
   ./scripts/dev-full.sh
   
   # 或分别启动
   cd backend && go run cmd/myblog/main.go
   cd frontend && npm run dev
   ```

3. **访问应用**
   - 前端: http://localhost:5173
   - 后端API: http://localhost:3000/api

### 生产部署

```bash
# Docker 方式
docker-compose up -d

# 手动构建
./scripts/build.sh
```

## 文档

- [前端整合指南](docs/frontend-integration.md)
- [API 文档](docs/api/user_api.md)
- [Windows 开发环境](docs/windows-dev.md)

## 技术栈

**后端:**
- Go 1.20
- Gin Web Framework
- GORM + MySQL
- Viper 配置管理

**前端:**
- SvelteKit
- Vite
- TypeScript
- TailwindCSS

## 开发指南

详细的开发指南请参考 [前端整合文档](docs/frontend-integration.md)。
EOF

echo "🔧 更新后端脚本路径..."

# 更新后端CLAUDE.md中的路径引用
if [ -f "CLAUDE.md" ]; then
    # 备份原文件
    cp CLAUDE.md CLAUDE.md.backup
    
    # 更新路径引用
    sed -i.bak 's/go run cmd\/myblog\/main.go/cd backend \&\& go run cmd\/myblog\/main.go/g' CLAUDE.md
    sed -i.bak 's/configs\/config.yaml/backend\/configs\/config.yaml/g' CLAUDE.md
    rm CLAUDE.md.bak
    
    echo "📝 CLAUDE.md 已更新，备份保存为 CLAUDE.md.backup"
fi

# 创建全栈开发的 CLAUDE.md 更新
cat >> CLAUDE.md << 'EOF'

## 全栈开发

### 项目结构
项目现已重构为 Monorepo 结构：
- `backend/` - Go 后端服务
- `frontend/` - SvelteKit 前端应用

### 全栈启动
```bash
# 启动完整开发环境
./scripts/dev-full.sh

# Windows 用户
scripts\dev-full.bat
```

### 前端整合
详细的前端整合指南请参考 `docs/frontend-integration.md`
EOF

echo "✅ 项目重构完成!"
echo ""
echo "📋 重构总结:"
echo "   - ✅ 后端文件已移动到 backend/ 目录"
echo "   - ✅ 创建了 frontend/ 目录占位符"
echo "   - ✅ 更新了 .gitignore 文件"
echo "   - ✅ 创建了新的 README.md"
echo "   - ✅ 更新了 CLAUDE.md"
echo ""
echo "🎯 下一步操作:"
echo "   1. 将你的前端项目文件复制到 frontend/ 目录"
echo "   2. 配置前端的 vite.config.js 代理设置"
echo "   3. 在后端添加 CORS 支持"
echo "   4. 运行 ./scripts/dev-full.sh 启动全栈开发环境"
echo ""
echo "📖 详细指南请查看: docs/frontend-integration.md"