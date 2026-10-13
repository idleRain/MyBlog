import type { StatsOverview } from '@myblog/api/modules/stats/types'
import { DEFAULT_TREND_DAYS } from '$lib/constants/stats'
import { toast } from 'svelte-sonner'
import { StatsAPI } from '$lib/api'

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface StatsPageData {
  overview: StatsOverview | null
  trendDates: string[]
  trendValues: number[]
}

// 空页面数据，load 失败时使用。
export function emptyStatsPageData(): StatsPageData {
  return { overview: null, trendDates: [], trendValues: [] }
}

// 站点统计页状态：集中持有概览、趋势数据与全部读写操作，页面组件只做渲染。
class StatsPageState {
  overview = $state<StatsOverview | null>(null)
  isLoading = $state(false)
  // 趋势天数以字符串承载，适配 ToggleGroup 的字符串值约束。
  trendDays = $state<string>(DEFAULT_TREND_DAYS)
  trendDates = $state<string[]>([])
  trendValues = $state<number[]>([])

  // 用 load 结果初始化页面状态，重复挂载时以最新数据覆盖。
  initialize(data: StatsPageData) {
    this.overview = data.overview
    this.trendDates = data.trendDates
    this.trendValues = data.trendValues
  }

  // 加载统计概览与浏览量趋势，趋势按当前选择的天数取数。
  async loadStats() {
    this.isLoading = true
    try {
      const [overviewResult, trendResult] = await Promise.all([
        StatsAPI.getOverview(),
        StatsAPI.getArticleViewsTrend(Number(this.trendDays))
      ])
      if (overviewResult.code === 200 && overviewResult.data) {
        this.overview = overviewResult.data
      }
      if (trendResult.code === 200 && trendResult.data) {
        this.trendDates = trendResult.data.dates ?? []
        this.trendValues = trendResult.data.values ?? []
      }
    } catch (error) {
      console.error('加载统计数据失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 趋势天数变更后按新天数重新加载。
  setTrendDays(days: string) {
    this.trendDays = days
    void this.loadStats()
  }
}

// 页面级共享状态实例。
export const statsPageState = new StatsPageState()
