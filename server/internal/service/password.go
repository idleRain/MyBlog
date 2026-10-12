// Package service 业务逻辑层
package service

import (
	"fmt"
	"regexp"
)

// 密码长度上下限，与前端注册与改密表单的提示文案保持一致，任一侧调整须同步。
const (
	passwordMinLength = 6
	passwordMaxLength = 100
)

// 密码字符构成检查，编译一次复用，避免每次校验重复编译正则。
var (
	passwordDigitPattern  = regexp.MustCompile(`[0-9]`)
	passwordLetterPattern = regexp.MustCompile(`[a-zA-Z]`)
)

// weakPasswords 常见弱密码黑名单，命中即拒绝，防止爆破字典的首批口令被直接采用。
var weakPasswords = []string{
	"123456", "password", "123456789", "12345678", "12345",
	"1234567", "1234567890", "qwerty", "abc123", "admin",
	"root", "guest", "test", "user", "demo", "login",
}

// ValidatePasswordStrength 校验密码强度，供创建用户、更新用户与修改密码三处复用。
// 校验失败统一包装 ErrInvalidRequest，使 handler 分档为 400 并保留具体原因，
// 避免裸错误退化为 500 脱敏文案而丢失可修正的提示。
func ValidatePasswordStrength(password string) error {
	if len(password) < passwordMinLength {
		return fmt.Errorf("%w：密码长度不能少于 %d 位", ErrInvalidRequest, passwordMinLength)
	}
	if len(password) > passwordMaxLength {
		return fmt.Errorf("%w：密码长度不能超过 %d 位", ErrInvalidRequest, passwordMaxLength)
	}

	if !passwordDigitPattern.MatchString(password) || !passwordLetterPattern.MatchString(password) {
		return fmt.Errorf("%w：密码必须包含字母和数字", ErrInvalidRequest)
	}

	for _, weak := range weakPasswords {
		if password == weak {
			return fmt.Errorf("%w：密码过于简单，请使用更复杂的密码", ErrInvalidRequest)
		}
	}

	return nil
}
