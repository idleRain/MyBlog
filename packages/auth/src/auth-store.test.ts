// 认证 store 的单飞刷新行为与跨标签页同步单元测试。
import { AUTH_TOKEN_KEY, UNRELATED_KEY, installFakeBrowserEnv } from './test-fixtures'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { createAuthStore } from './auth-store'

// 构造依赖注入的最小替身，浏览器判定恒为 false 以隔离 localStorage。
function createTestStore() {
  return createAuthStore({
    isBrowser: () => false,
    logoutApi: async (_refreshToken?: string) => {}
  })
}

// 构造接入浏览器环境替身的 store，用于跨标签页同步相关用例。
function createBrowserStore() {
  return createAuthStore({
    isBrowser: () => true,
    logoutApi: async (_refreshToken?: string) => {}
  })
}

// 一份可用的测试会话，剩余有效期远大于提前刷新窗口。
const ACTIVE_SESSION = {
  accessToken: 'tab-a-access',
  refreshToken: 'tab-a-refresh',
  userId: 1,
  expiresInMs: 10 * 60 * 1000
}

afterEach(() => {
  vi.unstubAllGlobals()
})

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

describe('跨标签页令牌同步', () => {
  it('syncFromStorage 读到其他标签页写入的新令牌后更新内存状态', () => {
    const env = installFakeBrowserEnv()
    env.seedSession(ACTIVE_SESSION)
    const store = createBrowserStore()

    // 其他标签页完成旋转并把新令牌写入共享存储。
    env.seedSession({
      ...ACTIVE_SESSION,
      accessToken: 'tab-a-rotated',
      refreshToken: 'tab-a-rotated-r'
    })

    expect(store.getAccessToken()).toBe('tab-a-access')
    store.syncFromStorage()
    expect(store.getAccessToken()).toBe('tab-a-rotated')
    expect(store.getRefreshToken()).toBe('tab-a-rotated-r')
  })

  it('存储未发生变化时 syncFromStorage 不写入状态', () => {
    const env = installFakeBrowserEnv()
    env.seedSession(ACTIVE_SESSION)
    const store = createBrowserStore()

    const listener = vi.fn()
    store.subscribe(listener)
    // 订阅本身会立即触发一次，此处重置计数以只统计后续写入。
    listener.mockClear()

    store.syncFromStorage()

    expect(listener).not.toHaveBeenCalled()
  })

  it('storage 事件携带认证键时触发同步', () => {
    const env = installFakeBrowserEnv()
    env.seedSession(ACTIVE_SESSION)
    const store = createBrowserStore()

    env.seedSession({ ...ACTIVE_SESSION, accessToken: 'tab-a-rotated' })
    env.dispatchStorage(AUTH_TOKEN_KEY)

    expect(store.getAccessToken()).toBe('tab-a-rotated')
  })

  it('storage 事件携带与认证无关的键时不触发同步', () => {
    const env = installFakeBrowserEnv()
    env.seedSession(ACTIVE_SESSION)
    const store = createBrowserStore()

    env.seedSession({ ...ACTIVE_SESSION, accessToken: 'tab-a-rotated' })
    env.seedRaw(UNRELATED_KEY, 'dark')
    env.dispatchStorage(UNRELATED_KEY)

    expect(store.getAccessToken()).toBe('tab-a-access')
  })

  it('其他标签页登出后本页登录态一并失效', () => {
    const env = installFakeBrowserEnv()
    env.seedSession(ACTIVE_SESSION)
    const store = createBrowserStore()
    expect(store.getCurrentState().isAuthenticated).toBe(true)

    // 直接操作底层存储模拟其他标签页登出，本页内存状态不受影响。
    env.clearSession()
    env.dispatchStorage(null)

    expect(store.getCurrentState().isAuthenticated).toBe(false)
    expect(store.getAccessToken()).toBeNull()
  })
})
