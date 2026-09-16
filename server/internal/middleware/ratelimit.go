package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit 速率限制中间件
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(maxRequests, window)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiter 速率限制器
type RateLimiter struct {
	maxRequests int
	window      time.Duration
	clients     map[string]*clientInfo
	mutex       sync.RWMutex
}

type clientInfo struct {
	requests  int
	lastReset time.Time
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		clients:     make(map[string]*clientInfo),
	}

	// 启动清理协程
	go limiter.cleanup()

	return limiter
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientID]

	if !exists {
		rl.clients[clientID] = &clientInfo{
			requests:  1,
			lastReset: now,
		}
		return true
	}

	// 检查是否需要重置计数器
	if now.Sub(client.lastReset) >= rl.window {
		client.requests = 1
		client.lastReset = now
		return true
	}

	// 检查是否超过限制
	if client.requests >= rl.maxRequests {
		return false
	}

	client.requests++
	return true
}

// cleanup 清理过期的客户端信息
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()

		for clientID, client := range rl.clients {
			if now.Sub(client.lastReset) >= rl.window*2 {
				delete(rl.clients, clientID)
			}
		}

		rl.mutex.Unlock()
	}
}

// RateLimitPerUser 按用户限制速率
func RateLimitPerUser(maxRequests int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(maxRequests, window)

	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			// 如果没有用户ID，使用IP地址
			userID = c.ClientIP()
		}

		clientKey := getUserKey(userID)

		if !limiter.Allow(clientKey) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// userKeyPrefix 用户级限流键的统一前缀。
const userKeyPrefix = "user:"

// getUserKey 获取用户键，统一为十进制字符串形式。
// 历史 rune 转换对大数值 ID 会产生非法字符与键碰撞，限流配额因此串用。
func getUserKey(userID interface{}) string {
	switch v := userID.(type) {
	case string:
		return userKeyPrefix + v
	case uint:
		return userKeyPrefix + strconv.FormatUint(uint64(v), 10)
	default:
		return userKeyPrefix + "unknown"
	}
}
