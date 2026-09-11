import type { ArticleArchiveYear } from '@myblog/api/modules/article/types'
import type { Tag } from '@myblog/api/modules/tag/types'
import { RESPONSE_CODE_SUCCESS } from '@myblog/shared'
import { ArticleAPI, TagAPI } from '$lib/api'
import type { PageLoad } from './$types'
import { error } from '@sveltejs/kit'

// 首页无服务器端数据加载前显式关闭预渲染，接入业务数据后维持 SSR 按需渲染行为。
export const prerender = false

// 本期精选取热门文章榜首，热门侧栏与近期目录各取若干条。
const FEATURED_POPULAR_LIMIT = 4
const RECENT_POST_LIMIT = 6
const TAG_CLOUD_LIMIT = 8

export const load: PageLoad = async () => {
  // 近期与热门构成首页主体，任一失败即整页降级为服务错误。
  const [recentResponse, popularResponse, archivesResult, tagsResult] = await Promise.all([
    ArticleAPI.recent(RECENT_POST_LIMIT),
    ArticleAPI.popular(FEATURED_POPULAR_LIMIT),
    // 归档与标签为附属版面，失败时降级为空数据不阻断主体。
    ArticleAPI.archives().catch(() => null),
    TagAPI.getPopular(TAG_CLOUD_LIMIT).catch(() => null)
  ])

  if (
    recentResponse.code !== RESPONSE_CODE_SUCCESS ||
    !recentResponse.data ||
    popularResponse.code !== RESPONSE_CODE_SUCCESS ||
    !popularResponse.data
  ) {
    throw error(503, '首页数据加载失败，请稍后重试')
  }

  const archives: ArticleArchiveYear[] =
    archivesResult?.code === RESPONSE_CODE_SUCCESS ? (archivesResult.data ?? []) : []
  const tags: Tag[] =
    tagsResult?.code === RESPONSE_CODE_SUCCESS ? (tagsResult.data?.tags ?? []) : []

  return {
    recent: recentResponse.data.articles,
    popular: popularResponse.data.articles,
    archives,
    tags
  }
}
