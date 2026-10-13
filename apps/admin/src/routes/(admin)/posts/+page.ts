import type { PageLoad } from './$types'

import { emptyPostsPageData } from './posts-page-state.svelte'
import { ARTICLE_PAGE_SIZE } from '$lib/constants/article'
import { hasPermission } from '$lib/utils/permissions'
import { PERMISSIONS } from '$lib/constants/auth'
import { authStore } from '$lib/stores/auth'
import { ArticleAPI } from '$lib/api'

// 文章管理页初始数据：首屏加载第一页且不带筛选的文章列表。
// 状态筛选的可用性按后端登录下发的权限裁剪，权限读取走本地认证状态，
// 该状态在 store 创建时已同步水合，load 内不调用认证接口。
// 列表加载失败不抛错，回退空数据并由页面状态模块在后续交互中继续提示。
export const load: PageLoad = async () => {
  const pageData = emptyPostsPageData()
  pageData.canManageAllArticles = hasPermission(
    authStore.getCurrentState().user,
    PERMISSIONS.ARTICLE_MANAGE
  )

  try {
    const response = await ArticleAPI.list({
      page: 1,
      pageSize: ARTICLE_PAGE_SIZE,
      status: '',
      sortBy: 'created_at',
      order: 'desc'
    })
    if (response.code === 200 && response.data) {
      pageData.articles = response.data.articles ?? []
      pageData.total = response.data.total ?? 0
    } else {
      console.error('加载文章列表失败:', response.message)
    }
  } catch (error) {
    console.error('加载文章列表失败:', error)
  }
  return pageData
}
