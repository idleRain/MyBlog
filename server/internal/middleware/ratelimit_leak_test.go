package middleware

import (
	"runtime"
	"testing"
	"time"
)

// TestNoCleanupGoroutineLeak 验证反复构造限流器不再累积常驻协程。
func TestNoCleanupGoroutineLeak(t *testing.T) {
	runtime.GC()
	before := runtime.NumGoroutine()
	for i := 0; i < 50; i++ {
		limiter := NewRateLimiter(10, time.Minute)
		limiter.Allow("192.0.2.1")
		limiter.Allow("192.0.2.2")
	}
	time.Sleep(50 * time.Millisecond)
	runtime.GC()
	after := runtime.NumGoroutine()
	if after > before+2 {
		t.Fatalf("构造 50 个限流器后 goroutine 由 %d 增至 %d，存在泄漏", before, after)
	}
}
