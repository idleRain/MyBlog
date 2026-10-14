import type { PublicProfile } from '@myblog/api/modules/user/types'
import { ArticleAPI, UserAPI } from '$lib/api'
import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'
import { m } from '$i18n'

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
    throw error(404, m['author:notFound']())
  }
  const profile: PublicProfile = profileResponse.data

  const listResponse = await ArticleAPI.byAuthor(authorId, {
    page: currentPage,
    pageSize: AUTHOR_PAGE_SIZE,
    sortBy: 'published_at',
    order: 'desc'
  })
  if (listResponse.code !== SUCCESS_CODE || !listResponse.data) {
    throw error(503, m['ui:error.listFailed']())
  }

  return { profile, list: listResponse.data, currentPage }
}
