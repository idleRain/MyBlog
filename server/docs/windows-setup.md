# Windows 环境配置指南

## 问题分析

Windows环境下可能遇到的问题：

1. `make` 命令不存在
2. `air` 工具未安装或不在PATH中
3. Go bin目录不在环境变量中

## 解决方案

### 方案1: 安装必要工具 (推荐)

#### 1. 安装Air热更新工具

```powershell
# 在PowerShell或CMD中运行
go install github.com/cosmtrek/air@latest
```

#### 2. 配置环境变量

检查Go bin目录是否在PATH中：

```powershell
# 查看Go环境信息
go env GOPATH
go env GOBIN

# 查看PATH环境变量
echo $env:PATH
```

如果Air不能运行，需要添加Go bin目录到PATH：

1. 找到Go bin目录 (通常是 `%USERPROFILE%\go\bin`)
2. 添加到系统PATH环境变量中
3. 重启命令行窗口

#### 3. 启动热更新

```powershell
# 方法1: 直接使用air
air -c .air.toml

# 方法2: 使用批处理脚本
scripts\dev.bat
```

### 方案2: 使用PowerShell脚本

创建PowerShell版本的启动脚本。

### 方案3: 手动热更新

如果air安装有问题，可以手动实现简单的热更新。

## 故障排除

### Air命令不存在

```powershell
# 重新安装Air
go install github.com/cosmtrek/air@latest

# 检查安装位置
where air

# 如果找不到，检查Go bin目录
go env GOPATH
```

### 权限问题

以管理员身份运行PowerShell或CMD。

### 编译错误

```powershell
# 清理依赖
go clean -modcache
go mod tidy
```

## 测试安装

```powershell
# 测试Go
go version

# 测试Air
air -v

# 启动项目
air -c .air.toml
```
