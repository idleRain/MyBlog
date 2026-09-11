// MyBlog 博客系统主程序：组合根入口，负责配置加载、数据库初始化与 HTTP 生命周期管理。
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MyBlog/internal/config"
	"MyBlog/internal/database"
	"MyBlog/internal/router"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 优雅关停的排空时限，超时后强制退出。
const shutdownTimeout = 10 * time.Second

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal("配置加载失败:", err)
	}

	// 设置Gin运行模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	db := initDatabase(cfg)

	// 按业务域装配依赖并注册路由
	deps := newDependencies(cfg, db)
	routerManager := router.NewRouter(cfg)
	routerManager.SetupRoutes(deps)

	// 以结构化 HTTP 服务启动，支持优雅关停。
	server := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: routerManager.GetEngine(),
	}

	// 独立 goroutine 监听端口，主 goroutine 等待中断信号。
	go func() {
		log.Printf("服务器启动成功，监听地址: %s", cfg.GetServerAddress())
		log.Printf("运行模式: %s", cfg.Server.Mode)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号后进入优雅关停，排空在途请求。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器关停异常: %v", err)
	}

	// 关闭数据库连接
	if err := database.Close(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}

	log.Println("服务器已关闭")
}

// initDatabase 初始化数据库连接，并确保结构迁移与全文索引就绪。
func initDatabase(cfg *config.Config) *gorm.DB {
	db, err := database.InitMySQL(cfg)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 运行数据库迁移：生产环境走 golang-migrate，开发环境走 GORM AutoMigrate。
	if cfg.Server.Mode != "debug" {
		if err := database.RunMigrations(cfg); err != nil {
			log.Fatal("数据库迁移失败:", err)
		}
	} else {
		log.Println("开发模式：跳过golang-migrate迁移，使用GORM AutoMigrate")
		if err := database.AutoMigrateWithFix(db); err != nil {
			log.Fatal("GORM自动迁移失败:", err)
		}
	}

	// 确保文章全文索引存在，GORM AutoMigrate 无法声明 FULLTEXT 索引。
	if err := database.EnsureArticleFulltextIndex(db); err != nil {
		log.Fatal("创建全文索引失败:", err)
	}

	return db
}
