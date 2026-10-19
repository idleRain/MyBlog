package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

// TestEnvVariableName 验证配置键到环境变量名的映射规则。
func TestEnvVariableName(t *testing.T) {
	cases := []struct {
		key      string
		expected string
	}{
		{key: "token.access_expire", expected: "MYBLOG_TOKEN_ACCESS_EXPIRE"},
		{key: "database.password", expected: "MYBLOG_DATABASE_PASSWORD"},
		{key: "server.port", expected: "MYBLOG_SERVER_PORT"},
	}

	for _, item := range cases {
		if actual := envVariableName(item.key); actual != item.expected {
			t.Errorf("envVariableName(%q) = %q, 期望 %q", item.key, actual, item.expected)
		}
	}
}

// TestApplyEnvOverridesScalar 验证标量配置项可被环境变量覆盖。
func TestApplyEnvOverridesScalar(t *testing.T) {
	t.Setenv("MYBLOG_TOKEN_ACCESS_EXPIRE", "30")
	t.Setenv("MYBLOG_SERVER_PORT", "8080")

	v := viper.New()
	v.SetDefault("token.access_expire", 15)
	v.SetDefault("server.port", 3000)

	applyEnvOverrides(v)

	if actual := v.GetInt("token.access_expire"); actual != 30 {
		t.Errorf("access_expire = %d, 期望环境变量覆盖值 30", actual)
	}
	if actual := v.GetInt("server.port"); actual != 8080 {
		t.Errorf("server.port = %d, 期望环境变量覆盖值 8080", actual)
	}
}

// TestApplyEnvOverridesSkipsContainer 验证列表配置项不受环境变量影响而标量项正常覆盖。
func TestApplyEnvOverridesSkipsContainer(t *testing.T) {
	t.Setenv("MYBLOG_SECURITY_INPUT_VALIDATION_BLOCKED_USER_AGENTS", "env-agent")
	t.Setenv("MYBLOG_SECURITY_RATE_LIMIT_ENABLED", "true")

	v := viper.New()
	v.Set("security.input_validation.blocked_user_agents", []any{"curl"})
	v.Set("security.rate_limit.enabled", false)

	applyEnvOverrides(v)

	list := v.Get("security.input_validation.blocked_user_agents").([]any)
	if len(list) != 1 || list[0] != "curl" {
		t.Errorf("blocked_user_agents = %v, 期望保持 YAML 原值不被覆盖", list)
	}
	if !v.GetBool("security.rate_limit.enabled") {
		t.Error("标量配置项应被环境变量覆盖，实际未被覆盖")
	}
}

// TestApplyEnvOverridesKeepsOriginalWhenEnvMissing 验证未设置环境变量时配置保持原值。
func TestApplyEnvOverridesKeepsOriginalWhenEnvMissing(t *testing.T) {
	v := viper.New()
	v.Set("database.host", "localhost")

	applyEnvOverrides(v)

	if actual := v.GetString("database.host"); actual != "localhost" {
		t.Errorf("database.host = %q, 期望保持原值 localhost", actual)
	}
}

// validTestConfig 构造可通过全部校验的最小配置，供非法值叠加测试复用。
func validTestConfig() *Config {
	return &Config{
		Server:   ServerConfig{Port: 3000},
		Database: DatabaseConfig{Host: "localhost", Username: "root", DBName: "blog"},
		Media:    MediaConfig{MaxSizeMB: 10},
		Token: TokenConfig{
			AccessExpire:  15,
			RefreshExpire: 168,
		},
		RBAC: RBACConfig{
			RoleHierarchy: map[string]int{
				"superadmin": 4,
				"admin":      3,
				"editor":     2,
				"user":       1,
			},
			RolePermissions: map[string][]string{
				"superadmin": {"article:read"},
				"admin":      {"article:read"},
				"editor":     {"article:read"},
				"user":       {"article:read"},
			},
		},
		I18N: I18NConfig{
			DefaultLanguage:    "zh",
			SupportedLanguages: []string{"zh", "en"},
		},
	}
}

// TestValidateConfigAcceptsValidConfig 验证合法配置通过校验。
func TestValidateConfigAcceptsValidConfig(t *testing.T) {
	if err := validateConfig(validTestConfig()); err != nil {
		t.Errorf("合法配置不应返回错误: %v", err)
	}
}

// TestValidateConfigRejectsNonPositiveMediaMaxSize 验证媒体单文件大小上限非正即拒绝启动。
// 上限为零会让上传读取被截断为一个字节，而声明大小照原样落库，存储内容与记录大小不一致。
func TestValidateConfigRejectsNonPositiveMediaMaxSize(t *testing.T) {
	for _, maxSizeMB := range []int{0, -1} {
		cfg := validTestConfig()
		cfg.Media.MaxSizeMB = maxSizeMB

		if err := validateConfig(cfg); err == nil {
			t.Errorf("media.max_size_mb = %d，应返回错误", maxSizeMB)
		}
	}
}

