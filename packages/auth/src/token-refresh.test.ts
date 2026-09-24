// 令牌刷新编排的单元测试。
// 重点覆盖刷新即旋转语义下同源多标签页共享刷新令牌所引入的竞态。
import { installFakeBrowserEnv, type FakeBrowserEnv, type SeededSession } from './test-fixtures'
import { createTokenRefresher, type TokenRefreshDeps } from './token-refresh'
import type { RefreshTokenData } from '@myblog/api/modules/user/types'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { createAuthStore } from './auth-store'

// 本页初始持有的会话，剩余有效期远大于提前刷新窗口。
const STALE_SESSION: SeededSession = {
  accessToken: 'stale-access',
  refreshToken: 'stale-refresh',
  userId: 1,
  expiresInMs: 10 * 60 * 1000
}

// 其他标签页旋转后写入的会话。
const ROTATED_SESSION: SeededSession = {
  accessToken: 'rotated-access',
  refreshToken: 'rotated-refresh',
  userId: 1,
  expiresInMs: 10 * 60 * 1000
}

const NEW_PAIR: RefreshTokenData = {
  accessToken: 'fresh-access',
  refreshToken: 'fresh-refresh',
  expiresIn: 900
}

// 刷新请求替身的签名，env 用于在请求过程中模拟其他标签页的落盘。
type RequestTokenPair = (
  refreshToken: string,
  env: FakeBrowserEnv
) => Promise<RefreshTokenData | null>

afterEach(() => {
  vi.unstubAllGlobals()
})

// 建立持有陈旧会话的 store 与配套刷新器，env 供用例模拟其他标签页的写入。
function createScenario(requestTokenPair: RequestTokenPair) {
  const env = installFakeBrowserEnv()
  env.seedSession(STALE_SESSION)

  const store = createAuthStore({
    isBrowser: () => true,
    logoutApi: async (_refreshToken?: string) => {}
  })
  const dispatchRefresh = (refreshToken: string) => requestTokenPair(refreshToken, env)

  return {
    env,
    store,
    refreshAccessToken: createTokenRefresher({ store, requestTokenPair: dispatchRefresh })
  }
}

describe('createTokenRefresher 跨标签页对齐', () => {
  it('其他标签页已完成旋转时复用其令牌且不发起刷新请求', async () => {
    const requestTokenPair = vi.fn<RequestTokenPair>(async () => NEW_PAIR)
    const { env, refreshAccessToken } = createScenario(requestTokenPair)

    env.seedSession(ROTATED_SESSION)
    const token = await refreshAccessToken()

    expect(token).toBe('rotated-access')
    expect(requestTokenPair).not.toHaveBeenCalled()
  })

  it('无其他标签页写入时发起真实刷新并写回新令牌', async () => {
    const requestTokenPair = vi.fn<RequestTokenPair>(async () => NEW_PAIR)
    const { store, refreshAccessToken } = createScenario(requestTokenPair)

    const token = await refreshAccessToken()

    expect(requestTokenPair).toHaveBeenCalledTimes(1)
    expect(requestTokenPair.mock.calls[0]?.[0]).toBe('stale-refresh')
    expect(token).toBe('fresh-access')
    expect(store.getAccessToken()).toBe('fresh-access')
    expect(store.getRefreshToken()).toBe('fresh-refresh')
  })

  it('刷新失败但其他标签页已兑换同一枚刷新令牌时复用其令牌', async () => {
    // 模拟并发竞态：本页刷新被服务端拒绝，同时其他标签页兑换成功并落盘。
    const requestTokenPair = vi.fn<RequestTokenPair>(async (_refreshToken, env) => {
      env.seedSession(ROTATED_SESSION)
      return null
    })
    const { store, refreshAccessToken } = createScenario(requestTokenPair)

    const token = await refreshAccessToken()

    expect(token).toBe('rotated-access')
    expect(store.getAccessToken()).toBe('rotated-access')
  })

  it('刷新失败且无其他标签页成功时返回 null', async () => {
    const requestTokenPair = vi.fn<RequestTokenPair>(async () => null)
    const { refreshAccessToken } = createScenario(requestTokenPair)

    expect(await refreshAccessToken()).toBeNull()
  })

  it('缺少刷新令牌时直接返回 null 且不发请求', async () => {
    installFakeBrowserEnv()
    const requestTokenPair = vi.fn<TokenRefreshDeps['requestTokenPair']>(async () => NEW_PAIR)
    const store = createAuthStore({
      isBrowser: () => true,
      logoutApi: async (_refreshToken?: string) => {}
    })
    const refreshAccessToken = createTokenRefresher({ store, requestTokenPair })

    expect(await refreshAccessToken()).toBeNull()
    expect(requestTokenPair).not.toHaveBeenCalled()
  })

  it('并发刷新共享同一次请求', async () => {
    const requestTokenPair = vi.fn<RequestTokenPair>(async () => {
      await new Promise(resolve => setTimeout(resolve, 10))
      return NEW_PAIR
    })
    const { refreshAccessToken } = createScenario(requestTokenPair)

    const [first, second] = await Promise.all([refreshAccessToken(), refreshAccessToken()])

    expect(requestTokenPair).toHaveBeenCalledTimes(1)
    expect(first).toBe('fresh-access')
    expect(second).toBe('fresh-access')
  })
})
