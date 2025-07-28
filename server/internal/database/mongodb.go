package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"MyBlog/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

// InitMongoDB 初始化MongoDB连接
func InitMongoDB(cfg *config.Config) error {
	// 构建MongoDB连接字符串
	mongoURI := buildMongoURI(cfg)

	// 设置客户端选项
	clientOptions := options.Client().ApplyURI(mongoURI)

	// 设置连接池配置
	clientOptions.SetMaxPoolSize(100)
	clientOptions.SetMinPoolSize(10)
	clientOptions.SetMaxConnIdleTime(30 * time.Minute)
	clientOptions.SetServerSelectionTimeout(5 * time.Second)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 连接到MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %w", err)
	}

	// 测试连接
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("MongoDB连接测试失败: %w", err)
	}

	MongoClient = client
	log.Println("MongoDB连接成功")

	return nil
}

// buildMongoURI 构建MongoDB连接URI
func buildMongoURI(cfg *config.Config) string {
	// 默认配置
	host := "localhost"
	port := "27017"
	database := "myblog_cache"
	username := ""
	password := ""
	authSource := "admin"

	// 从配置文件中读取（如果存在）
	if mongoConfig := cfg.Database.MongoDB; mongoConfig != nil {
		if h, ok := mongoConfig["host"].(string); ok && h != "" {
			host = h
		}
		if p, ok := mongoConfig["port"].(string); ok && p != "" {
			port = p
		}
		if d, ok := mongoConfig["database"].(string); ok && d != "" {
			database = d
		}
		if u, ok := mongoConfig["username"].(string); ok && u != "" {
			username = u
		}
		if pwd, ok := mongoConfig["password"].(string); ok && pwd != "" {
			password = pwd
		}
		if as, ok := mongoConfig["auth_source"].(string); ok && as != "" {
			authSource = as
		}
	}

	// 构建URI
	var uri string
	if username != "" && password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s",
			username, password, host, port, database, authSource)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s/%s", host, port, database)
	}

	return uri
}

// GetMongoDatabase 获取MongoDB数据库实例
func GetMongoDatabase() *mongo.Database {
	if MongoClient == nil {
		log.Fatal("MongoDB客户端未初始化")
	}
	return MongoClient.Database("myblog_cache")
}

// CloseMongoDB 关闭MongoDB连接
func CloseMongoDB() error {
	if MongoClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := MongoClient.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("关闭MongoDB连接失败: %w", err)
	}

	log.Println("MongoDB连接已关闭")
	return nil
}