// TestValidateConfigRejectsEmptyI18NSupportedLanguages 验证 i18n 白名单为空即拒绝启动。
func TestValidateConfigRejectsEmptyI18NSupportedLanguages(t *testing.T) {
	cfg := validTestConfig()
	cfg.I18N.SupportedLanguages = nil
	if err := validateConfig(cfg); err == nil {
		t.Error("受支持语言列表为空应返回错误")
	}
}

// TestValidateConfigRejectsEmptyI18NDefaultLanguage 验证 i18n 缺省语言缺失即拒绝启动。
func TestValidateConfigRejectsEmptyI18NDefaultLanguage(t *testing.T) {
	cfg := validTestConfig()
	cfg.I18N.DefaultLanguage = ""
	if err := validateConfig(cfg); err == nil {
		t.Error("缺省语言为空应返回错误")
	}
}

// TestValidateConfigRejectsI18NDefaultOutsideWhitelist 验证缺省语言不在白名单内即拒绝启动。
func TestValidateConfigRejectsI18NDefaultOutsideWhitelist(t *testing.T) {
	cfg := validTestConfig()
	cfg.I18N.DefaultLanguage = "fr"
	if err := validateConfig(cfg); err == nil {
		t.Error("缺省语言不在白名单内应返回错误")
	}
}

// TestValidateConfigRejectsInvalidMetricsAddr 验证启用指标时非法监听地址即拒绝启动。
func TestValidateConfigRejectsInvalidMetricsAddr(t *testing.T) {
	cfg := validTestConfig()
	cfg.Metrics.Enabled = true
	cfg.Metrics.Addr = "127.0.0.1"
	if err := validateConfig(cfg); err == nil {
		t.Error("缺少端口的指标监听地址应返回错误")
	}
}

// TestValidateConfigIgnoresMetricsAddrWhenDisabled 验证未启用指标时不校验监听地址。
func TestValidateConfigIgnoresMetricsAddrWhenDisabled(t *testing.T) {
	cfg := validTestConfig()
	cfg.Metrics.Enabled = false
	cfg.Metrics.Addr = ""
	if err := validateConfig(cfg); err != nil {
		t.Errorf("未启用指标时不应校验监听地址: %v", err)
	}
}

// TestLoadParsesConfigSections 验证 config.yaml 的 rbac、i18n 与 token 节可被正确解析并通过校验。
func TestLoadParsesConfigSections(t *testing.T) {
	cfg, err := Load("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.RBAC.RoleHierarchy["superadmin"] != 4 {
		t.Error("superadmin 层级应为 4")
	}
	if len(cfg.RBAC.RolePermissions["superadmin"]) == 0 {
		t.Error("superadmin 应配置至少一条权限")
	}
	if len(cfg.RBAC.RolePermissions["user"]) == 0 {
		t.Error("user 角色应配置基础权限")
	}
	if cfg.I18N.DefaultLanguage != "zh" {
		t.Errorf("DefaultLanguage = %q, 期望解析为 zh", cfg.I18N.DefaultLanguage)
	}
	if len(cfg.I18N.SupportedLanguages) == 0 {
		t.Error("受支持语言列表应解析出成员")
	}
	// 令牌有效期带有默认值，键名写错不会报错而是静默回退到默认值，
	// 因此必须直接断言 YAML 声明的是 token 节而非历史遗留的 jwt 节。
	if !viper.InConfig("token.access_expire") || !viper.InConfig("token.refresh_expire") {
		t.Error("config.yaml 应声明 token.access_expire 与 token.refresh_expire")
	}
	if viper.InConfig("jwt.access_expire") || viper.InConfig("jwt.refresh_expire") {
		t.Error("config.yaml 不应再出现 jwt 节")
	}
	// 代理信任链默认收敛：声明值缺失会被代码默认值兜住，但两者必须一致且仅覆盖回环。
	if len(cfg.Security.TrustedProxies) != 2 {
		t.Errorf("security.trusted_proxies = %v, 期望声明两个回环信任项", cfg.Security.TrustedProxies)
	}
}

// TestValidateConfigRejectsInvalidTrustedProxy 验证非法代理信任项即拒绝启动。
func TestValidateConfigRejectsInvalidTrustedProxy(t *testing.T) {
	cfg := validTestConfig()
	cfg.Security.TrustedProxies = []string{"127.0.0.1/32", "not-an-address"}

	if err := validateConfig(cfg); err == nil {
		t.Error("非法代理信任项应返回错误")
	}
}

// TestValidateConfigRejectsTrustAllTrustedProxy 验证覆盖全部地址的信任项即拒绝启动。
// 该取值等同于关闭信任链校验，与默认收敛的方向相反，必须拦截。
func TestValidateConfigRejectsTrustAllTrustedProxy(t *testing.T) {
	for _, entry := range []string{"0.0.0.0/0", "::/0"} {
		cfg := validTestConfig()
		cfg.Security.TrustedProxies = []string{entry}

		if err := validateConfig(cfg); err == nil {
			t.Errorf("信任项 %q 覆盖全部地址，应返回错误", entry)
		}
	}
}

