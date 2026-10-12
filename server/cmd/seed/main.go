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
	withDemo := flag.Bool("demo", false, "同时写入演示数据：分类、标签、文章与评论")
	allowProduction := flag.Bool("force", false, "允许在 release 模式下写入演示数据")
	flag.Parse()

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	db, err := database.InitMySQL(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 确保种子数据涉及的表存在，再写入种子数据。
	if err := db.AutoMigrate(seedModels(*withDemo)...); err != nil {
		log.Fatalf("种子数据表迁移失败: %v", err)
	}

	admin, err := database.EnsureSuperAdmin(db, database.AdminSeedOptions{
		Username: *username,
		Password: *password,
		Email:    *email,
	})
	if err != nil {
		log.Fatalf("初始化超级管理员失败: %v", err)
	}

	if !*withDemo {
		return
	}

	// 演示数据以超级管理员为作者；release 模式下由 SeedDemoData 拒绝写入，除非显式放行。
	err = database.SeedDemoData(db, database.DemoSeedOptions{
		AuthorID:        admin.ID,
		ServerMode:      cfg.Server.Mode,
		AllowProduction: *allowProduction,
	})
	if err != nil {
		log.Fatalf("写入演示数据失败: %v", err)
	}
}

// seedModels 返回种子数据涉及的表模型。
// 演示数据额外涉及分类、标签、文章与评论；多对多关联表 GORM 不会自动建表，必须显式登记。
func seedModels(withDemo bool) []interface{} {
	models := []interface{}{&model.User{}}
	if !withDemo {
		return models
	}

	return append(models,
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.ArticleCategory{},
		&model.ArticleTag{},
		&model.Comment{},
	)
}
