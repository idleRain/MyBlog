# Config 模块

配置管理模块，负责应用程序配置的加载、解析和验证。

## 功能特性

- 支持YAML格式配置文件
- 单例模式确保配置全局唯一
- 配置参数验证
- 默认值设置
- 类型安全的配置访问

## 配置结构

### 服务器配置 (ServerConfig)
- `host`: 服务器监听主机
- `port`: 服务器监听端口
- `mode`: 运行模式 (debug/release/test)

### 数据库配置 (DatabaseConfig)
- `host`: 数据库主机地址
- `port`: 数据库端口
- `username`: 数据库用户名
- `password`: 数据库密码
- `dbname`: 数据库名称
- `charset`: 字符集
- `parse_time`: 是否解析时间类型
- `loc`: 时区设置
- `max_idle_conns`: 最大空闲连接数
- `max_open_conns`: 最大打开连接数

### 日志配置 (LoggerConfig)
- `level`: 日志级别 (debug/info/warn/error)
- `output`: 输出方式 (stdout/file)
- `file_path`: 日志文件路径

### API配置 (APIConfig)
- `version`: API版本
- `timeout`: 请求超时时间（秒）

## 使用示例

```go
package main

import (
    "log"
    "MyBlog/internal/config"
)

func main() {
    // 加载配置
    cfg, err := config.Load("configs/config.yaml")
    if err != nil {
        log.Fatal("配置加载失败:", err)
    }

    // 获取数据库连接串
    dsn := cfg.GetDSN()
    
    // 获取服务器地址
    addr := cfg.GetServerAddress()
    
    // 在其他地方获取全局配置
    globalCfg := config.Get()
}
```

## 配置文件示例

参见 `configs/config.yaml` 文件。

## 注意事项

1. 必须先调用 `Load()` 方法初始化配置
2. 配置采用单例模式，整个应用生命周期内只加载一次
3. 配置文件路径相对于项目根目录
4. 所有配置项都有默认值，可以在代码中查看