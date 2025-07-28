package model

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Setting 系统设置模型
type Setting struct {
	ID             uint        `json:"id" gorm:"primaryKey;comment:设置ID"`
	KeyName        string      `json:"keyName" gorm:"uniqueIndex;not null;size:100;comment:配置键名"`
	Value          string      `json:"value" gorm:"type:longtext;comment:配置值（支持JSON格式）"`
	DefaultValue   string      `json:"defaultValue" gorm:"type:longtext;comment:默认值"`
	Description    string      `json:"description" gorm:"size:255;comment:配置描述"`
	Type           SettingType `json:"type" gorm:"default:string;comment:值类型"`
	GroupName      string      `json:"groupName" gorm:"default:general;size:50;index;comment:配置分组"`
	IsPublic       bool        `json:"isPublic" gorm:"default:false;index;comment:是否公开（前端可访问）"`
	IsReadonly     bool        `json:"isReadonly" gorm:"default:false;comment:是否只读"`
	ValidationRule string      `json:"validationRule" gorm:"size:200;comment:验证规则"`
	SortOrder      int         `json:"sortOrder" gorm:"default:0;index;comment:排序权重"`
	CreatedAt      time.Time   `json:"createdAt" gorm:"type:datetime(3);comment:创建时间"`
	UpdatedAt      time.Time   `json:"updatedAt" gorm:"type:datetime(3);comment:更新时间"`
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}

// 定义设置类型枚举
type SettingType string

const (
	SettingTypeString  SettingType = "string"  // 字符串
	SettingTypeNumber  SettingType = "number"  // 数字
	SettingTypeBoolean SettingType = "boolean" // 布尔值
	SettingTypeJSON    SettingType = "json"    // JSON对象
	SettingTypeArray   SettingType = "array"   // 数组
)

// 定义常用的系统设置键名
const (
	// 网站基本信息
	SettingSiteName        = "site_name"
	SettingSiteDescription = "site_description"
	SettingSiteKeywords    = "site_keywords"
	SettingSiteAuthor      = "site_author"
	SettingSiteLogo        = "site_logo"
	SettingSiteFavicon     = "site_favicon"

	// SEO设置
	SettingSEOTitle       = "seo_title"
	SettingSEODescription = "seo_description"
	SettingSEOKeywords    = "seo_keywords"

	// 内容设置
	SettingArticlesPerPage    = "articles_per_page"
	SettingDefaultCategory    = "default_category"
	SettingAllowGuestComment  = "allow_guest_comment"
	SettingCommentEnabled     = "comment_enabled"
	SettingCommentAutoApprove = "comment_auto_approve"
	SettingCommentMaxDepth    = "comment_max_depth"

	// 媒体设置
	SettingUploadMaxSize     = "upload_max_size"
	SettingAllowedFileTypes  = "allowed_file_types"
	SettingImageQuality      = "image_quality"
	SettingThumbnailSize     = "thumbnail_size"
	SettingWatermarkEnabled  = "watermark_enabled"
	SettingWatermarkText     = "watermark_text"
	SettingWatermarkPosition = "watermark_position"
	SettingWatermarkOpacity  = "watermark_opacity"

	// 邮件设置
	SettingMailHost     = "mail_host"
	SettingMailPort     = "mail_port"
	SettingMailUsername = "mail_username"
	SettingMailPassword = "mail_password"
	SettingMailFrom     = "mail_from"
	SettingMailFromName = "mail_from_name"

	// 安全设置
	SettingEnableRateLimit   = "enable_rate_limit"
	SettingRateLimitPerHour  = "rate_limit_per_hour"
	SettingSessionTimeout    = "session_timeout"
	SettingEnableCaptcha     = "enable_captcha"
	SettingFailedLoginLimit  = "failed_login_limit"
	SettingPasswordMinLength = "password_min_length"

	// 缓存设置
	SettingCacheEnabled    = "cache_enabled"
	SettingCacheExpire     = "cache_expire"
	SettingMongoHost       = "mongo_host"
	SettingMongoPort       = "mongo_port"
	SettingMongoUsername   = "mongo_username"
	SettingMongoPassword   = "mongo_password"
	SettingMongoDatabase   = "mongo_database"
	SettingMongoAuthSource = "mongo_auth_source"

	// 第三方集成
	SettingAnalyticsCode   = "analytics_code"
	SettingDisqusShortname = "disqus_shortname"
	SettingGitalkClientID  = "gitalk_client_id"
	SettingGitalkClientKey = "gitalk_client_key"
	SettingSocialGithub    = "social_github"
	SettingSocialTwitter   = "social_twitter"
	SettingSocialWeibo     = "social_weibo"
	SettingSocialEmail     = "social_email"
)

