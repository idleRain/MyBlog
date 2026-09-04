import type { User } from '@myblog/api/modules/user/types'
import { writable } from 'svelte/store'
import { local } from '@myblog/shared'

// 认证状态接口
export interface AuthState {
  isAuthenticated: boolean
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  expiresAt: number | null // 过期时间戳
  permissions: string[] // 登录时由后端下发的权限列表
}

/**
 * 认证 store 依赖注入集合。
 * 应用层负责提供环境判定与登出接口，避免包内耦合 SvelteKit 环境与页面路由。
 */
export interface AuthStoreDeps {
  /**
   * 判断当前是否运行于浏览器环境，SSR 阶段禁止访问 localStorage。
   */
  isBrowser: () => boolean

  /**
   * 调用后端登出接口，供会话登出时同步撤销服务端令牌。
   */
  logoutApi: () => Promise<void>
}

// 初始状态
const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  accessToken: null,
  refreshToken: null,
  expiresAt: null,
  permissions: []
}

// localStorage 令牌存储键
const AUTH_TOKEN_KEY = 'auth_access_token'
const AUTH_REFRESH_KEY = 'auth_refresh_token'
const AUTH_USER_KEY = 'auth_user'
const AUTH_EXPIRES_KEY = 'auth_expires_at'
const AUTH_PERMISSIONS_KEY = 'auth_permissions'

// 令牌过期前提前刷新的窗口，单位毫秒
const REFRESH_LEAD_TIME_MS = 5 * 60 * 1000

/**
 * 创建认证 store 实例，供各应用各持有一份，避免全局单例造成状态串扰。
 */
export function createAuthStore(deps: AuthStoreDeps) {
  // 从 localStorage 加载初始状态
  function loadInitialState(): AuthState {
    if (!deps.isBrowser()) return initialState

    try {
      const accessToken = local.get<string>(AUTH_TOKEN_KEY)
      const refreshToken = local.get<string>(AUTH_REFRESH_KEY)
      const user = local.get<User>(AUTH_USER_KEY)
      const expiresAt = local.get<number>(AUTH_EXPIRES_KEY)
      const permissions = local.get<string[]>(AUTH_PERMISSIONS_KEY)

      if (accessToken && refreshToken && user) {
        return {
          isAuthenticated: true,
          user,
          accessToken,
          refreshToken,
          expiresAt,
          permissions: permissions ?? []
        }
      }
    } catch (error) {
      console.error('Failed to load auth state:', error)
    }

    return initialState
  }

  const { subscribe, set, update } = writable<AuthState>(loadInitialState())

  // 缓存当前状态，供同步方法读取。
  let currentState: AuthState = loadInitialState()
  subscribe(state => {
    currentState = state
  })

  // 清除本地存储中的全部认证令牌。
  function clearLocalStorage() {
    if (!deps.isBrowser()) return
    local.rm(AUTH_TOKEN_KEY)
    local.rm(AUTH_REFRESH_KEY)
    local.rm(AUTH_USER_KEY)
    local.rm(AUTH_EXPIRES_KEY)
    local.rm(AUTH_PERMISSIONS_KEY)
  }

  return {
    subscribe,

    // 获取当前状态
    getCurrentState(): AuthState {
      return currentState
    },

    // 登录并持久化令牌与权限
    login(
      user: User,
      accessToken: string,
      refreshToken: string,
      expiresIn: number,
      permissions: string[] = []
    ) {
      const expiresAt = Date.now() + expiresIn * 1000 // 转换为毫秒时间戳

      const authState: AuthState = {
        isAuthenticated: true,
        user,
        accessToken,
        refreshToken,
        expiresAt,
        permissions
      }

      if (deps.isBrowser()) {
        local.set(AUTH_TOKEN_KEY, accessToken)
        local.set(AUTH_REFRESH_KEY, refreshToken)
        local.set(AUTH_USER_KEY, user)
        local.set(AUTH_EXPIRES_KEY, expiresAt)
        local.set(AUTH_PERMISSIONS_KEY, permissions)
      }

      set(authState)
    },

    // 登出，先尝试撤销服务端令牌再清除本地状态。
    async logout(skipApiCall: boolean = false) {
      if (!skipApiCall && deps.isBrowser() && currentState.isAuthenticated) {
        try {
          await deps.logoutApi()
          console.log('成功调用后端登出接口')
        } catch (error) {
          console.warn('调用后端登出接口失败，继续清除本地状态:', error)
        }
      }

      clearLocalStorage()
      set(initialState)
    },

    // 仅清除本地状态，用于 401 错误等无需通知后端的场景。
    clearLocalState() {
      clearLocalStorage()
      set(initialState)
    },

    // 更新用户信息并同步本地缓存。
    updateUser(user: User) {
      update(state => {
        const newState = { ...state, user }

        if (deps.isBrowser()) {
          local.set(AUTH_USER_KEY, user)
        }

        return newState
      })
    },

    // 更新令牌对并同步本地缓存。
    updateTokens(accessToken: string, refreshToken: string, expiresIn: number) {
      const expiresAt = Date.now() + expiresIn * 1000

      update(state => {
        const newState = {
          ...state,
          accessToken,
          refreshToken,
          expiresAt
        }

        if (deps.isBrowser()) {
          local.set(AUTH_TOKEN_KEY, accessToken)
          local.set(AUTH_REFRESH_KEY, refreshToken)
          local.set(AUTH_EXPIRES_KEY, expiresAt)
        }

        return newState
      })
    },

    // 检查令牌是否仍然有效，过期前预留刷新窗口。
    isTokenValid(): boolean {
      if (!currentState.isAuthenticated || !currentState.accessToken || !currentState.expiresAt) {
        return false
      }

      return Date.now() < currentState.expiresAt - REFRESH_LEAD_TIME_MS
    },

    // 检查令牌是否进入需要刷新的窗口。
    shouldRefreshToken(): boolean {
      if (!currentState.isAuthenticated || !currentState.accessToken || !currentState.expiresAt) {
        return false
      }

      return Date.now() >= currentState.expiresAt - REFRESH_LEAD_TIME_MS
    },

    // 获取访问令牌
    getAccessToken(): string | null {
      return currentState.accessToken
    },

    // 获取刷新令牌
    getRefreshToken(): string | null {
      return currentState.refreshToken
    },

    // 获取当前用户权限列表，权限唯一权威为后端下发值。
    getPermissions(): string[] {
      return currentState.permissions
    },

    // 检查当前用户是否拥有指定权限，未下发时返回 false 交由调用方降级。
    hasPermission(permission: string): boolean {
      return currentState.permissions.includes(permission)
    },

    // 向后兼容：获取访问令牌
    getToken(): string | null {
      return currentState.accessToken
    }
  }
}
