import type { CommentListData } from '@myblog/api/modules/comment/types'
import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import { renderMarkdown } from '$lib/markdown/render'
import { ArticleAPI, CommentAPI } from '$lib/api'
import type { PageServerLoad } from './$types'
import { error } from '@sveltejs/kit'

// 评论首屏加载数量，客户端的追加加载沿用该值。
const COMMENT_PAGE_SIZE = 20

export const load: PageServerLoad = async ({ params }) => {
  const response = await ArticleAPI.getBySlug(params.slug)

  if (response.code !== RESPONSE_CODE_SUCCESS || !response.data) {
    throw error(404, '文章不存在或尚未发布')
  }

  const article = response.data

  // 正文在前台渲染以接入 Shiki 双主题高亮，后端 contentHtml 缓存保持不变。
  const contentHtml = await renderMarkdown(article.content)

  // 评论首屏随 SSR 加载以利检索，加载失败时降级为空列表不阻断正文。
  let comments: CommentListData | null = null
  try {
    const commentsResponse = await CommentAPI.listByArticle(article.id, {
      page: 1,
      pageSize: COMMENT_PAGE_SIZE
    })
    if (commentsResponse.code === RESPONSE_CODE_SUCCESS) {
      comments = commentsResponse.data
    }
  } catch {
    comments = null
  }

  return { article, contentHtml, comments }
}
