/**
 * JWT 工具函数
 */

import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'
import { UserAPI } from '$lib/api'

/**
 * 解析JWT payload（不验证签名，仅用于前端展示）
 * @param token JWT令牌
 * @returns 解析后的payload或null
 */
export function parseJWTPayload(token: string): Record<string, any> | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) {
      return null
    }

    const payload = parts[1]!
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'))
    return JSON.parse(decoded)
  } catch (error) {
    console.error('JWT解析失败:', error)
    return null
  }
}

/**
 * 检查JWT是否即将过期（5分钟内）
 * @param token JWT令牌
 * @returns 是否即将过期
 */
export function isTokenExpiringSoon(token: string): boolean {
  const payload = parseJWTPayload(token)
  if (!payload || !payload.exp) {
    return true
  }

  const expirationTime = payload.exp * 1000 // 转换为毫秒
  const fiveMinutesFromNow = Date.now() + 5 * 60 * 1000

  return expirationTime <= fiveMinutesFromNow
}

/**
 * 手动刷新令牌
 * @returns 是否刷新成功
 */
export async function manualRefreshToken(): Promise<boolean> {
  if (!browser) return false

  const refreshToken = authStore.getRefreshToken()
  if (!refreshToken) {
    authStore.logout()
    await goto('/login')
    return false
  }

  try {
    const response = await UserAPI.refreshToken({
      refreshToken: refreshToken
    })

    if (response.code === 200 && response.data) {
      authStore.updateTokens(
        response.data.accessToken,
        response.data.refreshToken,
        response.data.expiresIn
      )
      return true
    }
  } catch (error) {
    console.error('手动刷新令牌失败:', error)
  }

  // 刷新失败，清除认证状态
  authStore.logout()
  await goto('/login')
  return false
}

/**
 * 执行登出操作
 * @param skipAPICall 是否跳过API调用（用于服务器错误时）
 */
export async function performLogout(skipAPICall = false): Promise<void> {
  if (!browser) return

  if (!skipAPICall) {
    try {
      await UserAPI.logout()
    } catch (error) {
      console.error('登出API调用失败:', error)
      // 即使API调用失败，也要清除本地状态
    }
  }

  authStore.logout()
  await goto('/login')
}

/**
 * 检查用户是否已认证且令牌有效
 * @returns 是否有效
 */
export function isAuthenticated(): boolean {
  return authStore.isTokenValid()
}

/**
 * 获取当前用户的认证状态
 * @returns 认证状态信息
 */
export function getAuthStatus() {
  const accessToken = authStore.getAccessToken()
  const refreshToken = authStore.getRefreshToken()

  if (!accessToken || !refreshToken) {
    return {
      isAuthenticated: false,
      tokenValid: false,
      needsRefresh: false
    }
  }

  const payload = parseJWTPayload(accessToken)
  const now = Date.now()
  const expirationTime = payload?.exp ? payload.exp * 1000 : 0

  return {
    isAuthenticated: true,
    tokenValid: now < expirationTime,
    needsRefresh: now >= expirationTime - 5 * 60 * 1000, // 5分钟内过期
    expiresAt: new Date(expirationTime),
    user: payload
      ? {
          id: payload.user_id,
          username: payload.username,
          tokenType: payload.token_type
        }
      : null
  }
}
