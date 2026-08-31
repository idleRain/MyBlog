// 带自动重试和错误处理的请求工具

import { refreshAccessToken } from '$lib/service'
import { goto } from '$lib/utils/navigation'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { toast } from 'svelte-sonner'

/**
 * 带自动重试的请求函数，遇到 401 时尝试刷新令牌并重试。
 */
export async function requestWithRetry<T = any>(
  requestFn: () => Promise<T>,
  maxRetries: number = 1
): Promise<T> {
  let lastError: any

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      return await requestFn()
    } catch (error: any) {
      lastError = error

      // 401 错误且还有重试机会时尝试刷新令牌。
      if (error.response?.status === 401 && attempt < maxRetries) {
        console.log(`请求失败 (401)，尝试第 ${attempt + 1} 次重试...`)

        try {
          const newToken = await refreshAccessToken()
          if (newToken) {
            console.log('令牌刷新成功，重试请求')
            continue
          }
          console.error('令牌刷新失败，停止重试')
          break
        } catch (refreshError) {
          console.error('令牌刷新过程中出错:', refreshError)
          break
        }
      } else {
        break
      }
    }
  }

  throw lastError
}

/**
 * 安全的 API 调用包装器，提供统一错误处理与用户提示。
 */
export async function safeApiCall<T = any>(
  apiCall: () => Promise<T>,
  options: {
    showErrorToast?: boolean
    redirectOnAuthError?: boolean
    errorMessage?: string
  } = {}
): Promise<{ data: T | null; error: any | null; success: boolean }> {
  const { showErrorToast = true, redirectOnAuthError = true, errorMessage } = options

  try {
    const data = await requestWithRetry(apiCall)
    return { data, error: null, success: true }
  } catch (error: any) {
    console.error('API 调用失败:', error)

    // 处理认证错误。
    if (error.response?.status === 401) {
      if (redirectOnAuthError && browser) {
        authStore.clearLocalState()
        toast.error('登录已过期，请重新登录')
        await goto('/login')
      }
      return { data: null, error, success: false }
    }

    // 显示错误提示。
    if (showErrorToast && browser) {
      const message =
        errorMessage || error.response?.data?.message || error.message || '操作失败，请稍后重试'
      toast.error(message)
    }

    return { data: null, error, success: false }
  }
}

/**
 * 创建带重试的 API 方法。
 */
export function createRetryableApi<P extends any[], R>(
  originalMethod: (...params: P) => Promise<R>
) {
  return async (...params: P): Promise<R> => {
    return requestWithRetry(() => originalMethod(...params))
  }
}
