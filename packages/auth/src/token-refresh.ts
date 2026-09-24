import type { RefreshTokenData } from '@myblog/api/modules/user/types'
import type { AuthStore } from './auth-store'

/**
 * 令牌刷新编排依赖。
 */
export interface TokenRefreshDeps {
  // 认证 store，提供令牌读取、写回与单飞调度。
  store: AuthStore
  // 发起一次真实的刷新请求，失败返回 null。
  requestTokenPair: (refreshToken: string) => Promise<RefreshTokenData | null>
}

/**
 * 创建令牌刷新器，返回单飞模式下的刷新函数。
 *
 * 刷新即旋转语义下同一枚刷新令牌只能被兑换一次，编排必须覆盖三种情形：
 * 其他标签页已完成旋转时复用其令牌、本页确实需要刷新时发起真实请求、
 * 并发下本页刷新失败但其他标签页已成功时重新对齐并复用。
 */
export function createTokenRefresher(deps: TokenRefreshDeps): () => Promise<string | null> {
  const { store, requestTokenPair } = deps

  // 对齐其他标签页写入的令牌，仅当取到与原值不同的令牌时才算可复用。
  // 本编排由 401 路径触发，原令牌已被服务端拒绝，原样返回只会重放同一枚失效令牌。
  function syncToDifferentToken(previousToken: string | null): string | null {
    store.syncFromStorage()

    const syncedToken = store.getAccessToken()
    if (!syncedToken || syncedToken === previousToken) return null

    return store.isTokenValid() ? syncedToken : null
  }

  async function refresh(): Promise<string | null> {
    const rejectedToken = store.getAccessToken()

    // 其他标签页刚完成旋转时直接复用，避免携带已被撤销的旧刷新令牌发起注定失败的请求。
    const rotatedToken = syncToDifferentToken(rejectedToken)
    if (rotatedToken) return rotatedToken

    const refreshToken = store.getRefreshToken()
    if (!refreshToken) {
      console.warn('没有刷新令牌，无法自动刷新')
      return null
    }

    const pair = await requestTokenPair(refreshToken)
    if (pair) {
      store.updateTokens(pair.accessToken, pair.refreshToken, pair.expiresIn)
      return pair.accessToken
    }

    // 刷新失败也可能是并发下其他标签页先兑换了同一枚刷新令牌，
    // 对方的旋转结果已落盘，重新对齐即可复用，无需判定为认证失效。
    return syncToDifferentToken(rejectedToken)
  }

  // 单飞调度确保同一页面内的并发刷新共享同一次请求。
  return () => store.refreshSingleFlight(refresh)
}