// TestValidateConfigAcceptsTrustedProxyEntries 验证单 IP 与网段两种写法均通过校验。
func TestValidateConfigAcceptsTrustedProxyEntries(t *testing.T) {
	cfg := validTestConfig()
	cfg.Security.TrustedProxies = []string{"127.0.0.1", "::1", "172.16.0.0/12"}

	if err := validateConfig(cfg); err != nil {
		t.Errorf("合法代理信任项不应返回错误: %v", err)
	}
}

// TestValidateConfigAcceptsEmptyTrustedProxies 验证空列表表示服务前方没有代理，属合法配置。
func TestValidateConfigAcceptsEmptyTrustedProxies(t *testing.T) {
	cfg := validTestConfig()
	cfg.Security.TrustedProxies = nil

	if err := validateConfig(cfg); err != nil {
		t.Errorf("未配置代理信任项不应返回错误: %v", err)
	}
}

// TestApplyTrustedProxiesOverrideSplitsList 验证列表值可经专用环境变量按逗号注入。
func TestApplyTrustedProxiesOverrideSplitsList(t *testing.T) {
	t.Setenv(trustedProxiesEnvKey, "172.16.0.0/12, 127.0.0.1/32")

	v := viper.New()
	v.Set("security.trusted_proxies", []string{"127.0.0.1/32"})

	applyTrustedProxiesOverride(v)

	entries := v.GetStringSlice("security.trusted_proxies")
	if len(entries) != 2 || entries[0] != "172.16.0.0/12" || entries[1] != "127.0.0.1/32" {
		t.Errorf("trusted_proxies = %v, 期望解析出两个信任项且按声明顺序排列", entries)
	}
}

// TestApplyTrustedProxiesOverrideEmptyMeansNoProxy 验证变量取值为空串时解析为空列表。
func TestApplyTrustedProxiesOverrideEmptyMeansNoProxy(t *testing.T) {
	t.Setenv(trustedProxiesEnvKey, "")

	v := viper.New()
	v.Set("security.trusted_proxies", []string{"127.0.0.1/32"})

	applyTrustedProxiesOverride(v)

	if entries := v.GetStringSlice("security.trusted_proxies"); len(entries) != 0 {
		t.Errorf("trusted_proxies = %v, 期望空列表表示不信任任何代理", entries)
	}
}

// TestApplyTrustedProxiesOverrideKeepsYAMLWhenEnvMissing 验证未注入环境变量时保持配置文件取值。
func TestApplyTrustedProxiesOverrideKeepsYAMLWhenEnvMissing(t *testing.T) {
	if _, ok := os.LookupEnv(trustedProxiesEnvKey); ok {
		t.Skipf("环境已存在 %s，跳过未注入分支", trustedProxiesEnvKey)
	}

	v := viper.New()
	v.Set("security.trusted_proxies", []string{"127.0.0.1/32"})

	applyTrustedProxiesOverride(v)

	entries := v.GetStringSlice("security.trusted_proxies")
	if len(entries) != 1 || entries[0] != "127.0.0.1/32" {
		t.Errorf("trusted_proxies = %v, 期望保持配置文件取值", entries)
	}
}

// TestApplyEnvOverridesSkipsStringSlice 验证字符串列表配置项不被通用覆盖改写，
// 列表类配置只能经各自的专用通道注入。
func TestApplyEnvOverridesSkipsStringSlice(t *testing.T) {
	t.Setenv("MYBLOG_I18N_SUPPORTED_LANGUAGES", "zh,en,fr")

	v := viper.New()
	v.Set("i18n.supported_languages", []string{"zh", "en"})

	applyEnvOverrides(v)

	if entries := v.GetStringSlice("i18n.supported_languages"); len(entries) != 2 {
		t.Errorf("supported_languages = %v, 期望保持配置文件原值", entries)
	}
}

// TestTrustedProxiesEnvOverrideReachesConfigStruct 验证环境注入的信任链取值贯通到配置结构体，
// 该路径是容器编排生效的唯一通道。
func TestTrustedProxiesEnvOverrideReachesConfigStruct(t *testing.T) {
	t.Setenv(trustedProxiesEnvKey, "172.16.0.0/12")

	// 复刻 Load 的取值顺序：先落默认值，再应用专用环境通道，最后解析到结构体。
	v := viper.New()
	v.SetDefault("security.trusted_proxies", defaultTrustedProxies)
	applyTrustedProxiesOverride(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatalf("解析配置失败: %v", err)
	}

	entries := cfg.Security.TrustedProxies
	if len(entries) != 1 || entries[0] != "172.16.0.0/12" {
		t.Errorf("Security.TrustedProxies = %v, 期望解析出环境注入的单个网段", entries)
	}
}
