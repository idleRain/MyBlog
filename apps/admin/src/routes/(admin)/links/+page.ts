import type { PageLoad } from './$types'

import { emptyLinkPageData } from './link-page-state.svelte'
import { LINK_PAGE_SIZE } from '$lib/constants/link'
import { FriendlyLinkAPI } from '$lib/api'

// 友链管理页初始数据：首屏第一页列表，筛选条件为空。
// 页面由 load 注入的数据渲染，加载失败回退空数据并记录日志，不阻断页面挂载。
export const load: PageLoad = async () => {
  const pageData = emptyLinkPageData()

  try {
    const response = await FriendlyLinkAPI.adminList({ page: 1, pageSize: LINK_PAGE_SIZE })
    if (response.code === 200 && response.data) {
      pageData.links = response.data.links ?? []
      pageData.total = response.data.total ?? 0
    } else {
      console.error('加载友链列表失败:', response.message)
    }
  } catch (error) {
    console.error('加载友链列表失败:', error)
  }

  return pageData
}
