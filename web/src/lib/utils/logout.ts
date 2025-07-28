// 退出登录工具函数

import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'
import { toast } from 'svelte-sonner'
import { UserAPI } from '$lib/api'

/**
 * 执行完整的退出登录流程
 * 1. 调用后端 logout 接口使 token 失效
 * 2. 清除本地认证状态
 * 3. 跳转到登录页面
 */
export async function performLogout(
  options: {
    showToast?: boolean
    redirectTo?: string | null
    skipApiCall?: boolean
  } = {}
): Promise<boolean> {
  const { showToast = true, redirectTo = '/login', skipApiCall = false } = options

  let apiCallSuccess = false

  try {
    if (!skipApiCall) {
      // 调用后端登出接口
      await authStore.logout(false) // 不跳过 API 调用
      apiCallSuccess = true
      console.log('成功调用后端登出接口')
    } else {
      // 仅清除本地状态
      authStore.clearLocalState()
      console.log('已清除本地认证状态')
    }

    if (showToast && browser) {
      if (apiCallSuccess) {
        toast.success('已成功退出登录')
      } else {
        toast.info('已清除本地登录状态')
      }
    }

    return true
  } catch (error) {
    console.error('退出登录失败:', error)

    // 即使后端调用失败，也要清除本地状态
    authStore.clearLocalState()

    if (showToast && browser) {
      toast.error('退出登录时发生错误，但已清除本地状态')
    }

    return false
  } finally {
    // 无论成功失败，都跳转到指定页面
    if (browser && redirectTo) {
      await goto(redirectTo)
    }
  }
}

/**
 * 强制退出登录（仅清除本地状态）
 * 用于 401 错误等无法调用后端接口的场景
 */
export async function forceLogout(
  options: {
    showToast?: boolean
    redirectTo?: string
    reason?: string
  } = {}
): Promise<void> {
  const { showToast = true, redirectTo = '/login', reason = '登录已过期，请重新登录' } = options

  console.log('强制退出登录:', reason)

  // 直接清除本地状态，不调用后端
  authStore.clearLocalState()

  if (showToast && browser) {
    toast.error(reason)
  }

  if (browser && redirectTo) {
    await goto(redirectTo)
  }
}

/**
 * 检查是否需要退出登录
 * 用于页面加载时检查认证状态
 */
export function shouldLogout(): boolean {
  const state = authStore.getCurrentState()

  // 如果没有认证信息，不需要退出
  if (!state.isAuthenticated || !state.accessToken) {
    return false
  }

  // 如果 token 有效，不需要退出
  if (authStore.isTokenValid()) {
    return false
  }

  // 如果有 refresh token，可以尝试刷新，不需要立即退出
  if (state.refreshToken) {
    return false
  }

  // 其他情况需要退出
  return true
}

/**
 * 安全退出登录的包装器
 * 确保在任何情况下都能正确清除状态
 */
export async function safeLogout(reason?: string): Promise<void> {
  try {
    await performLogout({
      showToast: true,
      redirectTo: '/login'
    })
  } catch (error) {
    console.error('安全退出登录失败:', error)

    // 备用方案：强制退出
    await forceLogout({
      reason: reason || '退出登录时发生错误，已清除本地状态'
    })
  }
}

/**
 * 批量退出（如果有多个标签页/窗口）
 * 通过 localStorage 事件通知其他窗口
 */
export function broadcastLogout(): void {
  if (browser) {
    // 设置一个临时标记，其他窗口会监听这个变化
    localStorage.setItem('logout_broadcast', Date.now().toString())

    // 立即清除这个标记
    setTimeout(() => {
      localStorage.removeItem('logout_broadcast')
    }, 100)
  }
}

/**
 * 监听其他窗口的退出登录事件
 */
export function listenForLogoutBroadcast(callback: () => void): () => void {
  if (!browser) return () => {}

  const handleStorageChange = (event: StorageEvent) => {
    if (event.key === 'logout_broadcast' && event.newValue) {
      console.log('收到其他窗口的退出登录广播')
      callback()
    }
  }

  window.addEventListener('storage', handleStorageChange)

  return () => {
    window.removeEventListener('storage', handleStorageChange)
  }
}
