// Package logger 提供结构化日志的初始化、请求标识传递与标准库日志桥接。
// 日志统一输出为 JSON 行，请求级记录经上下文携带请求标识，用于串联同一次请求的全部日志。
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 输出目标，取自 logger.output 配置项。
const (
	OutputStdout = "stdout"
	OutputFile   = "file"
)

// 级别名，取自 logger.level 配置项。
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

// 日志字段键。访问日志、错误日志与抓取端的查询表达式共用同一套键名，
// 集中声明以避免同一语义出现多种拼写。
const (
	FieldRequestID = "requestId"
	FieldUserID    = "userId"
	FieldClientIP  = "clientIp"
	FieldMethod    = "method"
	FieldRoute     = "route"
	FieldPath      = "path"
	FieldStatus    = "status"
	FieldLatencyMs = "latencyMs"
	FieldUserAgent = "userAgent"
	FieldError     = "error"
)

// requestIDContextKey 请求标识在 gin 上下文中的键，由本包独占拥有。
// 经 gin 上下文传递而非 context.Context，可在不带 *http.Request 的测试上下文中同样可用。
const requestIDContextKey = "requestID"

// Options 日志初始化参数。
type Options struct {
	Level    string    // 级别名，取值 debug / info / warn / error
	Output   string    // 输出目标，取值 stdout / file
	FilePath string    // Output 为 file 时的目标文件，缺失的目录不会自动创建
	Writer   io.Writer // 写入目标覆盖项，供测试注入；非空时忽略 Output 与 FilePath
}

var (
	loggerMu      sync.RWMutex
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logCloser     io.Closer
)

// Init 按配置初始化默认日志器，并把标准库 log 的输出重定向到同一处理器。
// 该函数在服务启动阶段调用一次，须早于任何日志写入；重复调用会释放先前的文件输出目标。
func Init(opts Options) error {
	writer, closer, err := outputWriter(opts)
	if err != nil {
		return err
	}
	level, err := parseLevel(opts.Level)
	if err != nil {
		return err
	}

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	configured := slog.New(handler)

	loggerMu.Lock()
	previousCloser := logCloser
	logCloser = closer
	defaultLogger = configured
	loggerMu.Unlock()

	if previousCloser != nil {
		if err := previousCloser.Close(); err != nil {
			return fmt.Errorf("关闭先前的日志文件失败: %w", err)
		}
	}

	slog.SetDefault(configured)
	bridgeStdlibLog(handler)
	bridgeGinOutput(handler)
	return nil
}

// Close 释放文件输出目标，未使用文件输出时为空操作。
// 调用后不应再写入日志，故须在进程退出前最后调用。
func Close() error {
	loggerMu.Lock()
	closer := logCloser
	logCloser = nil
	loggerMu.Unlock()

	if closer == nil {
		return nil
	}
	if err := closer.Close(); err != nil {
		return fmt.Errorf("关闭日志文件失败: %w", err)
	}
	return nil
}

// L 返回默认结构化日志器。
func L() *slog.Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return defaultLogger
}

// AttachRequestID 把请求标识写入 gin 上下文，供日志与错误处理读取。
func AttachRequestID(c *gin.Context, requestID string) {
	c.Set(requestIDContextKey, requestID)
}

// FromGin 返回带该请求标识的结构化日志器。
// 上下文缺少标识或标识类型不符时退化为默认日志器，日志写入不因标识缺失而失败。
func FromGin(c *gin.Context) *slog.Logger {
	value, exists := c.Get(requestIDContextKey)
	if !exists {
		return L()
	}
	requestID, ok := value.(string)
	if !ok || requestID == "" {
		return L()
	}
	return L().With(slog.String(FieldRequestID, requestID))
}

// outputWriter 按输出配置返回日志写入目标及其释放函数。
// 释放函数在切换到其他输出目标或进程退出时调用；标准输出与测试注入目标无需释放，返回 nil。
func outputWriter(opts Options) (io.Writer, io.Closer, error) {
	if opts.Writer != nil {
		return opts.Writer, nil, nil
	}

	switch opts.Output {
	case "", OutputStdout:
		return os.Stdout, nil, nil
	case OutputFile:
		// 以追加方式打开以免重启覆盖历史记录；释放由 Close 负责。
		file, err := os.OpenFile(opts.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("打开日志文件 %s 失败: %w", opts.FilePath, err)
		}
		return file, file, nil
	default:
		return nil, nil, fmt.Errorf("不支持的日志输出目标 %q", opts.Output)
	}
}

// parseLevel 把配置中的级别名转换为 slog 级别。
// 未知取值返回错误而非静默回退，避免级别配置的拼写错误让日志长期低于预期。
func parseLevel(name string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case LevelDebug:
		return slog.LevelDebug, nil
	case "", LevelInfo:
		return slog.LevelInfo, nil
	case LevelWarn:
		return slog.LevelWarn, nil
	case LevelError:
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("不支持的日志级别 %q", name)
	}
}

// stdlibLogLevelMarkers 存量 log.Printf 的级别标记，按声明顺序匹配消息前缀。
// 数据库启动、迁移与媒体清理等后台路径沿用该约定，桥接时据此恢复级别，
// 避免兼容通道把全部记录压平为同一级别。
var stdlibLogLevelMarkers = []struct {
	marker string
	level  slog.Level
}{
	{"[ERROR]", slog.LevelError},
	{"[WARN]", slog.LevelWarn},
	{"[DEBUG]", slog.LevelDebug},
}

// stdlibLogBridge 把标准库 log 的输出转写为结构化记录。
type stdlibLogBridge struct {
	handler slog.Handler
}

// Write 实现 io.Writer，按消息前缀决定记录级别。
func (b stdlibLogBridge) Write(p []byte) (int, error) {
	message := strings.TrimRight(string(p), "\r\n")
	level := slog.LevelInfo
	for _, entry := range stdlibLogLevelMarkers {
		if strings.HasPrefix(message, entry.marker) {
			level = entry.level
			message = strings.TrimSpace(strings.TrimPrefix(message, entry.marker))
			break
		}
	}

	record := slog.NewRecord(time.Now(), level, message, 0)
	if err := b.handler.Handle(context.Background(), record); err != nil {
		return 0, err
	}
	return len(p), nil
}

// bridgeStdlibLog 把标准库 log 的默认输出重定向到结构化处理器，
// 使不经 slog 的既有调用点同样输出 JSON 行。
// 标准库自带的时间前缀一并关闭，时间戳由结构化记录提供，避免同一行出现两个时间。
func bridgeStdlibLog(handler slog.Handler) {
	log.SetFlags(0)
	log.SetOutput(stdlibLogBridge{handler: handler})
}

// bridgeGinOutput 把 gin 的调试与错误输出接入同一处理器，
// 避免启动横幅与路由清单以明文混入结构化日志流。
// gin 在 release 模式下不输出路由清单，该桥接主要覆盖开发模式。
func bridgeGinOutput(handler slog.Handler) {
	bridge := stdlibLogBridge{handler: handler}
	gin.DefaultWriter = bridge
	gin.DefaultErrorWriter = bridge
}
