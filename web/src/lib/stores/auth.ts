import type { User } from '$lib/api/modules/user/types'
import { browser } from '$app/environment'
import { writable } from 'svelte/store'

// 认证状态接口
interface AuthState {
  isAuthenticated: boolean
  user: User | null
  token: string | null
}

// 初始状态
const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  token: null
}

// 从localStorage加载初始状态
function loadInitialState(): AuthState {
  if (!browser) return initialState

  try {
    const token = localStorage.getItem('auth_token')
    const user = localStorage.getItem('auth_user')

    if (token && user) {
      return {
        isAuthenticated: true,
        user: JSON.parse(user),
        token
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
    login(user: User, token: string) {
      const authState: AuthState = {
        isAuthenticated: true,
        user,
        token
      }

      if (browser) {
        localStorage.setItem('auth_token', token)
        localStorage.setItem('auth_user', JSON.stringify(user))
      }

      set(authState)
    },

    // 登出
    logout() {
      if (browser) {
        localStorage.removeItem('auth_token')
        localStorage.removeItem('auth_user')
      }

      set(initialState)
    },

    // 更新用户信息
    updateUser(user: User) {
      update(state => {
        const newState = { ...state, user }

        if (browser) {
          localStorage.setItem('auth_user', JSON.stringify(user))
        }

        return newState
      })
    },

    // 检查token是否有效
    isTokenValid(): boolean {
      const state = loadInitialState()
      return state.isAuthenticated && !!state.token
    },

    // 获取token
    getToken(): string | null {
      if (!browser) return null
      return localStorage.getItem('auth_token')
    }
  }
}

export const authStore = createAuthStore()
