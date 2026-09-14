// Package config 提供应用程序配置管理功能
package config

import (
	"fmt"
	"os"
	"strings"
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
	Security SecurityConfig `mapstructure:"security"`
	Media    MediaConfig    `mapstructure:"media"`
	RBAC     RBACConfig     `mapstructure:"rbac"`
}

// RBACConfig 权限配置，角色层级与权限映射的生产环境唯一权威。
type RBACConfig struct {
	RoleHierarchy   map[string]int      `mapstructure:"role_hierarchy"`   // 角色层级，数值越大权限越高
	RolePermissions map[string][]string `mapstructure:"role_permissions"` // 角色到权限列表的映射
}

// MediaConfig 媒体文件存储配置
type MediaConfig struct {
	UploadDir    string   `mapstructure:"upload_dir"`    // 本地存储目录
	BaseURL      string   `mapstructure:"base_url"`      // 文件访问 URL 前缀
	MaxSizeMB    int      `mapstructure:"max_size_mb"`   // 单文件大小上限，单位 MB
	AllowedTypes []string `mapstructure:"allowed_types"` // 允许的 MIME 类型，空表示不限制
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

// SecurityConfig 安全配置
type SecurityConfig struct {
	RateLimit       RateLimitConfig       `mapstructure:"rate_limit"`
	SecurityHeaders SecurityHeadersConfig `mapstructure:"security_headers"`
	InputValidation InputValidationConfig `mapstructure:"input_validation"`
	AdminSecurity   AdminSecurityConfig   `mapstructure:"admin_security"`
}

// RateLimitConfig 频率限制配置
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	MaxRequests       int  `mapstructure:"max_requests"`
	WindowMinutes     int  `mapstructure:"window_minutes"`
	UserMaxRequests   int  `mapstructure:"user_max_requests"`
	UserWindowMinutes int  `mapstructure:"user_window_minutes"`
}

// SecurityHeadersConfig 安全头配置
type SecurityHeadersConfig struct {
	Enabled                 bool   `mapstructure:"enabled"`
	ContentSecurityPolicy   string `mapstructure:"content_security_policy"`
	XFrameOptions           string `mapstructure:"x_frame_options"`
	XContentTypeOptions     string `mapstructure:"x_content_type_options"`
	ReferrerPolicy          string `mapstructure:"referrer_policy"`
	StrictTransportSecurity string `mapstructure:"strict_transport_security"`
}

// InputValidationConfig 输入验证配置
type InputValidationConfig struct {
	Enabled           bool     `mapstructure:"enabled"`
	MaxRequestSizeMB  int      `mapstructure:"max_request_size_mb"`
	BlockedUserAgents []string `mapstructure:"blocked_user_agents"`
}

// AdminSecurityConfig 管理员安全配置
type AdminSecurityConfig struct {
	Enabled         bool     `mapstructure:"enabled"`
	MaxRequests     int      `mapstructure:"max_requests"`
	UserMaxRequests int      `mapstructure:"user_max_requests"`
	IPWhitelist     []string `mapstructure:"ip_whitelist"`
}

var (
	config *Config
	once   sync.Once
)

// envKeyPrefix 环境变量统一前缀，完整变量名由该前缀与配置键转换拼接得到。
const envKeyPrefix = "MYBLOG"

