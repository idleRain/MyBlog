# Windows 开发环境使用指南

## 脚本乱码问题解决方案

如果在Windows环境下运行批处理脚本出现乱码，请按以下顺序尝试：

### 方案1: 使用最简版本（推荐）

```bash
scripts\start.bat
```

- 纯英文输出，无乱码问题
- 功能完整，自动监听文件变化
- 兼容所有Windows版本

### 方案2: 使用修复版本

```bash
scripts\dev-simple.bat
```

- 修复后的版本，应该不再有乱码
- 使用ASCII编码，兼容性更好

### 方案3: 使用PowerShell

```powershell
.\scripts\dev.ps1
```

- PowerShell版本，支持Unicode
- 功能最完整

## 常见问题排查

### 问题1: 出现乱码或无法识别的命令

**原因**: 批处理文件编码问题
**解决**: 使用 `scripts\start.bat`

### 问题2: Go命令不存在

**原因**: Go未安装或未添加到PATH
**解决**:

1. 下载并安装Go: https://golang.org/dl/
2. 确保Go安装目录在系统PATH中

### 问题3: 端口占用

**原因**: 3000端口被其他程序占用
**解决**:

1. 修改 `configs/config.yaml` 中的端口
2. 或关闭占用3000端口的程序

### 问题4: 权限不足

**原因**: 没有文件写入权限
**解决**: 以管理员身份运行命令行

## 手动启动方式

如果脚本都有问题，可以手动执行：

```bash
# 1. 清理和准备
rmdir /s /q tmp
mkdir tmp
go mod tidy

# 2. 编译
go build -o tmp/myblog.exe ./cmd/myblog

# 3. 运行
tmp\myblog.exe
```

## 开发建议

1. **推荐使用**: `scripts\start.bat` - 最稳定
2. **备选方案**: PowerShell 或手动启动
3. **避免使用**: 含中文字符的脚本

## 环境要求

- Windows 7 或更高版本
- Go 1.20 或更高版本
- MySQL 5.7 或更高版本

## 测试连接

启动成功后，访问以下地址测试：

- 健康检查: http://localhost:3000/api/health
- 使用POST方法发送请求