// GetStringValue 获取字符串值
func (s *Setting) GetStringValue() string {
	if s.Value == "" {
		return s.DefaultValue
	}
	return s.Value
}

// GetIntValue 获取整数值
func (s *Setting) GetIntValue() int {
	value := s.GetStringValue()
	if value == "" {
		return 0
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		// 尝试解析默认值
		if s.DefaultValue != "" {
			if defaultInt, err := strconv.Atoi(s.DefaultValue); err == nil {
				return defaultInt
			}
		}
		return 0
	}
	return intVal
}

// GetBoolValue 获取布尔值
func (s *Setting) GetBoolValue() bool {
	value := s.GetStringValue()
	if value == "" {
		return false
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		// 兼容数字格式 (0/1)
		if value == "1" {
			return true
		}
		// 尝试解析默认值
		if s.DefaultValue != "" {
			if defaultBool, err := strconv.ParseBool(s.DefaultValue); err == nil {
				return defaultBool
			}
		}
		return false
	}
	return boolVal
}

// GetJSONValue 获取JSON值并解析到指定结构
func (s *Setting) GetJSONValue(v interface{}) error {
	value := s.GetStringValue()
	if value == "" {
		return json.Unmarshal([]byte("{}"), v)
	}

	return json.Unmarshal([]byte(value), v)
}

// GetArrayValue 获取数组值
func (s *Setting) GetArrayValue() ([]string, error) {
	var arr []string
	err := s.GetJSONValue(&arr)
	return arr, err
}

// SetStringValue 设置字符串值
func (s *Setting) SetStringValue(value string) {
	s.Value = value
	s.Type = SettingTypeString
}

// SetIntValue 设置整数值
func (s *Setting) SetIntValue(value int) {
	s.Value = strconv.Itoa(value)
	s.Type = SettingTypeNumber
}

// SetBoolValue 设置布尔值
func (s *Setting) SetBoolValue(value bool) {
	s.Value = strconv.FormatBool(value)
	s.Type = SettingTypeBoolean
}

// SetJSONValue 设置JSON值
func (s *Setting) SetJSONValue(value interface{}) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.Value = string(jsonBytes)
	s.Type = SettingTypeJSON
	return nil
}

// SetArrayValue 设置数组值
func (s *Setting) SetArrayValue(value []string) error {
	s.Type = SettingTypeArray
	return s.SetJSONValue(value)
}

// IsEditable 检查设置是否可编辑
func (s *Setting) IsEditable() bool {
	return !s.IsReadonly
}

// CanAccess 检查用户是否可以访问此设置
func (s *Setting) CanAccess(user *User) bool {
	// 公开设置所有人都可以访问
	if s.IsPublic {
		return true
	}

	// 私有设置需要管理员权限
	return user != nil && user.IsAdmin()
}

// CanEdit 检查用户是否可以编辑此设置
func (s *Setting) CanEdit(user *User) bool {
	// 只读设置不能编辑
	if s.IsReadonly {
		return false
	}

	// 需要管理员权限
	return user != nil && user.IsAdmin()
}

// Validate 验证设置值是否符合规则
func (s *Setting) Validate() error {
	// 这里可以根据 ValidationRule 字段实现具体的验证逻辑
	// 比如正则表达式验证、范围验证等

	// 基本类型验证
	switch s.Type {
	case SettingTypeNumber:
		_, err := strconv.Atoi(s.Value)
		if err != nil {
			return err
		}
	case SettingTypeBoolean:
		_, err := strconv.ParseBool(s.Value)
		if err != nil && s.Value != "0" && s.Value != "1" {
			return err
		}
	case SettingTypeJSON, SettingTypeArray:
		var temp interface{}
		return json.Unmarshal([]byte(s.Value), &temp)
	}

	return nil
}

// GetDisplayValue 获取用于显示的值（隐藏敏感信息）
func (s *Setting) GetDisplayValue() string {
	// 隐藏密码等敏感信息
	if s.IsSensitive() {
		if s.Value == "" {
			return ""
		}
		return "********"
	}
	return s.Value
}

// IsSensitive 检查是否为敏感设置
func (s *Setting) IsSensitive() bool {
	sensitiveKeys := []string{
		SettingMailPassword,
		SettingMongoPassword,
		"password",
		"secret",
		"key",
		"token",
	}

	keyLower := strings.ToLower(s.KeyName)
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(keyLower, sensitive) {
			return true
		}
	}
	return false
}
