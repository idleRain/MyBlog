import type { User } from '$lib/api/modules/user/types'
import { local } from '$lib/utils/storage'
import { browser } from '$app/environment'
import { writable } from 'svelte/store'

// 认证状态接口
interface AuthState {
  isAuthenticated: boolean
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  expiresAt: number | null // 过期时间戳
}

// 初始状态
const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  accessToken: null,
  refreshToken: null,
  expiresAt: null
}

// 从localStorage加载初始状态
function loadInitialState(): AuthState {
  if (!browser) return initialState

  try {
    const accessToken = local.get<string>('auth_access_token')
    const refreshToken = local.get<string>('auth_refresh_token')
    const user = local.get<User>('auth_user')
    const expiresAt = local.get<number>('auth_expires_at')

    if (accessToken && refreshToken && user) {
      return {
        isAuthenticated: true,
        user,
        accessToken,
        refreshToken,
        expiresAt
      }
    }
  } catch (error) {
    console.error('Failed to load auth state:', error)
  }

  return initialState
}

// 创建认证store
function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(loadInitialState())

  return {
    subscribe,

    // 登录
    login(user: User, accessToken: string, refreshToken: string, expiresIn: number) {
      const expiresAt = Date.now() + expiresIn * 1000 // 转换为毫秒时间戳

      const authState: AuthState = {
        isAuthenticated: true,
        user,
        accessToken,
        refreshToken,
        expiresAt
      }

      if (browser) {
        local.set('auth_access_token', accessToken)
        local.set('auth_refresh_token', refreshToken)
        local.set('auth_user', user)
        local.set('auth_expires_at', expiresAt)
      }

      set(authState)
    },

    // 登出
    logout() {
      if (browser) {
        local.rm('auth_access_token')
        local.rm('auth_refresh_token')
        local.rm('auth_user')
        local.rm('auth_expires_at')
      }

      set(initialState)
    },

    // 更新用户信息
    updateUser(user: User) {
      update(state => {
        const newState = { ...state, user }

        if (browser) {
          local.set('auth_user', user)
        }

        return newState
      })
    },

    // 更新令牌
    updateTokens(accessToken: string, refreshToken: string, expiresIn: number) {
      const expiresAt = Date.now() + expiresIn * 1000

      update(state => {
        const newState = {
          ...state,
          accessToken,
          refreshToken,
          expiresAt
        }

        if (browser) {
          local.set('auth_access_token', accessToken)
          local.set('auth_refresh_token', refreshToken)
          local.set('auth_expires_at', expiresAt)
        }

        return newState
      })
    },

    // 检查token是否有效
    isTokenValid(): boolean {
      const state = loadInitialState()
      if (!state.isAuthenticated || !state.accessToken || !state.expiresAt) {
        return false
      }

      // 检查是否过期（提前5分钟刷新）
      return Date.now() < state.expiresAt - 5 * 60 * 1000
    },

    // 检查是否需要刷新token
    shouldRefreshToken(): boolean {
      const state = loadInitialState()
      if (!state.isAuthenticated || !state.accessToken || !state.expiresAt) {
        return false
      }

      // 如果在5分钟内过期，需要刷新
      return Date.now() >= state.expiresAt - 5 * 60 * 1000
    },

    // 获取访问令牌
    getAccessToken(): string | null {
      if (!browser) return null
      return local.get<string>('auth_access_token')
    },

    // 获取刷新令牌
    getRefreshToken(): string | null {
      if (!browser) return null
      return local.get<string>('auth_refresh_token')
    },

    // 向后兼容：获取token（返回access token）
    getToken(): string | null {
      return this.getAccessToken()
    }
  }
}

export const authStore = createAuthStore()
export type { AuthState }
