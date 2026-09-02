# MyBlog Makefile

.PHONY: help dev build run clean test deps air-install lint lint-install format quality-check

# 默认目标
help:
	@echo "MyBlog 项目命令:"
	@echo "  make dev        - 启动开发环境 (热更新)"
	@echo "  make build      - 编译项目"
	@echo "  make run        - 运行编译后的程序"
	@echo "  make clean      - 清理临时文件"
	@echo "  make test       - 运行测试"
	@echo "  make deps       - 安装/更新依赖"
	@echo "  make air-install- 安装Air热更新工具"
	@echo "  make lint       - 运行代码检查 (golangci-lint)"
	@echo "  make lint-install- 安装golangci-lint工具"
	@echo "  make format     - 格式化代码 (gofmt + goimports)"
	@echo "  make quality-check- 运行完整代码质量检查"

# 开发环境 (热更新)
dev:
	@echo "🚀 启动开发环境..."
	@if ! command -v air > /dev/null; then \
		echo "📦 正在安装 Air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@mkdir -p tmp
	@air -c .air.toml

# 编译项目
build:
	@echo "🔨 编译项目..."
	@mkdir -p bin
	@go build -o bin/myblog ./cmd/myblog
	@echo "✅ 编译完成: bin/myblog"

# 运行项目
run: build
	@echo "🚀 运行项目..."
	@./bin/myblog

# 清理临时文件
clean:
	@echo "🧹 清理临时文件..."
	@rm -rf tmp/ bin/ logs/
	@echo "✅ 清理完成"

# 运行测试
test:
	@echo "🧪 运行测试..."
	@go test -v ./...

# 安装/更新依赖
deps:
	@echo "📦 安装/更新依赖..."
	@go mod tidy
	@go mod download
	@echo "✅ 依赖更新完成"

# 安装Air热更新工具
air-install:
	@echo "📦 安装 Air 热更新工具..."
	@go install github.com/air-verse/air@latest
	@echo "✅ Air 安装完成"

# 安装golangci-lint工具
lint-install:
	@echo "📦 安装 golangci-lint 工具..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "正在下载并安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2; \
	else \
		echo "golangci-lint 已安装"; \
	fi
	@echo "✅ golangci-lint 安装完成"

# 运行代码检查 (golangci-lint)
lint:
	@echo "🔍 运行 golangci-lint 代码检查..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "📦 golangci-lint 未安装，正在安装..."; \
		$(MAKE) lint-install; \
	fi
	@golangci-lint run
	@echo "✅ 代码检查完成"

# 格式化代码 (gofmt + goimports)
format:
	@echo "🎨 格式化代码..."
	@echo "运行 gofmt..."
	@go fmt ./...
	@echo "运行 goimports..."
	@if ! command -v goimports > /dev/null; then \
		echo "📦 正在安装 goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@goimports -w .
	@echo "✅ 代码格式化完成"

# 使用 Node.js 脚本格式化代码 (跨平台)
format-js:
	@node scripts/format-code.js

# 格式化代码 (保持向后兼容)
fmt: format

# 代码检查
vet:
	@echo "🔍 代码检查..."
	@go vet ./...
	@echo "✅ 代码检查完成"

# 运行完整代码质量检查
quality-check: format lint vet test
	@echo "✅ 完整代码质量检查完成"
	@echo "📝 提示: 在 monorepo 根目录运行 'npm run lint:go' 可以在 CI/CD 流程中执行此检查"

# 完整检查 (格式化 + 检查 + 测试) - 保持向后兼容
check: quality-check