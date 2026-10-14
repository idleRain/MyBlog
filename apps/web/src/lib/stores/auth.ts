import { createAuthStore } from '@myblog/auth'

// 登出经动态导入调用，避免认证 store 与服务层形成循环依赖；
// 会话令牌经 Cookie 携带，登出由服务端撤销令牌并清除 Cookie。
export const authStore = createAuthStore({
  logoutApi: async () => {
    const { UserAPI } = await import('$lib/api')
    await UserAPI.logout()
  }
})
