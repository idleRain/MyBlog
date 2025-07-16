// Package database 提供数据库连接和管理功能
package database

import (
	"fmt"
	"log"
	"sync"
	"time"

	"MyBlog/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// InitMySQL 初始化MySQL数据库连接
func InitMySQL(cfg *config.Config) (*gorm.DB, error) {
	var err error

	once.Do(func() {
		// 配置GORM日志级别
		var logLevel logger.LogLevel
		switch cfg.Logger.Level {
		case "debug":
			logLevel = logger.Silent // 在debug模式下也不显示SQL查询
		case "info":
			logLevel = logger.Warn
		case "warn":
			logLevel = logger.Error
		default:
			logLevel = logger.Silent
		}

		// GORM配置
		gormConfig := &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
		}

		// 先创建数据库（如果不存在）
		if createErr := createDatabaseIfNotExists(cfg); createErr != nil {
			err = fmt.Errorf("创建数据库失败: %w", createErr)
			return
		}

		// 构建连接字符串（连接到具体数据库）
		dsn := cfg.GetDSN()

		// 建立数据库连接
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			err = fmt.Errorf("连接MySQL数据库失败: %w", err)
			return
		}

		// 获取底层sql.DB对象进行连接池配置
		sqlDB, sqlErr := db.DB()
		if sqlErr != nil {
			err = fmt.Errorf("获取数据库实例失败: %w", sqlErr)
			return
		}

		// 设置连接池参数
		sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)

		// 测试数据库连接
		if pingErr := sqlDB.Ping(); pingErr != nil {
			err = fmt.Errorf("数据库连接测试失败: %w", pingErr)
			return
		}

		log.Println("MySQL数据库连接成功")
	})

	return db, err
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	if db == nil {
		panic("数据库未初始化，请先调用 InitMySQL() 方法")
	}
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}

	log.Println("数据库连接已关闭")
	return nil
}

// createDatabaseIfNotExists 创建数据库（如果不存在）
func createDatabaseIfNotExists(cfg *config.Config) error {
	log.Printf("正在检查并创建数据库: %s", cfg.Database.DBName)

	// 构建不包含数据库名的连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=%t&loc=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Charset,
		cfg.Database.ParseTime,
		cfg.Database.Loc,
	)

	log.Printf("连接MySQL服务器: %s:%d", cfg.Database.Host, cfg.Database.Port)

	// 连接到MySQL服务器（不指定数据库）
	tempDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("连接MySQL服务器失败: %w", err)
	}

	// 确保临时连接最后关闭
	defer func() {
		if sqlDB, dbErr := tempDB.DB(); dbErr == nil {
			sqlDB.Close()
		}
	}()

	// 创建数据库
	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.Database.DBName)

	log.Printf("执行SQL: %s", createSQL)

	if err := tempDB.Exec(createSQL).Error; err != nil {
		return fmt.Errorf("创建数据库失败: %w", err)
	}

	log.Printf("数据库 %s 创建成功或已存在", cfg.Database.DBName)
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(models ...interface{}) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("数据库表结构迁移失败: %w", err)
	}

	log.Printf("数据库表结构迁移完成，共 %d 个模型", len(models))
	return nil
}

// HealthCheck 数据库健康检查
func HealthCheck() error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return nil
}
