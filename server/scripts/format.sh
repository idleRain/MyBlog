#!/bin/bash

# Go 代码格式化脚本
# 此脚本用于格式化 Go 代码，包括 gofmt 和 goimports

set -e

echo "🎨 开始格式化 Go 代码..."

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ 错误: 请在 Go 项目根目录运行此脚本"
    exit 1
fi

# 运行 gofmt
echo "📝 运行 gofmt 格式化代码..."
if ! go fmt ./...; then
    echo "❌ gofmt 格式化失败"
    exit 1
fi
echo "✅ gofmt 格式化完成"

# 检查并安装 goimports
if ! command -v goimports &> /dev/null; then
    echo "📦 goimports 未安装，正在安装..."
    if ! go install golang.org/x/tools/cmd/goimports@latest; then
        echo "❌ goimports 安装失败"
        exit 1
    fi
    echo "✅ goimports 安装完成"
fi

# 运行 goimports
echo "📝 运行 goimports 整理导入..."
if ! goimports -w .; then
    echo "❌ goimports 执行失败"
    exit 1
fi
echo "✅ goimports 整理完成"

# 检查并安装 gci (可选的导入排序工具)
if command -v gci &> /dev/null; then
    echo "📝 运行 gci 排序导入..."
    if ! gci write --skip-generated -s standard -s default -s "prefix($(go list -m))" .; then
        echo "⚠️  gci 执行失败，但继续执行"
    else
        echo "✅ gci 排序完成"
    fi
fi

echo "🎉 Go 代码格式化全部完成！"

# 显示格式化统计信息
echo ""
echo "📊 格式化统计:"
echo "   - 已处理的 Go 文件数量: $(find . -name "*.go" -not -path "./vendor/*" -not -path "./tmp/*" | wc -l)"
echo "   - 跳过的目录: vendor/, tmp/, .git/"