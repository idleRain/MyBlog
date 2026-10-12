package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"MyBlog/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// requestIDHeader 请求标识的传递头，服务端生成的标识同样经该头回写。
const requestIDHeader = "X-Request-ID"

// inboundRequestIDPattern 允许复用的调用方请求标识形态，长度上限 64 且仅含安全字符。
// 该取值会进入日志与响应头，不限形态时会被超长串或控制字符污染，并放大下游聚合的标签基数。
var inboundRequestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// requestLogSkipPaths 不写入访问日志的路径，取精确匹配，与 gin 既有的 SkipPaths 语义一致。
var requestLogSkipPaths = []string{"/api/health"}

// Logger 结构化访问日志中间件，每次请求输出一行 JSON。
// 记录关联请求标识与用户标识，同一次请求的全部日志可据此串联。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		if skipRequestLog(c.Request.URL.Path) {
			return
		}

		status := c.Writer.Status()
		attributes := []slog.Attr{
			slog.String(logger.FieldClientIP, c.ClientIP()),
			slog.String(logger.FieldMethod, c.Request.Method),
			slog.String(logger.FieldRoute, c.FullPath()),
			slog.String(logger.FieldPath, c.Request.URL.Path),
			slog.Int(logger.FieldStatus, status),
			slog.Int64(logger.FieldLatencyMs, time.Since(start).Milliseconds()),
			slog.String(logger.FieldUserAgent, c.Request.UserAgent()),
		}
		// 用户标识由认证中间件在本中间件内层写入，c.Next() 返回后必然已就绪。
		if userID, exists := c.Get("userID"); exists {
			attributes = append(attributes, slog.Any(logger.FieldUserID, userID))
		}
		if message := c.Errors.ByType(gin.ErrorTypePrivate).String(); message != "" {
			attributes = append(attributes, slog.String(logger.FieldError, message))
		}

		// 记录上下文取后台上下文：此处只写记录，请求上下文对处理器无额外语义，
		// 而测试构造的 gin 上下文不携带 *http.Request。
		logger.FromGin(c).LogAttrs(
			context.Background(), requestLogLevel(status), "请求处理完成", attributes...,
		)
	}
}

// requestLogLevel 按响应状态码决定访问日志级别。
func requestLogLevel(status int) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	case status >= http.StatusBadRequest:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

// skipRequestLog 判定路径是否跳过访问日志。
func skipRequestLog(path string) bool {
	for _, skipped := range requestLogSkipPaths {
		if path == skipped {
			return true
		}
	}
	return false
}

// Recovery panic 恢复中间件，保留 gin 的恢复语义并输出结构化记录。
// gin 默认恢复输出为明文行且不带请求标识，无法与同一次请求的其余日志关联；
// 写入目标传 nil 以关闭 gin 自带的转储，避免与下方结构化记录重复且丢失请求标识。
func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		logger.FromGin(c).Error("请求处理发生 panic",
			slog.Any(logger.FieldError, recovered),
			slog.String("stack", string(debug.Stack())),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// RequestID 请求标识中间件。
// 标识优先复用调用方传入的 X-Request-ID，取值不合法时生成 UUID v4；
// 标识写入响应头与 gin 上下文，日志据此串联同一次请求的全部记录。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := inboundRequestID(c.GetHeader(requestIDHeader))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Header(requestIDHeader, requestID)
		logger.AttachRequestID(c, requestID)
		c.Next()
	}
}

// inboundRequestID 校验调用方传入的请求标识，不合形态时返回空串由调用方重新生成。
func inboundRequestID(value string) string {
	if !inboundRequestIDPattern.MatchString(value) {
		return ""
	}
	return value
}
