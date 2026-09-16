// 认证 store 单飞刷新行为的单元测试。
import { describe, expect, it, vi } from 'vitest'
import { createAuthStore } from './auth-store'

// 构造依赖注入的最小替身，浏览器判定恒为 false 以隔离 localStorage。
function createTestStore() {
  return createAuthStore({
    isBrowser: () => false,
    logoutApi: async (_refreshToken?: string) => {}
  })
}

describe('refreshSingleFlight 单飞刷新', () => {
  it('并发触发刷新时共享同一次请求', async () => {
    const store = createTestStore()
    const refreshFn = vi.fn(async () => {
      // 引入异步间隙，模拟真实的网络刷新耗时窗口。
      await new Promise(resolve => setTimeout(resolve, 10))
      return 'new-access-token'
    })

    const [first, second] = await Promise.all([
      store.refreshSingleFlight(refreshFn),
      store.refreshSingleFlight(refreshFn)
    ])

    expect(refreshFn).toHaveBeenCalledTimes(1)
    expect(first).toBe('new-access-token')
    expect(second).toBe('new-access-token')
  })

  it('刷新完成后允许发起下一次刷新', async () => {
    const store = createTestStore()
    const refreshFn = vi.fn(async () => 'new-access-token')

    await store.refreshSingleFlight(refreshFn)
    await store.refreshSingleFlight(refreshFn)

    expect(refreshFn).toHaveBeenCalledTimes(2)
  })

  it('失败的刷新同样共享结果，结束后恢复可刷新状态', async () => {
    const store = createTestStore()
    const refreshFn = vi.fn(async () => null)

    const [first, second] = await Promise.all([
      store.refreshSingleFlight(refreshFn),
      store.refreshSingleFlight(refreshFn)
    ])

    expect(refreshFn).toHaveBeenCalledTimes(1)
    expect(first).toBeNull()
    expect(second).toBeNull()

    const third = await store.refreshSingleFlight(refreshFn)
    expect(third).toBeNull()
    expect(refreshFn).toHaveBeenCalledTimes(2)
  })
})
