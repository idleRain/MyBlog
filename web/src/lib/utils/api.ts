// API 工具函数

import type {
  BaseApiResponse,
  ApiError,
  AsyncState,
  AsyncStateActions,
  RequestStatus
} from '$lib/types'
import { writable, derived, type Writable } from 'svelte/store'

/**
 * 创建异步状态管理
 */
export function createAsyncState<T = any>(
  initialData: T | null = null
): {
  state: Writable<AsyncState<T>>
  actions: AsyncStateActions<T>
  derived: {
    loading: any
    error: any
    data: any
    hasData: any
    hasError: any
  }
} {
  const state = writable<AsyncState<T>>({
    status: 'idle',
    data: initialData,
    error: null,
    loading: false
  })

  const actions: AsyncStateActions<T> = {
    execute: async (...args: any[]) => {
      state.update(s => ({ ...s, status: 'loading', loading: true, error: null }))

      try {
        // 这里的实际执行逻辑需要在使用时提供
        throw new Error('Execute function not implemented')
      } catch (error) {
        const apiError = normalizeError(error)
        state.update(s => ({
          ...s,
          status: 'error',
          loading: false,
          error: apiError
        }))
        throw apiError
      }
    },

    reset: () => {
      state.set({
        status: 'idle',
        data: initialData,
        error: null,
        loading: false
      })
    },

    setLoading: (loading: boolean) => {
      state.update(s => ({
        ...s,
        loading,
        status: loading ? 'loading' : s.status
      }))
    },

    setError: (error: ApiError | null) => {
      state.update(s => ({
        ...s,
        error,
        status: error ? 'error' : s.status,
        loading: false
      }))
    },

    setData: (data: T | null) => {
      state.update(s => ({
        ...s,
        data,
        status: 'success',
        loading: false,
        error: null
      }))
    }
  }

  // 派生状态
  const derivedStates = {
    loading: derived(state, $state => $state.loading),
    error: derived(state, $state => $state.error),
    data: derived(state, $state => $state.data),
    hasData: derived(state, $state => $state.data !== null),
    hasError: derived(state, $state => $state.error !== null)
  }

  return {
    state,
    actions,
    derived: derivedStates
  }
}

/**
 * 创建带有执行函数的异步状态
 */
export function createAsyncOperation<T, Args extends any[] = any[]>(
  operation: (...args: Args) => Promise<T>,
  initialData: T | null = null
) {
  const { state, actions, derived } = createAsyncState<T>(initialData)

  const execute = async (...args: Args): Promise<T> => {
    state.update(s => ({ ...s, status: 'loading' as RequestStatus, loading: true, error: null }))

    try {
      const result = await operation(...args)
      state.update(s => ({
        ...s,
        status: 'success' as RequestStatus,
        loading: false,
        data: result,
        error: null
      }))
      return result
    } catch (error) {
      const apiError = normalizeError(error)
      state.update(s => ({
        ...s,
        status: 'error' as RequestStatus,
        loading: false,
        error: apiError
      }))
      throw apiError
    }
  }

  return {
    state,
    execute,
    ...actions,
    derived
  }
}

/**
 * 标准化错误对象
 */
export function normalizeError(error: any): ApiError {
  if (isApiError(error)) {
    return error
  }

  if (error?.response?.data) {
    const { code, message, details, field } = error.response.data
    return {
      code: code || error.response.status || 500,
      message: message || '请求失败',
      details,
      field,
      timestamp: Date.now()
    }
  }

  if (error instanceof Error) {
    return {
      code: 500,
      message: error.message || '未知错误',
      timestamp: Date.now()
    }
  }

  return {
    code: 500,
    message: '未知错误',
    timestamp: Date.now()
  }
}

/**
 * 检查是否为 API 错误
 */
export function isApiError(obj: any): obj is ApiError {
  return (
    obj &&
    typeof obj === 'object' &&
    typeof obj.code === 'number' &&
    typeof obj.message === 'string' &&
    typeof obj.timestamp === 'number'
  )
}

/**
 * 检查 API 响应是否成功
 */
export function isApiSuccess<T>(
  response: BaseApiResponse<T>
): response is BaseApiResponse<T> & { code: 200 } {
  return response.code === 200
}

/**
 * 提取 API 响应数据
 */
export function extractApiData<T>(response: BaseApiResponse<T>): T {
  if (!isApiSuccess(response)) {
    throw normalizeError({
      code: response.code,
      message: response.message,
      timestamp: response.timestamp || Date.now()
    })
  }
  return response.data
}

