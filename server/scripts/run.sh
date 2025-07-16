#!/bin/bash

# MyBlog 应用启动脚本

echo "正在启动 MyBlog 应用..."

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装，请先安装 Go"
    exit 1
fi

# 显示Go版本
echo "Go版本: $(go version)"

# 清理并安装依赖
echo "正在清理并安装依赖..."
rm -f go.sum
go mod tidy

if [ $? -ne 0 ]; then
    echo "错误: 依赖安装失败"
    exit 1
fi

echo "依赖安装完成"

# 启动应用
echo "正在启动应用..."
go run cmd/myblog/main.go