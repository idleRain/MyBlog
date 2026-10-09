import { fetchAuthenticated } from '$lib/service/ssr'
import { error, redirect } from '@sveltejs/kit'
import { createArticleAPI } from '@myblog/api'
import type { PageServerLoad } from './$types'

// 收藏列表每页数量，与其他目录页保持一致的版面节奏。
const FAVORITES_PAGE_SIZE = 12

// 未登录或会话缺失时统一跳转登录页，由登录页承接身份建立。
const LOGIN_PATH = '/login'

// 收藏读取失败时的页面级提示，与既有客户端补拉的文案保持一致。
const LOAD_FAILURE_MESSAGE = '收藏列表加载失败，请稍后重试'

/**
 * 收藏列表页在服务端取数。
 * 会话 Cookie 由 fetchAuthenticated 注入请求，后端据此识别登录身份；
 * 分页参数取自 URL 查询串，翻页时由 load 重新求值。
 */
export const load: PageServerLoad = async ({ request, url }) => {
  // 页码参数缺失或非法时回退到首页，负数与小数统一归一到合法页码。
  const requestedPage = Number(url.searchParams.get('page')) || 1
  const currentPage = Math.max(1, Math.floor(requestedPage))

  const result = await fetchAuthenticated(request.headers.get('cookie'), client => {
    const articleApi = createArticleAPI(client)
    return {
      call: () => articleApi.bookmarks({ page: currentPage, pageSize: FAVORITES_PAGE_SIZE })
    }
  })

  if (result.status === 'unauthorized') {
    throw redirect(302, LOGIN_PATH)
  }
  if (result.status === 'failed') {
    throw error(503, LOAD_FAILURE_MESSAGE)
  }

  return { list: result.data, currentPage }
}
