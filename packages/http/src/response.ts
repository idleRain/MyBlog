import type { BaseApiResponse } from '@myblog/shared'

export type { BaseApiResponse }

/**
 * API 错误结构，供错误规范化使用。
 */
export interface ApiError {
  code: number
  message: string
  details?: Record<string, any>
  field?: string
  timestamp: number
}

/**
 * 检查 API 响应是否成功。
 */
export function isApiSuccess<T>(
  response: BaseApiResponse<T>
): response is BaseApiResponse<T> & { code: 200 } {
  return response.code === 200
}

/**
 * 标准化任意错误为 ApiError 结构。
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
 * 判断对象是否为 ApiError 结构。
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
 * 提取 API 响应数据，失败时抛出规范化错误。
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
 * 安全提取 API 响应数据，失败时返回默认值。
 */
export function safeExtractApiData<T>(response: BaseApiResponse<T>, defaultValue: T): T {
  try {
    return extractApiData(response)
  } catch {
    return defaultValue
  }
}
