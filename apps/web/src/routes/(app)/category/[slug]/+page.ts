import { ArticleAPI, CategoryAPI } from '$lib/api'
import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'

// 分类页每页数量，与目录页保持一致的版面节奏。
const CATEGORY_PAGE_SIZE = 12

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

export const load: PageLoad = async ({ params, url }) => {
  // 页码参数缺失或非法时回退到首页，负数与小数统一归一到合法页码。
  const requestedPage = Number(url.searchParams.get('page')) || 1
  const currentPage = Math.max(1, Math.floor(requestedPage))

  const categoryResponse = await CategoryAPI.getBySlug(params.slug)
  if (categoryResponse.code !== SUCCESS_CODE || !categoryResponse.data) {
    throw error(404, '分类不存在')
  }

  const listResponse = await ArticleAPI.byCategory(categoryResponse.data.id, {
    page: currentPage,
    pageSize: CATEGORY_PAGE_SIZE,
    sortBy: 'published_at',
    order: 'desc'
  })
  if (listResponse.code !== SUCCESS_CODE || !listResponse.data) {
    throw error(503, '文章列表加载失败，请稍后重试')
  }

  return { category: categoryResponse.data, list: listResponse.data, currentPage }
}
