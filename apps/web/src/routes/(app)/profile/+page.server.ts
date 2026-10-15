import { fetchAuthenticated } from '$lib/service/ssr'
import { error, redirect } from '@sveltejs/kit'
import type { PageServerLoad } from './$types'
import { createUserAPI } from '@myblog/api'
import { m } from '$i18n'

// 未登录或会话缺失时统一跳转登录页，由登录页承接身份建立。
const LOGIN_PATH = '/login'

/**
 * 个人资料页在服务端取数。
 * 会话 Cookie 由 fetchAuthenticated 注入请求，后端据此识别登录身份，
 * 因此首屏直接携带资料而不存在「先渲染容器再补拉」的两段式。
 */
export const load: PageServerLoad = async ({ request }) => {
  const result = await fetchAuthenticated(request.headers.get('cookie'), client => {
    const userApi = createUserAPI(client)
    return { call: () => userApi.getProfile() }
  })

  if (result.status === 'unauthorized') {
    throw redirect(302, LOGIN_PATH)
  }
  if (result.status === 'failed') {
    throw error(503, m['account:profile.loadFailed']())
  }

  return { profile: result.data }
}
