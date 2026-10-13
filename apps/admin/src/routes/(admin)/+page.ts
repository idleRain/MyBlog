import type { PageLoad } from './$types'

import type { StatsOverview } from '@myblog/api/modules/stats/types'
import type { Article } from '@myblog/api/modules/article/types'
import type { UserRole } from '@myblog/api/modules/user/types'
import { hasPermission } from '$lib/utils/permissions'
import { PERMISSIONS } from '$lib/constants/auth'
import { ArticleAPI, StatsAPI } from '$lib/api'
import { authStore } from '$lib/stores/auth'

// 统计概览与趋势的取数天数，与区块标题「近 7 天」一致。
const DASHBOARD_TREND_DAYS = 7

// 最新文章区块的展示条数。
const RECENT_ARTICLE_LIMIT = 5

// 仪表盘页面数据，区块可见性标记与各区块首屏数据一并经 load 注入。
interface DashboardPageData {
  userRole: UserRole
  nickname: string
  canViewArticles: boolean
  canCreateArticle: boolean
  canManageUsers: boolean
  overview: StatsOverview | null
  trendDates: string[]
  trendValues: number[]
  recentArticles: Article[]
}

// 统计区块取数。概览与趋势任一失败只影响本区块，其余区块照常渲染。
async function loadStatsSection(): Promise<
  Pick<DashboardPageData, 'overview' | 'trendDates' | 'trendValues'>
> {
  const section: Pick<DashboardPageData, 'overview' | 'trendDates' | 'trendValues'> = {
    overview: null,
    trendDates: [],
    trendValues: []
  }

  const [overviewResult, trendResult] = await Promise.allSettled([
    StatsAPI.getOverview(),
    StatsAPI.getArticleViewsTrend(DASHBOARD_TREND_DAYS)
  ])

  if (overviewResult.status === 'fulfilled' && overviewResult.value.code === 200) {
    section.overview = overviewResult.value.data ?? null
  } else if (overviewResult.status === 'rejected') {
    console.error('加载统计概览失败:', overviewResult.reason)
  }

  if (trendResult.status === 'fulfilled' && trendResult.value.code === 200) {
    section.trendDates = trendResult.value.data?.dates ?? []
    section.trendValues = trendResult.value.data?.values ?? []
  } else if (trendResult.status === 'rejected') {
    console.error('加载浏览量趋势失败:', trendResult.reason)
  }

  return section
}

// 最新文章区块取数。该列表对具备 article:list 权限的角色统一走 /api/articles/list，
// 状态可见性由后端判定，前端不再按角色裁剪状态。
async function loadRecentArticles(): Promise<Article[]> {
  try {
    const response = await ArticleAPI.list({ page: 1, pageSize: RECENT_ARTICLE_LIMIT })
    if (response.code === 200 && response.data) {
      return response.data.articles ?? []
    }
    console.error('加载最新文章失败:', response.message)
  } catch (error) {
    console.error('加载最新文章失败:', error)
  }
  return []
}

// 仪表盘初始数据。区块可见性按后端登录下发的权限列表裁剪，权限读取走本地认证状态，
// 该状态在 store 创建时已同步水合，load 内不调用认证接口。
export const load: PageLoad = async () => {
  const currentUser = authStore.getCurrentState().user
  const canViewStats = hasPermission(currentUser, PERMISSIONS.SYSTEM_STATS)
  const canViewArticles = hasPermission(currentUser, PERMISSIONS.ARTICLE_LIST)

  const statsSection = canViewStats
    ? await loadStatsSection()
    : { overview: null, trendDates: [], trendValues: [] }
  const recentArticles = canViewArticles ? await loadRecentArticles() : []

  const pageData: DashboardPageData = {
    userRole: currentUser?.role ?? 'user',
    nickname: currentUser?.nickname ?? '',
    canViewArticles,
    canCreateArticle: hasPermission(currentUser, PERMISSIONS.ARTICLE_CREATE),
    canManageUsers: hasPermission(currentUser, PERMISSIONS.USER_LIST),
    ...statsSection,
    recentArticles
  }

  return pageData
}
