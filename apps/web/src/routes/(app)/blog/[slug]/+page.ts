import type { CommentListData } from '@myblog/api/modules/comment/types'
import { ArticleAPI, CommentAPI } from '$lib/api'
import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 评论首屏加载数量，客户端的追加加载沿用该值。
const COMMENT_PAGE_SIZE = 20

export const load: PageLoad = async ({ params }) => {
  const response = await ArticleAPI.getBySlug(params.slug)

  if (response.code !== SUCCESS_CODE || !response.data) {
    throw error(404, '文章不存在或尚未发布')
  }

  // 评论首屏随 SSR 加载以利检索，加载失败时降级为空列表不阻断正文。
  let comments: CommentListData | null = null
  try {
    const commentsResponse = await CommentAPI.listByArticle(response.data.id, {
      page: 1,
      pageSize: COMMENT_PAGE_SIZE
    })
    if (commentsResponse.code === SUCCESS_CODE) {
      comments = commentsResponse.data
    }
  } catch {
    comments = null
  }

  return { article: response.data, comments }
}
