#!/bin/bash

# MyBlog 全栈开发环境启动脚本

echo "🚀 启动 MyBlog 全栈开发环境"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查依赖
echo -e "${BLUE}[1/6]${NC} 检查环境依赖..."

if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go 未安装${NC}"
    echo "   请访问 https://golang.org/dl/ 下载安装"
    exit 1
fi

if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js 未安装${NC}"
    echo "   请访问 https://nodejs.org/ 下载安装"
    exit 1
fi

echo -e "${GREEN}✅ 环境依赖检查通过${NC}"
echo "   Go版本: $(go version | cut -d' ' -f3)"
echo "   Node版本: $(node --version)"
echo ""

# 检查项目结构
echo -e "${BLUE}[2/6]${NC} 检查项目结构..."

if [ ! -d "backend" ]; then
    echo -e "${YELLOW}⚠️  backend 目录不存在${NC}"
    echo "   请先运行项目重构脚本: ./scripts/restructure.sh"
    exit 1
fi

if [ ! -d "frontend" ]; then
    echo -e "${YELLOW}⚠️  frontend 目录不存在${NC}"
    echo "   请先创建前端项目或运行: ./scripts/restructure.sh"
    exit 1
fi

if [ ! -f "frontend/package.json" ]; then
    echo -e "${YELLOW}⚠️  前端项目未初始化${NC}"
    echo "   请在 frontend 目录下创建 SvelteKit 项目"
    echo "   或复制现有的前端项目文件到 frontend 目录"
    exit 1
fi

echo -e "${GREEN}✅ 项目结构检查通过${NC}"
echo ""

# 安装后端依赖
echo -e "${BLUE}[3/6]${NC} 安装后端依赖..."
cd backend
if [ ! -f "go.mod" ]; then
    echo -e "${RED}❌ 后端 go.mod 文件不存在${NC}"
    exit 1
fi

go mod tidy
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ 后端依赖安装失败${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 后端依赖安装完成${NC}"
cd ..

# 安装前端依赖
echo -e "${BLUE}[4/6]${NC} 安装前端依赖..."
cd frontend
npm install
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ 前端依赖安装失败${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 前端依赖安装完成${NC}"
cd ..

# 启动后端服务
echo -e "${BLUE}[5/6]${NC} 启动后端服务..."
cd backend

# 检查MySQL连接
echo "   正在检查数据库连接..."
if ! nc -z localhost 3306 2>/dev/null; then
    echo -e "${YELLOW}⚠️  MySQL未运行或无法连接到 localhost:3306${NC}"
    echo "   请确保MySQL服务已启动"
    echo "   或检查 backend/configs/config.yaml 中的数据库配置"
fi

# 后台启动后端
echo "   正在启动后端服务..."
nohup go run cmd/myblog/main.go > ../backend.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > ../backend.pid

# 等待后端启动
sleep 3

# 检查后端是否启动成功
echo "   正在检查后端服务状态..."
for i in {1..10}; do
    if curl -s http://localhost:3000/api/health >/dev/null 2>&1; then
        echo -e "${GREEN}✅ 后端服务启动成功${NC}"
        echo "   地址: http://localhost:3000"
        break
    fi
    if [ $i -eq 10 ]; then
        echo -e "${RED}❌ 后端服务启动失败${NC}"
        echo "   请检查 backend.log 日志文件"
        kill $BACKEND_PID 2>/dev/null
        exit 1
    fi
    sleep 1
done
cd ..

# 启动前端服务
echo -e "${BLUE}[6/6]${NC} 启动前端服务..."
cd frontend

# 后台启动前端
echo "   正在启动前端服务..."
nohup npm run dev > ../frontend.log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > ../frontend.pid

# 等待前端启动
sleep 3

# 检查前端是否启动成功
echo "   正在检查前端服务状态..."
for i in {1..10}; do
    if curl -s http://localhost:5173 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ 前端服务启动成功${NC}"
        echo "   地址: http://localhost:5173"
        break
    fi
    if [ $i -eq 10 ]; then
        echo -e "${YELLOW}⚠️  前端服务可能仍在启动中${NC}"
        echo "   请稍等片刻或检查 frontend.log 日志文件"
        break
    fi
    sleep 1
done
cd ..

echo ""
echo "🎉 MyBlog 全栈开发环境启动完成!"
echo ""
echo "📍 服务地址:"
echo "   前端应用: http://localhost:5173"
echo "   后端API:  http://localhost:3000/api"
echo "   健康检查: http://localhost:3000/api/health"
echo ""
echo "📋 进程信息:"
echo "   后端PID: $BACKEND_PID (保存在 backend.pid)"
echo "   前端PID: $FRONTEND_PID (保存在 frontend.pid)"
echo ""
echo "📄 日志文件:"
echo "   后端日志: backend.log"
echo "   前端日志: frontend.log"
echo ""
echo "🛑 停止服务:"
echo "   运行: ./scripts/stop.sh"
echo "   或手动: kill $BACKEND_PID $FRONTEND_PID"
echo ""
echo "💡 提示:"
echo "   - 修改后端代码需要手动重启后端服务"
echo "   - 前端代码修改会自动热更新"
echo "   - 按 Ctrl+C 不会停止后台服务，请使用 stop.sh"

# 创建停止脚本
cat > scripts/stop.sh << 'EOF'
#!/bin/bash

echo "🛑 停止 MyBlog 开发服务..."

# 停止后端
if [ -f "backend.pid" ]; then
    BACKEND_PID=$(cat backend.pid)
    if kill -0 $BACKEND_PID 2>/dev/null; then
        kill $BACKEND_PID
        echo "✅ 后端服务已停止 (PID: $BACKEND_PID)"
    else
        echo "⚠️  后端服务未运行"
    fi
    rm backend.pid
fi

# 停止前端
if [ -f "frontend.pid" ]; then
    FRONTEND_PID=$(cat frontend.pid)
    if kill -0 $FRONTEND_PID 2>/dev/null; then
        kill $FRONTEND_PID
        echo "✅ 前端服务已停止 (PID: $FRONTEND_PID)"
    else
        echo "⚠️  前端服务未运行"
    fi
    rm frontend.pid
fi

# 清理日志文件
if [ -f "backend.log" ]; then
    rm backend.log
    echo "🧹 后端日志已清理"
fi

if [ -f "frontend.log" ]; then
    rm frontend.log
    echo "🧹 前端日志已清理"
fi

echo "✅ 所有服务已停止"
EOF

chmod +x scripts/stop.sh

echo ""
echo "✨ 开发愉快! Happy Coding!"