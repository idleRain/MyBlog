// 带自动重试和错误处理的请求工具

import { refreshAccessToken } from '$lib/service'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { goto } from '$app/navigation'
import { toast } from 'svelte-sonner'
import request from '$lib/service'

/**
 * 带自动重试的请求函数
 * 当遇到 401 错误时，会自动尝试刷新令牌并重试请求
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

      // 如果是 401 错误且还有重试机会
      if (error.response?.status === 401 && attempt < maxRetries) {
        console.log(`请求失败 (401)，尝试第 ${attempt + 1} 次重试...`)

        try {
          // 尝试刷新令牌
          const newToken = await refreshAccessToken()

          if (newToken) {
            console.log('令牌刷新成功，重试请求')
            continue // 继续下一次循环，重试请求
          } else {
            console.error('令牌刷新失败，停止重试')
            break // 刷新失败，停止重试
          }
        } catch (refreshError) {
          console.error('令牌刷新过程中出错:', refreshError)
          break // 刷新过程出错，停止重试
        }
      } else {
        // 不是 401 错误或已达到最大重试次数
        break
      }
    }
  }

  // 如果最终失败，抛出最后的错误
  throw lastError
}

/**
 * 安全的 API 调用包装器
 * 提供统一的错误处理和用户提示
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

    // 处理认证错误
    if (error.response?.status === 401) {
      if (redirectOnAuthError && browser) {
        authStore.clearLocalState() // 401 错误时只清除本地状态，不调用后端
        toast.error('登录已过期，请重新登录')
        await goto('/login')
      }
      return { data: null, error, success: false }
    }

    // 显示错误提示
    if (showErrorToast && browser) {
      const message =
        errorMessage || error.response?.data?.message || error.message || '操作失败，请稍后重试'
      toast.error(message)
    }

    return { data: null, error, success: false }
  }
}

/**
 * 检查响应是否成功
 */
export function isApiSuccess(response: any): boolean {
  return response?.code === 200
}

/**
 * 提取 API 响应数据
 */
export function extractApiData<T>(response: any, defaultValue?: T): T {
  if (isApiSuccess(response)) {
    return response.data
  }

  if (defaultValue !== undefined) {
    return defaultValue
  }

  throw new Error(response?.message || '请求失败')
}

/**
 * 创建带重试的 API 方法
 */
export function createRetryableApi<P extends any[], R>(
  originalMethod: (...params: P) => Promise<R>
) {
  return async (...params: P): Promise<R> => {
    return requestWithRetry(() => originalMethod(...params))
  }
}

// 示例使用方法：
// const safeUserList = createRetryableApi(UserAPI.getUserList)
// const result = await safeApiCall(() => safeUserList(1, 10))