/**
 * 安全地提取 API 响应数据
 */
export function safeExtractApiData<T>(response: BaseApiResponse<T>, defaultValue: T): T {
  try {
    return extractApiData(response)
  } catch {
    return defaultValue
  }
}

/**
 * 创建 API 响应
 */
export function createApiResponse<T>(
  data: T,
  code: number = 200,
  message: string = 'Success'
): BaseApiResponse<T> {
  return {
    code,
    message,
    data,
    timestamp: Date.now()
  }
}

/**
 * 创建 API 错误响应
 */
export function createApiErrorResponse(
  code: number,
  message: string,
  details?: Record<string, any>
): BaseApiResponse<null> {
  return {
    code,
    message,
    data: null,
    timestamp: Date.now()
  }
}

/**
 * 重试机制
 */
export async function withRetry<T>(
  operation: () => Promise<T>,
  maxRetries: number = 3,
  delay: number = 1000
): Promise<T> {
  let lastError: any

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await operation()
    } catch (error) {
      lastError = error

      if (attempt === maxRetries) {
        throw lastError
      }

      // 指数退避延迟
      const waitTime = delay * Math.pow(2, attempt - 1)
      await new Promise(resolve => setTimeout(resolve, waitTime))
    }
  }

  throw lastError
}

/**
 * 超时处理
 */
export function withTimeout<T>(promise: Promise<T>, timeoutMs: number): Promise<T> {
  return Promise.race([
    promise,
    new Promise<never>((_, reject) => {
      setTimeout(() => {
        reject(
          normalizeError({
            code: 408,
            message: '请求超时',
            timestamp: Date.now()
          })
        )
      }, timeoutMs)
    })
  ])
}

/**
 * 防抖处理 API 调用
 */
export function debounceApiCall<Args extends any[], Return>(
  fn: (...args: Args) => Promise<Return>,
  delay: number = 300
): (...args: Args) => Promise<Return> {
  let timeoutId: NodeJS.Timeout | null = null
  let latestResolve: ((value: Return) => void) | null = null
  let latestReject: ((reason: any) => void) | null = null

  return (...args: Args): Promise<Return> => {
    return new Promise<Return>((resolve, reject) => {
      // 清除之前的定时器
      if (timeoutId) {
        clearTimeout(timeoutId)
      }

      // 如果有之前的 Promise，先拒绝它
      if (latestReject) {
        latestReject(
          normalizeError({
            code: 499,
            message: '请求被取消',
            timestamp: Date.now()
          })
        )
      }

      latestResolve = resolve
      latestReject = reject

      timeoutId = setTimeout(async () => {
        try {
          const result = await fn(...args)
          if (latestResolve === resolve) {
            resolve(result)
          }
        } catch (error) {
          if (latestReject === reject) {
            reject(error)
          }
        } finally {
          timeoutId = null
          latestResolve = null
          latestReject = null
        }
      }, delay)
    })
  }
}

/**
 * 批量处理 API 请求
 */
export async function batchApiRequests<T, R>(
  items: T[],
  processor: (item: T) => Promise<R>,
  options: {
    batchSize?: number
    delay?: number
    onProgress?: (processed: number, total: number) => void
  } = {}
): Promise<R[]> {
  const { batchSize = 5, delay = 100, onProgress } = options
  const results: R[] = []

  for (let i = 0; i < items.length; i += batchSize) {
    const batch = items.slice(i, i + batchSize)
    const batchPromises = batch.map(processor)

    try {
      const batchResults = await Promise.all(batchPromises)
      results.push(...batchResults)

      if (onProgress) {
        onProgress(results.length, items.length)
      }

      // 批次之间的延迟
      if (i + batchSize < items.length && delay > 0) {
        await new Promise(resolve => setTimeout(resolve, delay))
      }
    } catch (error) {
      throw normalizeError(error)
    }
  }

  return results
}

/**
 * 格式化错误消息
 */
export function formatErrorMessage(error: ApiError): string {
  switch (error.code) {
    case 400:
      return '请求参数错误'
    case 401:
      return '未授权，请重新登录'
    case 403:
      return '权限不足'
    case 404:
      return '资源不存在'
    case 408:
      return '请求超时'
    case 409:
      return '资源冲突'
    case 422:
      return '数据验证失败'
    case 429:
      return '请求过于频繁，请稍后再试'
    case 500:
      return '服务器内部错误'
    case 502:
      return '网关错误'
    case 503:
      return '服务暂不可用'
    default:
      return error.message || '未知错误'
  }
}
