package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"MyBlog/internal/config"
	"MyBlog/internal/model"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

// getMigrationsPath 获取迁移文件路径
func getMigrationsPath() (string, error) {
	// 可能的迁移文件路径
	possiblePaths := []string{
		"./migrations",      // 从 server 目录运行
		"../migrations",     // 从 tmp 目录运行
		"../../migrations",  // 从更深的子目录运行
		"server/migrations", // 从项目根目录运行
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, err := filepath.Abs(path)
			if err != nil {
				continue
			}

			// 处理Windows路径格式
			if runtime.GOOS == "windows" {
				// 将反斜杠转换为正斜杠
				absPath = strings.Replace(absPath, "\\", "/", -1)
				// Windows文件URL格式
				return fmt.Sprintf("file:///%s", absPath), nil
			}

			return fmt.Sprintf("file://%s", absPath), nil
		}
	}

	// 如果都找不到，返回默认路径
	return "file://./migrations", fmt.Errorf("未找到迁移文件目录")
}

// createMigrateInstance 创建migrate实例的通用函数
func createMigrateInstance(cfg *config.Config) (*migrate.Migrate, error) {
	// 连接数据库
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	// 创建 MySQL 驱动实例
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("创建MySQL驱动失败: %w", err)
	}

	// 简化路径处理 - 寻找迁移目录
	var migrationsPath string
	possiblePaths := []string{
		"./migrations",
		"../migrations",
		"../../migrations",
		"server/migrations",
	}

	found := false
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			migrationsPath = fmt.Sprintf("file://%s", path)
			found = true
			log.Printf("使用迁移路径: %s", path)
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("未找到迁移文件目录，请确保 migrations 目录存在")
	}

	// 创建migrate实例
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return nil, fmt.Errorf("创建migrate实例失败: %w", err)
	}

	return m, nil
}

// RunMigrations 运行数据库迁移
func RunMigrations(cfg *config.Config) error {
	// 创建migrate实例
	m, err := createMigrateInstance(cfg)
	if err != nil {
		return err
	}
	defer m.Close()

	// 运行迁移
	log.Println("开始运行数据库迁移...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("运行迁移失败: %w", err)
	}

	// 获取当前版本
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取迁移版本失败: %w", err)
	}

	if err == migrate.ErrNilVersion {
		log.Println("数据库迁移完成，当前版本: 无")
	} else {
		status := "clean"
		if dirty {
			status = "dirty"
		}
		log.Printf("数据库迁移完成，当前版本: %d (%s)", version, status)
	}

	return nil
}

// MigrateDown 回滚数据库迁移
func MigrateDown(cfg *config.Config, steps int) error {
	// 连接数据库
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	// 创建 MySQL 驱动实例
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("创建MySQL驱动失败: %w", err)
	}

	// 获取迁移文件路径
	migrationsPath := "file://./migrations"

	// 创建migrate实例
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return fmt.Errorf("创建migrate实例失败: %w", err)
	}
	defer m.Close()

	// 回滚迁移
	log.Printf("开始回滚数据库迁移 %d 步...", steps)
	if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("回滚迁移失败: %w", err)
	}

	// 获取当前版本
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("获取迁移版本失败: %w", err)
	}

	if err == migrate.ErrNilVersion {
		log.Println("数据库迁移回滚完成，当前版本: 无")
	} else {
		status := "clean"
		if dirty {
			status = "dirty"
		}
		log.Printf("数据库迁移回滚完成，当前版本: %d (%s)", version, status)
	}

	return nil
}

// MigrateToVersion 迁移到指定版本
func MigrateToVersion(cfg *config.Config, version uint) error {
	// 连接数据库
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	// 创建 MySQL 驱动实例
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("创建MySQL驱动失败: %w", err)
	}

	// 获取迁移文件路径
	migrationsPath := "file://./migrations"

	// 创建migrate实例
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return fmt.Errorf("创建migrate实例失败: %w", err)
	}
	defer m.Close()

	// 迁移到指定版本
	log.Printf("开始迁移到版本 %d...", version)
	if err := m.Migrate(version); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("迁移到版本 %d 失败: %w", version, err)
	}

	log.Printf("成功迁移到版本 %d", version)
	return nil
}

// GetMigrationVersion 获取当前迁移版本
func GetMigrationVersion(cfg *config.Config) (uint, bool, error) {
	// 连接数据库
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		return 0, false, fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	// 创建 MySQL 驱动实例
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("创建MySQL驱动失败: %w", err)
	}

	// 获取迁移文件路径
	migrationsPath := "file://./migrations"

	// 创建migrate实例
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return 0, false, fmt.Errorf("创建migrate实例失败: %w", err)
	}
	defer m.Close()

	// 获取当前版本
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("获取迁移版本失败: %w", err)
	}

	return version, dirty, nil
}

// ForceMigrationVersion 强制设置迁移版本（用于修复dirty状态）
func ForceMigrationVersion(cfg *config.Config, version int) error {
	// 连接数据库
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	// 创建 MySQL 驱动实例
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("创建MySQL驱动失败: %w", err)
	}

	// 获取迁移文件路径
	migrationsPath := "file://./migrations"

	// 创建migrate实例
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		return fmt.Errorf("创建migrate实例失败: %w", err)
	}
	defer m.Close()

	// 强制设置版本
	log.Printf("强制设置迁移版本为 %d...", version)
	if err := m.Force(version); err != nil {
		return fmt.Errorf("强制设置版本失败: %w", err)
	}

	log.Printf("成功强制设置迁移版本为 %d", version)
	return nil
}

// AutoMigrateWithFix 保留原有的GORM自动迁移功能作为备用方案
func AutoMigrateWithFix(db *gorm.DB) error {
	log.Println("开始GORM自动迁移（仅用于开发环境）...")

	// 使用新的模型结构进行迁移
	if err := model.AutoMigrate(db); err != nil {
		return fmt.Errorf("GORM自动迁移失败: %w", err)
	}
	log.Println("GORM自动迁移完成")

	return nil
}
