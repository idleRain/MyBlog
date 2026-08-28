// API 基础类型：与具体业务模块无关的通用请求与响应结构。

/**
 * 基础 API 响应结构
 */
export interface BaseApiResponse<T = any> {
  code: number
  message: string
  data: T
  timestamp?: number
  requestId?: string
}

/**
 * 分页参数
 */
export interface PaginationRequest {
  page: number
  pageSize: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

/**
 * 分页响应数据
 */
export interface PaginationResponse<T> {
  page: number
  pageSize: number
  pages: number
  total: number
  items: T[]
  hasNext: boolean
  hasPrev: boolean
}

/**
 * 列表查询参数基类
 */
export interface BaseListRequest extends PaginationRequest {
  keyword?: string
  status?: number | string
  createdAt?: {
    start?: string
    end?: string
  }
}

/**
 * API 错误响应
 */
export interface ApiError {
  code: number
  message: string
  details?: Record<string, any>
  field?: string
  timestamp: number
}

/**
 * 请求状态枚举
 */
export type RequestStatus = 'idle' | 'loading' | 'success' | 'error'

/**
 * 异步操作状态
 */
export interface AsyncState<T = any> {
  status: RequestStatus
  data: T | null
  error: ApiError | null
  loading: boolean
}

/**
 * 创建异步状态的工具函数返回类型
 */
export interface AsyncStateActions<T> {
  execute: (...args: any[]) => Promise<T>
  reset: () => void
  setLoading: (loading: boolean) => void
  setError: (error: ApiError | null) => void
  setData: (data: T | null) => void
}

/**
 * 通用响应类型别名
 */
export type ApiResponse<T = any> = BaseApiResponse<T>
export type ListResponse<T> = BaseApiResponse<PaginationResponse<T>>
export type DetailResponse<T> = BaseApiResponse<T>
export type CreateResponse<T> = BaseApiResponse<T>
export type UpdateResponse<T> = BaseApiResponse<T>
export type DeleteResponse = BaseApiResponse<null>
export type BatchResponse<T = any> = BaseApiResponse<{
  success: number
  failed: number
  results: T[]
}>

/**
 * HTTP 方法类型
 */
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

/**
 * API 请求配置
 */
export interface ApiRequestConfig {
  method: HttpMethod
  url: string
  data?: any
  params?: Record<string, any>
  headers?: Record<string, string>
  timeout?: number
  retry?: number
  cache?: boolean
}
