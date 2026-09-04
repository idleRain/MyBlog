// 认证守卫工具

import { goto } from '$lib/utils/navigation'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'

/**
 * 检查用户是否已认证
 */
export function isAuthenticated(): boolean {
  const state = authStore.getCurrentState()
  return state.isAuthenticated && !!state.user && authStore.isTokenValid()
}

/**
 * 要求用户登录
 * 如果用户未登录，跳转到登录页面
 */
export async function requireAuth(redirectTo: string = '/login'): Promise<boolean> {
  if (!browser) return true // SSR 时不检查

  if (!isAuthenticated()) {
    console.log('用户未认证，跳转到登录页面')
    await goto(redirectTo)
    return false
  }

  return true
}

/**
 * 要求用户未登录
 * 如果用户已登录，跳转到首页或指定页面
 */
export async function requireGuest(redirectTo: string = '/'): Promise<boolean> {
  if (!browser) return true // SSR 时不检查

  if (isAuthenticated()) {
    console.log('用户已认证，跳转到首页')
    await goto(redirectTo)
    return false
  }

  return true
}

/**
 * 页面加载时的认证检查
 * 确保在页面完全加载前完成认证检查
 */
export async function checkAuthOnLoad(): Promise<{
  isAuthenticated: boolean
  user: any
  needsRedirect: boolean
  redirectTo?: string
}> {
  if (!browser) {
    return {
      isAuthenticated: false,
      user: null,
      needsRedirect: false
    }
  }

  const state = authStore.getCurrentState()
  const authenticated = state.isAuthenticated && !!state.user

  // 检查 token 是否有效
  if (authenticated && !authStore.isTokenValid()) {
    // Token 无效，尝试刷新
    try {
      const { refreshAccessToken } = await import('$lib/service')
      const newToken = await refreshAccessToken()

      if (!newToken) {
        // 刷新失败，清除本地状态，需要重新登录
        authStore.clearLocalState()
        return {
          isAuthenticated: false,
          user: null,
          needsRedirect: true,
          redirectTo: '/login'
        }
      }
    } catch (error) {
      console.error('Token 刷新失败:', error)
      authStore.clearLocalState()
      return {
        isAuthenticated: false,
        user: null,
        needsRedirect: true,
        redirectTo: '/login'
      }
    }
  }

  return {
    isAuthenticated: authenticated,
    user: state.user,
    needsRedirect: false
  }
}
