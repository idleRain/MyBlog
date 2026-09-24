import type { User } from '@myblog/api/modules/user/types'
import { writable } from 'svelte/store'
import { local } from '@myblog/shared'

// 认证状态接口。
// 会话令牌经 HttpOnly Cookie 由浏览器自动携带，前端不再持有任何令牌，
// 状态仅承载用于界面渲染的用户信息与权限列表。
export interface AuthState {
  isAuthenticated: boolean
  user: User | null
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
   * 调用后端登出接口，服务端撤销令牌并清除会话 Cookie。
   */
  logoutApi: () => Promise<void>
}

// 初始状态
const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  permissions: []
}

// localStorage 用户信息存储键
const AUTH_USER_KEY = 'auth_user'
const AUTH_PERMISSIONS_KEY = 'auth_permissions'

// 认证状态占用的全部存储键，跨标签页同步据此过滤与认证无关的存储变更。
const AUTH_STORAGE_KEYS = new Set([AUTH_USER_KEY, AUTH_PERMISSIONS_KEY])

/**
 * 判断两份认证状态是否等价，避免跨标签页同步触发无意义的状态写入与重渲染。
 */
function isSameAuthState(left: AuthState, right: AuthState): boolean {
  return (
    left.isAuthenticated === right.isAuthenticated &&
    left.user?.id === right.user?.id &&
    left.user?.updatedAt === right.user?.updatedAt
  )
}

/**
 * 创建认证 store 实例，供各应用各持有一份，避免全局单例造成状态串扰。
 */
export function createAuthStore(deps: AuthStoreDeps) {
  // 从 localStorage 加载初始状态。
  // 仅凭本地用户信息判定登录态，会话 Cookie 是否有效由首次请求的服务端裁决。
  function loadInitialState(): AuthState {
    if (!deps.isBrowser()) return initialState

    try {
      const user = local.get<User>(AUTH_USER_KEY)
      const permissions = local.get<string[]>(AUTH_PERMISSIONS_KEY)

      if (user) {
        return {
          isAuthenticated: true,
          user,
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

  // 进行中的续期任务，并发触发续期时共享同一次请求。
  // 续期即旋转语义下，并发各自续期会导致后到的请求携带已被撤销的旧 Cookie 而失败。
  let refreshInFlight: Promise<unknown> | null = null

  // 清除本地存储中的全部认证状态。
  function clearLocalStorage() {
    if (!deps.isBrowser()) return
    local.rm(AUTH_USER_KEY)
    local.rm(AUTH_PERMISSIONS_KEY)
  }

  /**
   * 从持久化存储重新加载认证状态，用于跨标签页对齐。
   * 同源标签页共享 localStorage 但各自持有独立的内存状态，
   * 任一标签页登出后，其余标签页必须跟随，避免展示已失效的登录态。
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

    // 登录成功后记录用户信息并持久化；会话令牌已由服务端写入 HttpOnly Cookie。
    login(user: User, permissions: string[] = []) {
      const authState: AuthState = {
        isAuthenticated: true,
        user,
        permissions
      }

      if (deps.isBrowser()) {
        local.set(AUTH_USER_KEY, user)
        local.set(AUTH_PERMISSIONS_KEY, permissions)
      }

      set(authState)
    },

    // 登出，先尝试撤销服务端会话再清除本地状态。
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

    // 从持久化存储对齐认证状态，用于跨标签页同步。
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

    // 以单飞模式执行会话续期，并发调用共享同一次请求。
    // 续期完成后清除进行中标记，失败结果同样共享，由调用方按失败处理。
    refreshSingleFlight<T>(refreshFn: () => Promise<T>): Promise<T> {
      if (!refreshInFlight) {
        refreshInFlight = refreshFn().finally(() => {
          refreshInFlight = null
        })
      }
      return refreshInFlight as Promise<T>
    },

    // 获取当前用户权限列表，权限唯一权威为后端下发值。
    getPermissions(): string[] {
      return currentState.permissions
    },

    // 检查当前用户是否拥有指定权限，未下发时返回 false 交由调用方降级。
    hasPermission(permission: string): boolean {
      return currentState.permissions.includes(permission)
    }
  }
}

/**
 * 认证 store 实例类型，供消费方声明注入依赖。
 */
export type AuthStore = ReturnType<typeof createAuthStore>
