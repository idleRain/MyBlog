import type { PageLoad } from './$types'

import { emptyCommentPageData } from './comment-page-state.svelte'
import { COMMENT_PAGE_SIZE } from '$lib/constants/comment'
import { CommentAPI } from '$lib/api'

// 评论管理页初始数据：固定加载第一页且不带筛选条件，与页面初始筛选状态一致。
// 加载失败时回退空数据并记录日志，由页面呈现空态，不中断路由。
export const load: PageLoad = async () => {
  const pageData = emptyCommentPageData()

  try {
    const response = await CommentAPI.adminList({ page: 1, pageSize: COMMENT_PAGE_SIZE })
    if (response.code === 200 && response.data) {
      pageData.comments = response.data.comments ?? []
      pageData.total = response.data.total ?? 0
    }
  } catch (error) {
    console.error('加载评论列表失败:', error)
  }

  return pageData
}
