// Package response 提供统一的API响应格式
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`           // 响应码
	Message string      `json:"message"`        // 响应消息
	Data    interface{} `json:"data,omitempty"` // 响应数据
}

// 预定义响应码
const (
	CodeSuccess  = 200 // 成功
	CodeError    = 500 // 服务器内部错误
	CodeInvalid  = 400 // 请求参数错误
	CodeAuth     = 401 // 认证失败
	CodeForbid   = 403 // 权限不足
	CodeNotFound = 404 // 资源不存在
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "操作成功",
		Data:    data,
	})
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, message string) {
	Error(c, CodeInvalid, message)
}

// InternalError 服务器内部错误
func InternalError(c *gin.Context, message string) {
	Error(c, CodeError, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	Error(c, CodeAuth, message)
}

// Forbidden 权限不足
func Forbidden(c *gin.Context, message string) {
	Error(c, CodeForbid, message)
}
