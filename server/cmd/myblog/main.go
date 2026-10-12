// MyBlog 博客系统主程序：组合根入口，负责配置加载、数据库初始化与 HTTP 生命周期管理。
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MyBlog/internal/config"
	"MyBlog/internal/database"
	"MyBlog/internal/logger"
	"MyBlog/internal/metrics"
	"MyBlog/internal/router"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HTTP 服务器超时与关停时限，防止慢速连接长期占用服务资源。
const (
	readHeaderTimeout = 5 * time.Second   // 请求头读取上限，阻断慢速 HTTP 攻击
	readTimeout       = 15 * time.Second  // 整个请求的读取上限，媒体上传上限 10MB 时 15s 足够
	writeTimeout      = 30 * time.Second  // 响应写出上限，覆盖媒体下载等大响应场景
	idleTimeout       = 120 * time.Second // 空闲连接保活上限，到期回收
	shutdownTimeout   = 10 * time.Second  // 优雅关停的排空时限，超时后强制退出
)

func main() {
	// 先以缺省输出初始化结构化日志，使配置加载失败同样产生可解析的记录。
	initLogging(logger.Options{Level: logger.LevelInfo, Output: logger.OutputStdout})

	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		exitWithError("配置加载失败", err)
	}

	// 配置就绪后按 logger 配置节重建日志输出。
	initLogging(logger.Options{
		Level:    cfg.Logger.Level,
		Output:   cfg.Logger.Output,
		FilePath: cfg.Logger.FilePath,
	})

	// 设置Gin运行模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	db := initDatabase(cfg)

	// 按业务域装配依赖并注册路由
	deps := newDependencies(cfg, db)
	metricsRegistry := metrics.New()
	routerManager := router.NewRouter(cfg, metricsRegistry.Middleware())
	routerManager.SetupRoutes(deps)

	// 指标监听独立于业务端口，未启用时返回空句柄。
	metricsServer, err := startMetricsServer(cfg, metricsRegistry)
	if err != nil {
		exitWithError("指标监听启动失败", err)
	}

	// 以结构化 HTTP 服务启动，支持优雅关停，四项超时防止连接资源泄漏。
	server := &http.Server{
		Addr:              cfg.GetServerAddress(),
		Handler:           routerManager.GetEngine(),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	// 独立 goroutine 监听端口，主 goroutine 等待中断信号。
	go func() {
		logger.L().Info("服务器启动成功",
			slog.String("addr", cfg.GetServerAddress()),
			slog.String("mode", cfg.Server.Mode),
		)
		// 监听失败必须终止进程：此时进程已无业务端口，继续运行只会表现为静默不可用。
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			exitWithError("服务器启动失败", err)
		}
	}()

	// 等待中断信号后进入优雅关停，排空在途请求。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.L().Info("正在关闭服务器")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.L().Error("服务器关停异常", slog.Any(logger.FieldError, err))
	}
	if metricsServer != nil {
		if err := metricsServer.Shutdown(ctx); err != nil {
			logger.L().Error("指标监听关停异常", slog.Any(logger.FieldError, err))
		}
	}

	// 关闭数据库连接
	if err := database.Close(); err != nil {
		logger.L().Error("关闭数据库连接失败", slog.Any(logger.FieldError, err))
	}

	logger.L().Info("服务器已关闭")

	// 释放日志输出目标，此后不再写日志；错误直接写标准错误以免经已关闭的目标写出。
	if err := logger.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "关闭日志输出失败: %v\n", err)
	}
}

// initLogging 初始化结构化日志，失败时无法继续记录日志，只能经标准库输出后退出。
func initLogging(opts logger.Options) {
	if err := logger.Init(opts); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}
}

// exitWithError 记录终止原因后释放日志输出并以失败码退出。
func exitWithError(message string, err error) {
	logger.L().Error(message, slog.Any(logger.FieldError, err))
	// 退出路径上的释放失败已无上报目标，忽略其错误。
	_ = logger.Close()
	os.Exit(1)
}

// initDatabase 初始化数据库连接，并确保结构迁移与全文索引就绪。
func initDatabase(cfg *config.Config) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		exitWithError("数据库初始化失败", err)
	}

	// 运行数据库迁移：生产环境走 golang-migrate，开发环境走 GORM AutoMigrate。
	if cfg.Server.Mode != "debug" {
		if err := database.RunMigrations(cfg); err != nil {
			exitWithError("数据库迁移失败", err)
		}
	} else {
		logger.L().Info("开发模式：跳过golang-migrate迁移，使用GORM AutoMigrate")
		if err := database.AutoMigrateWithFix(db); err != nil {
			exitWithError("GORM自动迁移失败", err)
		}
	}

	// 确保文章与其翻译表的全文索引存在，GORM AutoMigrate 无法声明 FULLTEXT 索引。
	if err := database.EnsureArticleFulltextIndex(db); err != nil {
		exitWithError("创建全文索引失败", err)
	}
	if err := database.EnsureArticleTranslationFulltextIndex(db); err != nil {
		exitWithError("创建翻译表全文索引失败", err)
	}

	// 字典种子数据幂等写入，开发与生产两条迁移路径执行后均保持初始字典可用。
	if err := database.SeedDicts(db); err != nil {
		exitWithError("字典种子数据写入失败", err)
	}

	return db
}
