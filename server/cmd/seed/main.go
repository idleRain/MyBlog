// MyBlog 数据库种子数据初始化工具
package main

import (
	"flag"
	"log"

	"MyBlog/internal/config"
	"MyBlog/internal/database"
	"MyBlog/internal/model"
)

// 超级管理员默认参数
const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "Admin@123456"
	defaultAdminEmail    = "admin@myblog.local"
)

func main() {
	username := flag.String("username", defaultAdminUsername, "超级管理员用户名")
	password := flag.String("password", defaultAdminPassword, "超级管理员密码")
	email := flag.String("email", defaultAdminEmail, "超级管理员邮箱")
	flag.Parse()

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	db, err := database.InitMySQL(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 确保用户表存在，再写入种子数据
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("用户表迁移失败: %v", err)
	}

	if err := database.EnsureSuperAdmin(db, database.AdminSeedOptions{
		Username: *username,
		Password: *password,
		Email:    *email,
	}); err != nil {
		log.Fatalf("初始化超级管理员失败: %v", err)
	}
}
