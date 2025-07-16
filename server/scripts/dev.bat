@echo off
chcp 65001 > nul

echo 🚀 启动 MyBlog 开发环境 (热更新模式)

REM 检查Go是否安装
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 错误: Go 未安装，请先安装 Go
    pause
    exit /b 1
)

REM 检查Air是否安装
air -v >nul 2>&1
if %errorlevel% neq 0 (
    echo 📦 Air 未安装，正在安装...
    go install github.com/cosmtrek/air@latest
    if %errorlevel% neq 0 (
        echo ❌ Air 安装失败，请手动安装: go install github.com/cosmtrek/air@latest
        pause
        exit /b 1
    )
    echo ✅ Air 安装完成
)

REM 显示配置信息
echo 📋 配置信息:
for /f "tokens=*" %%i in ('go version') do echo    - Go版本: %%i
echo    - 项目目录: %cd%
echo    - 监听端口: 3000
echo.

REM 清理临时文件
echo 🧹 清理临时文件...
if exist tmp rmdir /s /q tmp
mkdir tmp

REM 安装/更新依赖
echo 📦 检查并安装依赖...
go mod tidy
if %errorlevel% neq 0 (
    echo ❌ 依赖安装失败
    pause
    exit /b 1
)
echo ✅ 依赖检查完成
echo.

REM 启动热更新
echo 🔥 启动热更新模式...
echo 💡 提示: 修改代码后会自动重新编译和重启
echo 💡 按 Ctrl+C 停止开发服务器
echo.

air -c .air.toml