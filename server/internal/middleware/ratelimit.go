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

// NewRateLimiter 创建新的速率限制器。
// 过期客户端随新客户端加入惰性清扫，不启动独立清扫协程，
// 避免无停止通道的常驻 goroutine 在测试反复构造限流器时累积泄漏。
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		clients:     make(map[string]*clientInfo),
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientID]

	if !exists {
		// 表只在新客户端加入时增长，清扫挂在增长点执行，频率与新客户端出现频率一致。
		rl.pruneExpiredLocked(now)
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

// pruneExpiredLocked 移除超过两个窗口未活动的客户端记录，约束表规模。
// 调用方必须已持有写锁，清扫仅由新客户端加入触发，无需独立协程。
func (rl *RateLimiter) pruneExpiredLocked(now time.Time) {
	for clientID, client := range rl.clients {
		if now.Sub(client.lastReset) >= rl.window*2 {
			delete(rl.clients, clientID)
		}
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
