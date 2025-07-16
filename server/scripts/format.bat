@echo off
REM Go 代码格式化脚本 (Windows 版本)
REM 此脚本用于格式化 Go 代码，包括 gofmt 和 goimports

echo 🎨 开始格式化 Go 代码...

REM 检查是否在正确的目录
if not exist "go.mod" (
    echo ❌ 错误: 请在 Go 项目根目录运行此脚本
    exit /b 1
)

REM 运行 gofmt
echo 📝 运行 gofmt 格式化代码...
go fmt ./...
if errorlevel 1 (
    echo ❌ gofmt 格式化失败
    exit /b 1
)
echo ✅ gofmt 格式化完成

REM 检查并安装 goimports
where goimports >nul 2>nul
if errorlevel 1 (
    echo 📦 goimports 未安装，正在安装...
    go install golang.org/x/tools/cmd/goimports@latest
    if errorlevel 1 (
        echo ❌ goimports 安装失败
        exit /b 1
    )
    echo ✅ goimports 安装完成
)

REM 运行 goimports
echo 📝 运行 goimports 整理导入...
goimports -w .
if errorlevel 1 (
    echo ❌ goimports 执行失败
    exit /b 1
)
echo ✅ goimports 整理完成

echo 🎉 Go 代码格式化全部完成！

REM 显示格式化统计信息
echo.
echo 📊 格式化统计:
for /f %%i in ('dir /s /b *.go ^| find /c /v ""') do echo    - 已处理的 Go 文件数量: %%i
echo    - 跳过的目录: vendor\, tmp\, .git\

pause