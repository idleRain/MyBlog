// API 相关类型增强

import type { UserStatus } from '$lib/api/modules/user/types'
import type { UserRole } from './auth'

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
 * 用户相关 API 类型增强
 */
export namespace UserAPI {
  export interface ListRequest extends BaseListRequest {
    role?: UserRole
    status?: UserStatus
    email?: string
  }

  export interface ListResponse extends PaginationResponse<UserListItem> {}

  export interface UserListItem {
    id: number
    username: string
    email: string
    nickname: string
    avatar?: string
    role: UserRole
    status: UserStatus
    lastLoginAt?: string
    createdAt: string
    updatedAt: string
  }

  export interface DetailRequest {
    id: number
    includeStats?: boolean
    includePermissions?: boolean
  }

  export interface DetailResponse {
    user: UserDetail
    stats?: UserStats
    permissions?: string[]
  }

  export interface UserDetail {
    id: number
    username: string
    email: string
    nickname: string
    avatar?: string
    birthday?: string
    role: UserRole
    status: UserStatus
    lastLoginAt?: string
    loginCount: number
    createdAt: string
    updatedAt: string
  }

  export interface UserStats {
    loginCount: number
    lastLoginAt?: string
    articlesCount: number
    commentsCount: number
  }

  export interface CreateRequest {
    username: string
    email: string
    password: string
    nickname?: string
    birthday?: string
    role?: UserRole
    avatar?: string
  }

  export interface UpdateRequest {
    id: number
    username?: string
    email?: string
    password?: string
    nickname?: string
    birthday?: string
    role?: UserRole
    status?: UserStatus
    avatar?: string
  }

  export interface BatchUpdateRequest {
    ids: number[]
    updates: {
      status?: UserStatus
      role?: UserRole
    }
  }

  export interface DeleteRequest {
    id: number
    force?: boolean // 是否强制删除
  }

  export interface BatchDeleteRequest {
    ids: number[]
    force?: boolean
  }
}

/**
 * 认证相关 API 类型
 */
export namespace AuthAPI {
  export interface LoginRequest {
    username: string
    password: string
    remember?: boolean
    captcha?: string
  }

  export interface LoginResponse {
    user: {
      id: number
      username: string
      email: string
      nickname: string
      avatar?: string
      role: UserRole
      status: UserStatus
    }
    tokens: {
      accessToken: string
      refreshToken: string
      expiresIn: number
      expiresAt: number
    }
    session: {
      sessionId: string
      userAgent?: string
      ipAddress?: string
      location?: string
    }
  }

  export interface RegisterRequest {
    username: string
    email: string
    password: string
    nickname?: string
    birthday?: string
    inviteCode?: string
  }

  export interface RefreshTokenRequest {
    refreshToken: string
  }

  export interface RefreshTokenResponse {
    accessToken: string
    refreshToken: string
    expiresIn: number
    expiresAt: number
  }

  export interface LogoutRequest {
    refreshToken?: string
    allDevices?: boolean
  }

  export interface ResetPasswordRequest {
    token: string
    newPassword: string
  }

  export interface ForgotPasswordRequest {
    email: string
  }
}

/**
 * 文章相关 API 类型（预留）
 */
export namespace ArticleAPI {
  export interface ListRequest extends BaseListRequest {
    category?: string
    tags?: string[]
    authorId?: number
    published?: boolean
  }

  export interface ArticleListItem {
    id: number
    title: string
    slug: string
    excerpt?: string
    author: {
      id: number
      username: string
      nickname: string
      avatar?: string
    }
    category?: string
    tags: string[]
    published: boolean
    viewCount: number
    commentCount: number
    publishedAt?: string
    createdAt: string
    updatedAt: string
  }
}

/**
 * 系统相关 API 类型
 */
export namespace SystemAPI {
  export interface StatsResponse {
    users: {
      total: number
      active: number
      newToday: number
      disabled: number
    }
    articles: {
      total: number
      published: number
      draft: number
      viewsToday: number
    }
    system: {
      status: 'normal' | 'warning' | 'error'
      uptime: number
      version: string
      lastBackup?: string
    }
  }

  export interface ConfigRequest {
    key: string
    value: any
    description?: string
  }

  export interface ConfigResponse {
    key: string
    value: any
    description?: string
    updatedAt: string
  }
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
