package database

import (
	"fmt"
	"log"

	"MyBlog/internal/repository"

	"gorm.io/gorm"
)

// AutoMigrateWithFix 自动迁移数据库表结构并修复字段类型
func AutoMigrateWithFix(db *gorm.DB) error {
	log.Println("开始数据库表结构迁移...")

	// 定义需要迁移的模型
	models := []interface{}{
		&repository.User{},
		// 在这里添加其他模型
	}

	// 执行自动迁移
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("迁移模型 %T 失败: %w", model, err)
		}
		log.Printf("成功迁移模型: %T", model)
	}

	// 修复现有的日期时间字段类型（只在首次运行时需要）
	if err := fixDateTimeColumns(db); err != nil {
		log.Printf("修复日期时间字段警告: %v", err)
		// 不返回错误，因为可能是首次运行或已经修复过了
	}

	log.Println("数据库表结构迁移完成")
	return nil
}

// fixDateTimeColumns 修复现有的日期时间字段类型
func fixDateTimeColumns(db *gorm.DB) error {
	// 检查是否需要修复 users 表的字段类型
	var tableExists bool
	if err := db.Raw("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'users'").Scan(&tableExists).Error; err != nil {
		return err
	}

	if !tableExists {
		log.Println("users 表不存在，跳过字段类型修复")
		return nil
	}

	// 检查 created_at 字段类型
	var createdAtType string
	if err := db.Raw("SELECT DATA_TYPE FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'users' AND column_name = 'created_at'").Scan(&createdAtType).Error; err != nil {
		return err
	}

	// 如果字段类型是 date，则修复为 datetime
	if createdAtType == "date" {
		log.Println("检测到 created_at 字段类型为 date，正在修复为 datetime(3)...")
		if err := db.Exec("ALTER TABLE users MODIFY COLUMN created_at DATETIME(3)").Error; err != nil {
			return fmt.Errorf("修复 created_at 字段失败: %w", err)
		}
		log.Println("成功修复 created_at 字段类型")
	}

	// 检查 updated_at 字段类型
	var updatedAtType string
	if err := db.Raw("SELECT DATA_TYPE FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'users' AND column_name = 'updated_at'").Scan(&updatedAtType).Error; err != nil {
		return err
	}

	// 如果字段类型是 date，则修复为 datetime
	if updatedAtType == "date" {
		log.Println("检测到 updated_at 字段类型为 date，正在修复为 datetime(3)...")
		if err := db.Exec("ALTER TABLE users MODIFY COLUMN updated_at DATETIME(3)").Error; err != nil {
			return fmt.Errorf("修复 updated_at 字段失败: %w", err)
		}
		log.Println("成功修复 updated_at 字段类型")
	}

	return nil
}
