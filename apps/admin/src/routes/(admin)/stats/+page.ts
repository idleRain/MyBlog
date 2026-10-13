import type { PageLoad } from './$types'

import { emptyStatsPageData } from './stats-page-state.svelte'
import { DEFAULT_TREND_DAYS } from '$lib/constants/stats'
import { StatsAPI } from '$lib/api'

// 站点统计页初始数据：首屏并行加载概览与默认天数的浏览量趋势。
// 加载失败不抛错，回退空数据并由页面状态模块在后续交互中继续提示。
export const load: PageLoad = async () => {
  const pageData = emptyStatsPageData()

  const [overviewResult, trendResult] = await Promise.allSettled([
    StatsAPI.getOverview(),
    StatsAPI.getArticleViewsTrend(Number(DEFAULT_TREND_DAYS))
  ])

  if (overviewResult.status === 'fulfilled' && overviewResult.value.code === 200) {
    pageData.overview = overviewResult.value.data ?? null
  } else if (overviewResult.status === 'rejected') {
    console.error('加载统计概览失败:', overviewResult.reason)
  }

  if (trendResult.status === 'fulfilled' && trendResult.value.code === 200) {
    pageData.trendDates = trendResult.value.data?.dates ?? []
    pageData.trendValues = trendResult.value.data?.values ?? []
  } else if (trendResult.status === 'rejected') {
    console.error('加载浏览量趋势失败:', trendResult.reason)
  }

  return pageData
}
