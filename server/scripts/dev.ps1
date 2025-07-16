# MyBlog PowerShell 开发环境启动脚本

Write-Host "🚀 启动 MyBlog 开发环境 (热更新模式)" -ForegroundColor Green

# 检查Go是否安装
try {
    $goVersion = go version
    Write-Host "✅ Go已安装: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ 错误: Go 未安装，请先安装 Go" -ForegroundColor Red
    Write-Host "下载地址: https://golang.org/dl/" -ForegroundColor Yellow
    Read-Host "按任意键退出"
    exit 1
}

# 检查Air是否安装
try {
    $airVersion = air -v 2>$null
    Write-Host "✅ Air已安装" -ForegroundColor Green
} catch {
    Write-Host "📦 Air 未安装，正在安装..." -ForegroundColor Yellow
    try {
        go install github.com/cosmtrek/air@latest
        Write-Host "✅ Air 安装完成" -ForegroundColor Green
    } catch {
        Write-Host "❌ Air 安装失败，请检查网络连接或手动安装" -ForegroundColor Red
        Write-Host "手动安装命令: go install github.com/cosmtrek/air@latest" -ForegroundColor Yellow
        Read-Host "按任意键退出"
        exit 1
    }
}

# 显示配置信息
Write-Host "📋 配置信息:" -ForegroundColor Cyan
Write-Host "   - Go版本: $goVersion" -ForegroundColor White
Write-Host "   - 项目目录: $(Get-Location)" -ForegroundColor White
Write-Host "   - 监听端口: 3000" -ForegroundColor White
Write-Host ""

# 清理临时文件
Write-Host "🧹 清理临时文件..." -ForegroundColor Yellow
if (Test-Path "tmp") {
    Remove-Item -Recurse -Force "tmp"
}
New-Item -ItemType Directory -Force -Path "tmp" | Out-Null

# 安装/更新依赖
Write-Host "📦 检查并安装依赖..." -ForegroundColor Yellow
try {
    go mod tidy
    Write-Host "✅ 依赖检查完成" -ForegroundColor Green
} catch {
    Write-Host "❌ 依赖安装失败" -ForegroundColor Red
    Read-Host "按任意键退出"
    exit 1
}

Write-Host ""

# 启动热更新
Write-Host "🔥 启动热更新模式..." -ForegroundColor Green
Write-Host "💡 提示: 修改代码后会自动重新编译和重启" -ForegroundColor Cyan
Write-Host "💡 按 Ctrl+C 停止开发服务器" -ForegroundColor Cyan
Write-Host ""

# 检查Air配置文件
if (-not (Test-Path ".air.toml")) {
    Write-Host "❌ 未找到 .air.toml 配置文件" -ForegroundColor Red
    Read-Host "按任意键退出"
    exit 1
}

# 启动Air
try {
    air -c .air.toml
} catch {
    Write-Host "❌ Air 启动失败" -ForegroundColor Red
    Write-Host "请检查配置文件或手动运行: air -c .air.toml" -ForegroundColor Yellow
    Read-Host "按任意键退出"
    exit 1
}