// knownWeakSecrets 历史版本公开泄漏过的弱密钥集合，任一 JWT 密钥命中都必须拒绝启动。
// 这些值已随公开仓库扩散，继续使用等同于放弃令牌签名防护。
var knownWeakSecrets = []string{
	"myblog_access_secret_key_2025",
	"myblog_refresh_secret_key_2025",
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	var err error
	once.Do(func() {
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")

		// 设置默认值
		setDefaults()

		// 读取配置文件
		if err = viper.ReadInConfig(); err != nil {
			err = fmt.Errorf("读取配置文件失败: %w", err)
			return
		}

		// 环境变量覆盖标量配置，敏感项在生产环境经环境注入，不再依赖 YAML 明文。
		applyEnvOverrides(viper.GetViper())

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

	// JWT 密钥不设代码默认值，缺失时由 validateConfig 拒绝启动，避免弱默认值随代码分发。
	viper.SetDefault("jwt.access_expire", 15)
	viper.SetDefault("jwt.refresh_expire", 168)
	viper.SetDefault("jwt.issuer", "myblog")

	// 安全配置默认值
	viper.SetDefault("security.rate_limit.enabled", true)
	viper.SetDefault("security.rate_limit.max_requests", 100)
	viper.SetDefault("security.rate_limit.window_minutes", 1)
	viper.SetDefault("security.rate_limit.user_max_requests", 300)
	viper.SetDefault("security.rate_limit.user_window_minutes", 1)

	viper.SetDefault("security.security_headers.enabled", true)
	viper.SetDefault("security.security_headers.content_security_policy", "default-src 'self'")
	viper.SetDefault("security.security_headers.x_frame_options", "SAMEORIGIN")
	viper.SetDefault("security.security_headers.x_content_type_options", "nosniff")
	viper.SetDefault("security.security_headers.referrer_policy", "strict-origin-when-cross-origin")
	viper.SetDefault("security.security_headers.strict_transport_security", "max-age=31536000; includeSubDomains")

	viper.SetDefault("security.input_validation.enabled", true)
	viper.SetDefault("security.input_validation.max_request_size_mb", 10)

	viper.SetDefault("security.admin_security.enabled", true)
	viper.SetDefault("security.admin_security.max_requests", 30)
	viper.SetDefault("security.admin_security.user_max_requests", 50)

	// 媒体文件存储默认配置
	viper.SetDefault("media.upload_dir", "uploads")
	viper.SetDefault("media.base_url", "/uploads")
	viper.SetDefault("media.max_size_mb", 10)
}

// applyEnvOverrides 使用环境变量覆盖标量配置项，环境变量优先级高于 YAML 与代码默认值。
// 映射规则为 MYBLOG_ 前缀加配置键的大写下划线形式，例如 jwt.access_secret 对应 MYBLOG_JWT_ACCESS_SECRET。
// 列表与映射结构无法经单个环境变量无损表达，此类配置项保持 YAML 原值。
func applyEnvOverrides(v *viper.Viper) {
	for _, key := range v.AllKeys() {
		if !isScalarConfigValue(v.Get(key)) {
			continue
		}
		if value, ok := os.LookupEnv(envVariableName(key)); ok {
			v.Set(key, value)
		}
	}
}

// envVariableName 将配置键转换为环境变量名，点分隔符替换为下划线并整体转为大写。
func envVariableName(key string) string {
	upperKey := strings.ToUpper(key)
	return envKeyPrefix + "_" + strings.ReplaceAll(upperKey, ".", "_")
}

// isScalarConfigValue 判断配置值是否为可经环境变量表达的标量，列表与映射等容器类型返回 false。
func isScalarConfigValue(value any) bool {
	switch value.(type) {
	case map[string]any, map[any]any, []any:
		return false
	default:
		return true
	}
}

// isKnownWeakSecret 判断给定密钥是否命中已公开的弱密钥集合。
func isKnownWeakSecret(secret string) bool {
	for _, weak := range knownWeakSecrets {
		if secret == weak {
			return true
		}
	}
	return false
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

	// 密钥为已公开的弱默认值时直接拒绝启动，强制部署方轮换为随机强密钥。
	if isKnownWeakSecret(cfg.JWT.AccessSecret) {
		return fmt.Errorf("JWT访问令牌密钥使用了已公开的弱默认值，必须更换为随机强密钥")
	}

	if isKnownWeakSecret(cfg.JWT.RefreshSecret) {
		return fmt.Errorf("JWT刷新令牌密钥使用了已公开的弱默认值，必须更换为随机强密钥")
	}

	// 双密钥互异校验，防止单一密钥泄漏同时波及访问令牌与刷新令牌两条签名链路。
	if cfg.JWT.AccessSecret == cfg.JWT.RefreshSecret {
		return fmt.Errorf("JWT访问令牌密钥与刷新令牌密钥不能相同")
	}

	// RBAC 配置校验：四类角色必须全部登记层级与权限。
	for _, role := range []string{"superadmin", "admin", "editor", "user"} {
		if _, ok := cfg.RBAC.RoleHierarchy[role]; !ok {
			return fmt.Errorf("RBAC角色层级缺少 %s 的定义", role)
		}
		if _, ok := cfg.RBAC.RolePermissions[role]; !ok {
			return fmt.Errorf("RBAC角色权限映射缺少 %s 的定义", role)
		}
	}

	return nil
}
