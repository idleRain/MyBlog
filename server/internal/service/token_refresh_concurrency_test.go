package service

import (
	"sync"
	"testing"

	"MyBlog/internal/domain"
)

const (
	// refreshRaceRounds 并发兑换的重复轮数，逐轮更换刷新令牌以覆盖不同的调度交错。
	refreshRaceRounds = 300
	// refreshRaceGoroutines 单轮内争抢同一刷新令牌的并发调用方数量。
	refreshRaceGoroutines = 8
)

// TestRefreshAccessTokenConcurrentSingleUse 验证同一刷新令牌被并发兑换时恰有一次成功。
// 刷新即旋转要求旧刷新令牌单次使用，多于一次成功即意味着同一凭据换出了多份新令牌对。
func TestRefreshAccessTokenConcurrentSingleUse(t *testing.T) {
	svc := newTestTokenService(t)
	user := &domain.User{ID: 21}

	for round := 0; round < refreshRaceRounds; round++ {
		pair, err := svc.GenerateTokenPair(user)
		if err != nil {
			t.Fatalf("第 %d 轮生成令牌对失败: %v", round, err)
		}

		// 所有调用方阻塞在同一闸门上，令它们尽可能同时进入兑换流程。
		startGate := make(chan struct{})
		grantedPairs := make(chan *TokenPair, refreshRaceGoroutines)

		var wg sync.WaitGroup
		wg.Add(refreshRaceGoroutines)
		for i := 0; i < refreshRaceGoroutines; i++ {
			go func() {
				defer wg.Done()
				<-startGate
				newPair, err := svc.RefreshAccessToken(pair.RefreshToken)
				if err == nil {
					grantedPairs <- newPair
				}
			}()
		}

		close(startGate)
		wg.Wait()
		close(grantedPairs)

		successCount := 0
		for range grantedPairs {
			successCount++
		}
		if successCount != 1 {
			t.Fatalf("第 %d 轮同一刷新令牌兑换成功 %d 次，期望 1 次", round, successCount)
		}
	}
}
