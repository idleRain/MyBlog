import { createAuthStore } from '@myblog/auth'
import { browser } from '$app/environment'

// 认证 store 实例：注入环境判定与登出接口，核心逻辑统一下沉至 @myblog/auth。
export const authStore = createAuthStore({
  isBrowser: () => browser,
  logoutApi: async () => {
    // 动态导入避免与服务层形成循环依赖。
    const { UserAPI } = await import('$lib/api')
    await UserAPI.logout()
  }
})

export type { AuthState } from '@myblog/auth'
