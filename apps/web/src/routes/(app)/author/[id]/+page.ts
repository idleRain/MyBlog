import type { PublicProfile } from '@myblog/api/modules/user/types'
import { ArticleAPI, UserAPI } from '$lib/api'
import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'

// 作者页每页数量，与目录页保持一致的版面节奏。
const AUTHOR_PAGE_SIZE = 12

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

export const load: PageLoad = async ({ params, url }) => {
  const requestedPage = Number(url.searchParams.get('page')) || 1
  const currentPage = Math.max(1, Math.floor(requestedPage))
  const authorId = Number(params.id)

  const profileResponse = await UserAPI.getPublicProfile(authorId)
  if (profileResponse.code !== SUCCESS_CODE || !profileResponse.data) {
    throw error(404, '作者不存在')
  }
  const profile: PublicProfile = profileResponse.data

  const listResponse = await ArticleAPI.byAuthor(authorId, {
    page: currentPage,
    pageSize: AUTHOR_PAGE_SIZE,
    sortBy: 'published_at',
    order: 'desc'
  })
  if (listResponse.code !== SUCCESS_CODE || !listResponse.data) {
    throw error(503, '文章列表加载失败，请稍后重试')
  }

  return { profile, list: listResponse.data, currentPage }
}
