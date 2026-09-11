import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'
import { ArticleAPI } from '$lib/api'

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

export const load: PageLoad = async ({ params }) => {
  const response = await ArticleAPI.getBySlug(params.slug)

  if (response.code !== SUCCESS_CODE || !response.data) {
    throw error(404, '文章不存在或尚未发布')
  }

  return { article: response.data }
}
