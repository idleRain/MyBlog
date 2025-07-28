package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"MyBlog/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CacheItem 缓存项结构
type CacheItem struct {
	Key       string      `bson:"_id"`
	Value     interface{} `bson:"value"`
	ExpiresAt time.Time   `bson:"expires_at"`
	CreatedAt time.Time   `bson:"created_at"`
}

// CacheService 缓存服务接口
type CacheService interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string, dest interface{}) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Clear() error
	// 批量操作
	SetMany(items map[string]interface{}, expiration time.Duration) error
	GetMany(keys []string) (map[string]interface{}, error)
	DeleteMany(keys []string) error
	// 过期清理
	CleanupExpired() error
}

// mongoCacheService MongoDB缓存服务实现
type mongoCacheService struct {
	collection *mongo.Collection
}

// NewMongoCacheService 创建MongoDB缓存服务实例
func NewMongoCacheService() CacheService {
	db := database.GetMongoDatabase()
	collection := db.Collection("cache")

	// 创建TTL索引，自动清理过期数据
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"expires_at": 1},
		Options: options.Index().SetExpireAfterSeconds(0),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		// 索引可能已存在，记录警告但不终止
		fmt.Printf("创建TTL索引警告: %v\n", err)
	}

	return &mongoCacheService{
		collection: collection,
	}
}

// Set 设置缓存项
func (c *mongoCacheService) Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	expiresAt := time.Now().Add(expiration)
	if expiration <= 0 {
		// 永不过期，设置为很远的未来时间
		expiresAt = time.Now().Add(100 * 365 * 24 * time.Hour)
	}

	item := CacheItem{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	opts := options.Replace().SetUpsert(true)
	_, err := c.collection.ReplaceOne(ctx, bson.M{"_id": key}, item, opts)
	return err
}

// Get 获取缓存项
func (c *mongoCacheService) Get(key string, dest interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var item CacheItem
	filter := bson.M{
		"_id":        key,
		"expires_at": bson.M{"$gt": time.Now()}, // 确保未过期
	}

	err := c.collection.FindOne(ctx, filter).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("缓存键不存在或已过期: %s", key)
		}
		return err
	}

	// 将值反序列化到目标对象
	jsonData, err := json.Marshal(item.Value)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, dest)
}

// Delete 删除缓存项
func (c *mongoCacheService) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.collection.DeleteOne(ctx, bson.M{"_id": key})
	return err
}

// Exists 检查缓存项是否存在
func (c *mongoCacheService) Exists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":        key,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	count, err := c.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Clear 清空所有缓存
func (c *mongoCacheService) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.collection.DeleteMany(ctx, bson.M{})
	return err
}

// SetMany 批量设置缓存项
func (c *mongoCacheService) SetMany(items map[string]interface{}, expiration time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expiresAt := time.Now().Add(expiration)
	if expiration <= 0 {
		expiresAt = time.Now().Add(100 * 365 * 24 * time.Hour)
	}

	var operations []mongo.WriteModel
	for key, value := range items {
		item := CacheItem{
			Key:       key,
			Value:     value,
			ExpiresAt: expiresAt,
			CreatedAt: time.Now(),
		}

		operation := mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": key}).
			SetReplacement(item).
			SetUpsert(true)

		operations = append(operations, operation)
	}

	_, err := c.collection.BulkWrite(ctx, operations)
	return err
}

// GetMany 批量获取缓存项
func (c *mongoCacheService) GetMany(keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":        bson.M{"$in": keys},
		"expires_at": bson.M{"$gt": time.Now()},
	}

	cursor, err := c.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]interface{})
	for cursor.Next(ctx) {
		var item CacheItem
		if err := cursor.Decode(&item); err != nil {
			continue
		}
		result[item.Key] = item.Value
	}

	return result, cursor.Err()
}

// DeleteMany 批量删除缓存项
func (c *mongoCacheService) DeleteMany(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": bson.M{"$in": keys}}
	_, err := c.collection.DeleteMany(ctx, filter)
	return err
}

// CleanupExpired 清理过期的缓存项（虽然有TTL索引，但可以手动触发清理）
func (c *mongoCacheService) CleanupExpired() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"expires_at": bson.M{"$lte": time.Now()}}
	result, err := c.collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	fmt.Printf("清理了 %d 个过期缓存项\n", result.DeletedCount)
	return nil
}

// 缓存键名常量
const (
	// 用户相关缓存
	CacheKeyUserInfo  = "user:info:%d"  // 用户信息
	CacheKeyUserToken = "user:token:%d" // 用户令牌
	CacheKeyUserPerm  = "user:perm:%d"  // 用户权限

	// 文章相关缓存
	CacheKeyArticle     = "article:%d"      // 文章详情
	CacheKeyArticleList = "article:list:%s" // 文章列表
	CacheKeyHotArticles = "article:hot"     // 热门文章

	// 系统配置缓存
	CacheKeySettings = "settings"      // 系统设置
	CacheKeyCategory = "category:tree" // 分类树
	CacheKeyTags     = "tags:all"      // 所有标签

	// 统计数据缓存
	CacheKeyStats     = "stats:site" // 站点统计
	CacheKeyViewCount = "view:%d"    // 浏览量计数
)

// GetCacheKey 生成缓存键
func GetCacheKey(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}
