import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'
import { ArticleAPI } from '$lib/api'

// 目录页每页数量，独立于后端默认值显式声明。
const BLOG_PAGE_SIZE = 12

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

export const load: PageLoad = async ({ url }) => {
  // 页码参数缺失或非法时回退到首页，负数与小数统一归一到合法页码。
  const requestedPage = Number(url.searchParams.get('page')) || 1
  const currentPage = Math.max(1, Math.floor(requestedPage))

  // 关键词来自标签云与搜索入口，空值时后端按全量列表处理。
  const search = url.searchParams.get('search') ?? ''

  const response = await ArticleAPI.list({
    page: currentPage,
    pageSize: BLOG_PAGE_SIZE,
    status: 'published',
    sortBy: 'published_at',
    order: 'desc',
    search
  })

  if (response.code !== SUCCESS_CODE || !response.data) {
    throw error(503, '文章列表加载失败，请稍后重试')
  }

  return { list: response.data, currentPage, search }
}
