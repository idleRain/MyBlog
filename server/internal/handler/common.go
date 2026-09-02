package handler

import (
	"github.com/gin-gonic/gin"
)

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
