// Package config 提供应用程序配置管理功能
package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

// Config 应用程序配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	API      APIConfig      `mapstructure:"api"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	Charset      string `mapstructure:"charset"`
	ParseTime    bool   `mapstructure:"parse_time"`
	Loc          string `mapstructure:"loc"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// APIConfig API配置
type APIConfig struct {
	Version string `mapstructure:"version"`
	Timeout int    `mapstructure:"timeout"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	AccessSecret  string `mapstructure:"access_secret"`
	RefreshSecret string `mapstructure:"refresh_secret"`
	AccessExpire  int    `mapstructure:"access_expire"`  // 分钟
	RefreshExpire int    `mapstructure:"refresh_expire"` // 小时
	Issuer        string `mapstructure:"issuer"`
}

var (
	config *Config
	once   sync.Once
)

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	var err error
	once.Do(func() {
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")

		// 设置默认值
		setDefaults()

		// 启用环境变量支持
		viper.AutomaticEnv()
		viper.SetEnvPrefix("MYBLOG")

		// 读取配置文件
		if err = viper.ReadInConfig(); err != nil {
			err = fmt.Errorf("读取配置文件失败: %w", err)
			return
		}

		// 解析配置到结构体
		config = &Config{}
		if err = viper.Unmarshal(config); err != nil {
			err = fmt.Errorf("解析配置文件失败: %w", err)
			return
		}

		// 验证配置
		if err = validateConfig(config); err != nil {
			err = fmt.Errorf("配置验证失败: %w", err)
			return
		}
	})

	return config, err
}

// Get 获取全局配置实例
func Get() *Config {
	if config == nil {
		panic("配置未初始化，请先调用 Load() 方法")
	}
	return config
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		c.Database.Charset,
		c.Database.ParseTime,
		c.Database.Loc,
	)
}

// GetServerAddress 获取服务器监听地址
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// setDefaults 设置默认配置值
func setDefaults() {
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.parse_time", true)
	viper.SetDefault("database.loc", "Local")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)

	viper.SetDefault("logger.level", "debug")
	viper.SetDefault("logger.output", "stdout")

	viper.SetDefault("api.version", "v1")
	viper.SetDefault("api.timeout", 30)

	viper.SetDefault("jwt.access_secret", "myblog_access_secret_key_2025")
	viper.SetDefault("jwt.refresh_secret", "myblog_refresh_secret_key_2025")
	viper.SetDefault("jwt.access_expire", 15)
	viper.SetDefault("jwt.refresh_expire", 168)
	viper.SetDefault("jwt.issuer", "myblog")
}

// validateConfig 验证配置的有效性
func validateConfig(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", cfg.Server.Port)
	}

	if cfg.Database.Host == "" {
		return fmt.Errorf("数据库主机不能为空")
	}

	if cfg.Database.Username == "" {
		return fmt.Errorf("数据库用户名不能为空")
	}

	if cfg.Database.DBName == "" {
		return fmt.Errorf("数据库名不能为空")
	}

	if cfg.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT访问令牌密钥不能为空")
	}

	if cfg.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT刷新令牌密钥不能为空")
	}

	if cfg.JWT.AccessExpire <= 0 {
		return fmt.Errorf("JWT访问令牌过期时间必须大于0")
	}

	if cfg.JWT.RefreshExpire <= 0 {
		return fmt.Errorf("JWT刷新令牌过期时间必须大于0")
	}

	return nil
}
