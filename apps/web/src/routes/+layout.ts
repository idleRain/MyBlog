import type { FriendlyLink } from '@myblog/api/modules/friendlyLink/types'
import type { CategoryTreeNode } from '@myblog/api/modules/category/types'
import { CategoryAPI, FriendlyLinkAPI } from '$lib/api'
import type { LayoutLoad } from './$types'

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

export const load: LayoutLoad = async () => {
  // 全站导航数据：友链与分类均为附属信息，失败时降级为空列表不阻断页面。
  const [linksResult, categoriesResult] = await Promise.allSettled([
    FriendlyLinkAPI.list(),
    CategoryAPI.getTree()
  ])

  const friendlyLinks: FriendlyLink[] =
    linksResult.status === 'fulfilled' && linksResult.value.code === SUCCESS_CODE
      ? (linksResult.value.data?.links ?? [])
      : []

  const categories: CategoryTreeNode[] =
    categoriesResult.status === 'fulfilled' && categoriesResult.value.code === SUCCESS_CODE
      ? (categoriesResult.value.data?.tree ?? [])
      : []

  return { friendlyLinks, categories }
}
