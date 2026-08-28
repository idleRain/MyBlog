/**
 * 认证守卫 - 用于保护需要认证的页面
 */

import { manualRefreshToken } from '$lib/utils/jwt'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'

/**
 * 检查用户是否已认证，如果未认证则重定向到登录页
 * @param redirectTo 登录后重定向的页面
 * @returns 是否已认证
 */
export async function requireAuth(redirectTo?: string): Promise<boolean> {
  if (!browser) return false

  // 检查基本认证状态
  if (!authStore.isTokenValid()) {
    // 尝试刷新令牌
    if (authStore.shouldRefreshToken()) {
      const refreshed = await manualRefreshToken()
      if (refreshed) {
        return true
      }
    }

    // 认证失败，重定向到登录页
    const loginUrl = redirectTo ? `/login?redirect=${encodeURIComponent(redirectTo)}` : '/login'
    await goto(loginUrl)
    return false
  }

  return true
}

/**
 * 检查用户是否已认证，如果已认证则重定向到首页（用于登录、注册页面）
 * @returns 是否需要重定向
 */
export async function requireGuest(): Promise<boolean> {
  if (!browser) return false

  if (authStore.isTokenValid()) {
    await goto('/')
    return true
  }

  return false
}

/**
 * 可选认证检查 - 不强制要求认证，但会尝试刷新即将过期的令牌
 * @returns 当前认证状态
 */
export async function optionalAuth(): Promise<{
  isAuthenticated: boolean
  user: any
}> {
  if (!browser) {
    return { isAuthenticated: false, user: null }
  }

  // 如果令牌即将过期，尝试刷新
  if (authStore.shouldRefreshToken()) {
    await manualRefreshToken()
  }

  const isAuthenticated = authStore.isTokenValid()
  let user = null

  if (isAuthenticated) {
    // 从store中获取用户信息
    authStore.subscribe(state => {
      user = state.user
    })()
  }

  return { isAuthenticated, user }
}

/**
 * 管理员权限检查
 * @returns 是否具有管理员权限
 */
export async function requireAdmin(): Promise<boolean> {
  const authenticated = await requireAuth()
  if (!authenticated) return false

  // TODO: 实现管理员权限检查逻辑
  // 这里可以从JWT payload或用户信息中检查角色
  // 目前先返回true，后续根据具体需求实现

  return true
}

/**
 * 创建路由守卫函数 - 用作SvelteKit的页面加载器
 * @param options 守卫选项
 * @returns 页面加载器函数
 */
export function createAuthGuard(
  options: {
    requireAuth?: boolean
    requireGuest?: boolean
    requireAdmin?: boolean
    redirectTo?: string
  } = {}
) {
  return async ({ url }: { url: URL }) => {
    const currentPath = url.pathname

    if (options.requireGuest) {
      await requireGuest()
      return {}
    }

    if (options.requireAuth) {
      const authenticated = await requireAuth(options.redirectTo || currentPath)
      if (!authenticated) {
        return {}
      }
    }

    if (options.requireAdmin) {
      const hasAdminAccess = await requireAdmin()
      if (!hasAdminAccess) {
        await goto('/unauthorized')
        return {}
      }
    }

    return {}
  }
}
