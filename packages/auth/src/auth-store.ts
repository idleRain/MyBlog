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
   * 刷新令牌可选传入，后端将访问与刷新令牌一并撤销。
   */
  logoutApi: (refreshToken?: string) => Promise<void>
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

// 认证状态占用的全部存储键，跨标签页同步据此过滤与认证无关的存储变更。
const AUTH_STORAGE_KEYS = new Set([
  AUTH_TOKEN_KEY,
  AUTH_REFRESH_KEY,
  AUTH_USER_KEY,
  AUTH_EXPIRES_KEY,
  AUTH_PERMISSIONS_KEY
])

// 令牌过期前提前刷新的窗口，单位毫秒
const REFRESH_LEAD_TIME_MS = 5 * 60 * 1000

/**
 * 判断两份认证状态是否等价，避免跨标签页同步触发无意义的状态写入与重渲染。
 * 权限与用户信息随令牌一同写入，令牌与身份一致时视为等价。
 */
function isSameAuthState(left: AuthState, right: AuthState): boolean {
  return (
    left.isAuthenticated === right.isAuthenticated &&
    left.accessToken === right.accessToken &&
    left.refreshToken === right.refreshToken &&
    left.expiresAt === right.expiresAt &&
    left.user?.id === right.user?.id
  )
}

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

  // 进行中的刷新任务，并发触发刷新时共享同一次请求。
  // 刷新即旋转语义下，并发各自刷新会导致后到的请求携带已撤销的旧令牌而失败。
  let refreshInFlight: Promise<string | null> | null = null

  // 清除本地存储中的全部认证令牌。
  function clearLocalStorage() {
    if (!deps.isBrowser()) return
    local.rm(AUTH_TOKEN_KEY)
    local.rm(AUTH_REFRESH_KEY)
    local.rm(AUTH_USER_KEY)
    local.rm(AUTH_EXPIRES_KEY)
    local.rm(AUTH_PERMISSIONS_KEY)
  }

  /**
   * 从持久化存储重新加载认证状态，用于跨标签页对齐。
   * 同源标签页共享 localStorage 但各自持有独立的内存状态，
   * 任一标签页登出或旋转令牌后，其余标签页必须跟随，否则会继续使用已被撤销的旧令牌。
   */
  function syncFromStorage() {
    if (!deps.isBrowser()) return

    const nextState = loadInitialState()
    if (isSameAuthState(currentState, nextState)) return

    set(nextState)
  }

  // 注册跨标签页同步监听。storage 事件仅在写入方之外的同源标签页触发，因此不会形成回环。
  if (deps.isBrowser()) {
    window.addEventListener('storage', event => {
      // 整库清空时 event.key 为 null，此时无法按键过滤，同样需要重新加载。
      if (event.key === null || AUTH_STORAGE_KEYS.has(event.key)) {
        syncFromStorage()
      }
    })
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

    // 登出，先尝试撤销服务端令牌对再清除本地状态。
    async logout(skipApiCall: boolean = false) {
      if (!skipApiCall && deps.isBrowser() && currentState.isAuthenticated) {
        try {
          await deps.logoutApi(currentState.refreshToken ?? undefined)
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

    // 从持久化存储对齐认证状态，供刷新前读取其他标签页可能已写入的最新令牌。
    syncFromStorage,

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

    // 以单飞模式执行令牌刷新，并发调用共享同一次刷新结果。
    // 刷新完成后清除进行中标记，失败结果同样共享，由调用方按 null 处理。
    refreshSingleFlight(refreshFn: () => Promise<string | null>): Promise<string | null> {
      if (!refreshInFlight) {
        refreshInFlight = refreshFn().finally(() => {
          refreshInFlight = null
        })
      }
      return refreshInFlight
    },

    // 检查令牌是否仍然有效，过期前预留刷新窗口。
    isTokenValid(): boolean {
      if (!currentState.isAuthenticated || !currentState.accessToken || !currentState.expiresAt) {
        return false
      }

      return Date.now() < currentState.expiresAt - REFRESH_LEAD_TIME_MS
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

/**
 * 认证 store 实例类型，供令牌刷新编排等消费方声明注入依赖。
 */
export type AuthStore = ReturnType<typeof createAuthStore>
