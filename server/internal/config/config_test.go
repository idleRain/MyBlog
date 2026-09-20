package config

import (
	"testing"

	"github.com/spf13/viper"
)

// TestEnvVariableName 验证配置键到环境变量名的映射规则。
func TestEnvVariableName(t *testing.T) {
	cases := []struct {
		key      string
		expected string
	}{
		{key: "jwt.access_secret", expected: "MYBLOG_JWT_ACCESS_SECRET"},
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
	t.Setenv("MYBLOG_JWT_ACCESS_EXPIRE", "30")
	t.Setenv("MYBLOG_JWT_ISSUER", "env-issuer")

	v := viper.New()
	v.SetDefault("jwt.access_expire", 15)
	v.SetDefault("jwt.issuer", "myblog")

	applyEnvOverrides(v)

	if actual := v.GetInt("jwt.access_expire"); actual != 30 {
		t.Errorf("access_expire = %d, 期望环境变量覆盖值 30", actual)
	}
	if actual := v.GetString("jwt.issuer"); actual != "env-issuer" {
		t.Errorf("issuer = %q, 期望环境变量覆盖值 env-issuer", actual)
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
		JWT: JWTConfig{
			AccessSecret:  "unit-test-access-secret",
			RefreshSecret: "unit-test-refresh-secret",
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

// TestValidateConfigRejectsEmptySecrets 验证 JWT 密钥缺失即启动失败。
func TestValidateConfigRejectsEmptySecrets(t *testing.T) {
	cfg := validTestConfig()
	cfg.JWT.AccessSecret = ""
	if err := validateConfig(cfg); err == nil {
		t.Error("访问令牌密钥为空应返回错误")
	}

	cfg = validTestConfig()
	cfg.JWT.RefreshSecret = ""
	if err := validateConfig(cfg); err == nil {
		t.Error("刷新令牌密钥为空应返回错误")
	}
}

// TestValidateConfigRejectsWeakSecrets 验证已公开的弱默认密钥被拒绝。
func TestValidateConfigRejectsWeakSecrets(t *testing.T) {
	for _, weak := range knownWeakSecrets {
		cfg := validTestConfig()
		cfg.JWT.AccessSecret = weak
		if err := validateConfig(cfg); err == nil {
			t.Errorf("访问令牌密钥使用弱默认值 %q 应返回错误", weak)
		}

		cfg = validTestConfig()
		cfg.JWT.RefreshSecret = weak
		if err := validateConfig(cfg); err == nil {
			t.Errorf("刷新令牌密钥使用弱默认值 %q 应返回错误", weak)
		}
	}
}

// TestValidateConfigRejectsIdenticalSecrets 验证访问与刷新密钥互异校验。
func TestValidateConfigRejectsIdenticalSecrets(t *testing.T) {
	cfg := validTestConfig()
	cfg.JWT.RefreshSecret = cfg.JWT.AccessSecret
	if err := validateConfig(cfg); err == nil {
		t.Error("访问与刷新密钥相同应返回错误")
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

// TestLoadParsesConfigSections 验证 config.yaml 的 rbac 与 i18n 节可被正确解析并通过校验。
// YAML 中的 JWT 密钥为空值，经环境变量注入后应通过环境覆盖链路生效。
func TestLoadParsesConfigSections(t *testing.T) {
	t.Setenv("MYBLOG_JWT_ACCESS_SECRET", "config-test-access-secret")
	t.Setenv("MYBLOG_JWT_REFRESH_SECRET", "config-test-refresh-secret")

	cfg, err := Load("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.JWT.AccessSecret != "config-test-access-secret" {
		t.Errorf("AccessSecret = %q, 期望经环境变量覆盖为 config-test-access-secret", cfg.JWT.AccessSecret)
	}
	if cfg.JWT.RefreshSecret != "config-test-refresh-secret" {
		t.Errorf("RefreshSecret = %q, 期望经环境变量覆盖为 config-test-refresh-secret", cfg.JWT.RefreshSecret)
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
}
