#!/bin/bash

# MyBlog 开发环境热更新启动脚本

echo "🚀 启动 MyBlog 开发环境 (热更新模式)"

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误: Go 未安装，请先安装 Go"
    exit 1
fi

# 检查Air是否安装
if ! command -v air &> /dev/null; then
    echo "📦 Air 未安装，正在安装..."
    if command -v go &> /dev/null; then
        go install github.com/cosmtrek/air@latest
        echo "✅ Air 安装完成"
    else
        echo "❌ 无法安装 Air，请手动安装: go install github.com/cosmtrek/air@latest"
        exit 1
    fi
fi

# 显示配置信息
echo "📋 配置信息:"
echo "   - Go版本: $(go version)"
echo "   - Air版本: $(air -v 2>/dev/null || echo '未知')"
echo "   - 项目目录: $(pwd)"
echo "   - 监听端口: 3000"
echo ""

# 清理临时文件
echo "🧹 清理临时文件..."
rm -rf tmp/
mkdir -p tmp/

# 安装/更新依赖
echo "📦 检查并安装依赖..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "❌ 依赖安装失败"
    exit 1
fi

echo "✅ 依赖检查完成"
echo ""

# 启动热更新
echo "🔥 启动热更新模式..."
echo "💡 提示: 修改代码后会自动重新编译和重启"
echo "💡 按 Ctrl+C 停止开发服务器"
echo ""

# 使用Air启动热更新
air -c .air.toml