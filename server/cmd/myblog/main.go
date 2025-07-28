// MyBlog 博客系统主程序
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"MyBlog/internal/config"
	"MyBlog/internal/database"
	"MyBlog/internal/handler"
	"MyBlog/internal/repository"
	"MyBlog/internal/router"
	"MyBlog/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal("配置加载失败:", err)
	}

	// 设置Gin运行模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	db, err := database.InitMySQL(cfg)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 运行数据库迁移（仅在生产环境）
	if cfg.Server.Mode != "debug" {
		if err := database.RunMigrations(cfg); err != nil {
			log.Fatal("数据库迁移失败:", err)
		}
	} else {
		log.Println("开发模式：跳过golang-migrate迁移，使用GORM AutoMigrate")
		// 开发模式使用GORM自动迁移
		if err := database.AutoMigrateWithFix(db); err != nil {
			log.Fatal("GORM自动迁移失败:", err)
		}
	}

	// 初始化依赖注入
	userRepo := repository.NewUserRepository(db)
	jwtService := service.NewJWTService(cfg)
	userSvc := service.NewUserService(userRepo, jwtService)
	userHandler := handler.NewUserHandler(userSvc)

	// 创建路由管理器
	routerManager := router.NewRouter()

	// 设置依赖
	deps := &router.Dependencies{
		UserHandler:    userHandler,
		JWTService:     jwtService,
		UserRepository: userRepo,
	}

	// 注册路由
	routerManager.SetupRoutes(deps)

	// 可选：注册 V2 版本路由
	// routerManager.SetupV2Routes(deps)

	// 获取 Gin 引擎
	engine := routerManager.GetEngine()

	// 启动服务器
	log.Printf("服务器启动成功，监听地址: %s", cfg.GetServerAddress())
	log.Printf("运行模式: %s", cfg.Server.Mode)

	// 优雅关闭
	go func() {
		if err := engine.Run(cfg.GetServerAddress()); err != nil {
			log.Fatal("服务器启动失败:", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 关闭数据库连接
	if err := database.Close(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}

	log.Println("服务器已关闭")
}
