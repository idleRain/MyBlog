package middleware

import (
	"strconv"
	"testing"
	"time"
)

// TestRateLimiterBlocksBeyondThreshold 验证窗口内超过阈值的请求被拒绝。
func TestRateLimiterBlocksBeyondThreshold(t *testing.T) {
	const (
		maxRequests = 3
		window      = time.Minute
	)

	limiter := NewRateLimiter(maxRequests, window)

	for i := 0; i < maxRequests; i++ {
		if !limiter.Allow("client-a") {
			t.Fatalf("阈值内第 %d 次请求应放行", i+1)
		}
	}

	if limiter.Allow("client-a") {
		t.Fatal("超过阈值后的请求应被拒绝")
	}
}

// TestRateLimiterResetsAfterWindow 验证窗口流逝后计数器重置，请求重新放行。
func TestRateLimiterResetsAfterWindow(t *testing.T) {
	const (
		maxRequests = 2
		window      = 20 * time.Millisecond
	)

	limiter := NewRateLimiter(maxRequests, window)

	for i := 0; i < maxRequests; i++ {
		if !limiter.Allow("client-a") {
			t.Fatalf("窗口内第 %d 次请求应放行", i+1)
		}
	}
	if limiter.Allow("client-a") {
		t.Fatal("窗口内超阈值的请求应被拒绝")
	}

	// 等待一个完整窗口后计数器应重置，为避免调度抖动取窗口的三倍等待。
	time.Sleep(window * 3)

	if !limiter.Allow("client-a") {
		t.Fatal("窗口结束后请求应重新放行")
	}
}

// TestRateLimiterIsolatesClients 验证不同客户端键的配额相互独立。
func TestRateLimiterIsolatesClients(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)

	if !limiter.Allow("client-a") {
		t.Fatal("客户端 A 的首次请求应放行")
	}
	if limiter.Allow("client-a") {
		t.Fatal("客户端 A 超阈值的请求应被拒绝")
	}
	if !limiter.Allow("client-b") {
		t.Fatal("客户端 B 的配额不应被客户端 A 消耗")
	}
}

// TestGetUserKeyStrategy 验证用户级限流键的构造策略：不同类型与取值
// 产生稳定且互不碰撞的键，未知类型回退为固定键。
func TestGetUserKeyStrategy(t *testing.T) {
	testCases := []struct {
		name   string
		userID interface{}
		expect string
	}{
		{name: "字符串类型保留原值", userID: "42", expect: "user:42"},
		{name: "数值类型转为十进制字符串", userID: uint(42), expect: "user:42"},
		{name: "大数值 ID 不产生非法字符", userID: uint(9999999), expect: "user:9999999"},
		{name: "零值 ID 保持可辨识", userID: uint(0), expect: "user:0"},
		{name: "未知类型回退固定键", userID: 3.14, expect: "user:unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := getUserKey(tc.userID); got != tc.expect {
				t.Errorf("键 = %q, 期望 %q", got, tc.expect)
			}
		})
	}

	// 相邻大数值 ID 的键必须互异，历史 rune 转换在高位取值存在碰撞风险。
	if getUserKey(uint(100)) == getUserKey(uint(101)) {
		t.Error("不同用户 ID 的限流键发生碰撞")
	}
}

// TestRateLimiterConcurrentAllow 验证并发取用不产生数据竞争，配合 -race 运行。
func TestRateLimiterConcurrentAllow(t *testing.T) {
	limiter := NewRateLimiter(1000, time.Minute)

	done := make(chan struct{})
	for worker := 0; worker < 8; worker++ {
		go func(base int) {
			defer func() { done <- struct{}{} }()
			for i := 0; i < 100; i++ {
				_ = limiter.Allow("user:" + strconv.Itoa(base))
			}
		}(worker)
	}
	for worker := 0; worker < 8; worker++ {
		<-done
	}
}
