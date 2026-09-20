package handler

import (
	"MyBlog/internal/domain"

	"github.com/gin-gonic/gin"
)

// ContentLanguageHeader 响应头名称，标注本次响应实际输出的内容语言。
const ContentLanguageHeader = "Content-Language"

// getOperatorID 从 gin 上下文读取当前操作者用户 ID。
// 上下文由认证中间件注入 userID，读取失败时返回 false 由调用方统一处理。
func getOperatorID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("userID")
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint)
	return userID, ok
}

// getOptionalUserID 从 gin 上下文读取当前用户 ID，未登录时返回 nil。
// 上下文由可选认证中间件注入 userID，结果交由服务层做角色化可见性判断。
func getOptionalUserID(c *gin.Context) *uint {
	value, exists := c.Get("userID")
	if !exists {
		return nil
	}

	userID, ok := value.(uint)
	if !ok {
		return nil
	}
	return &userID
}

// getRequestLanguage 读取语言中间件解析的语言标识，上下文缺失时回退缺省语言。
func getRequestLanguage(c *gin.Context) domain.Language {
	value, exists := c.Get(domain.LanguageContextKey)
	if !exists {
		return domain.DefaultLanguage
	}

	language, ok := value.(domain.Language)
	if !ok {
		return domain.DefaultLanguage
	}
	return language
}
