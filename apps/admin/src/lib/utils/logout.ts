// 退出登录工具函数

import { goto } from '$lib/utils/navigation'
import { authStore } from '$lib/stores/auth'
import { browser } from '$app/environment'
import { toast } from 'svelte-sonner'

/**
 * 执行完整的退出登录流程
 * 1. 调用后端 logout 接口使 token 失效
 * 2. 清除本地认证状态
 * 3. 跳转到登录页面
 */
export async function performLogout(
  options: {
    showToast?: boolean
    redirectTo?: string | null
    skipApiCall?: boolean
  } = {}
): Promise<boolean> {
  const { showToast = true, redirectTo = '/login', skipApiCall = false } = options

  let apiCallSuccess = false

  try {
    if (!skipApiCall) {
      // 调用后端登出接口
      await authStore.logout(false) // 不跳过 API 调用
      apiCallSuccess = true
      console.log('成功调用后端登出接口')
    } else {
      // 仅清除本地状态
      authStore.clearLocalState()
      console.log('已清除本地认证状态')
    }

    if (showToast && browser) {
      if (apiCallSuccess) {
        toast.success('已成功退出登录')
      } else {
        toast.info('已清除本地登录状态')
      }
    }

    return true
  } catch (error) {
    console.error('退出登录失败:', error)

    // 即使后端调用失败，也要清除本地状态
    authStore.clearLocalState()

    if (showToast && browser) {
      toast.error('退出登录时发生错误，但已清除本地状态')
    }

    return false
  } finally {
    // 无论成功失败，都跳转到指定页面
    if (browser && redirectTo) {
      await goto(redirectTo)
    }
  }
}

/**
 * 强制退出登录（仅清除本地状态）
 * 用于 401 错误等无法调用后端接口的场景
 */
