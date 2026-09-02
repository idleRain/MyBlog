import type { KyInstance } from 'ky'
import type { StatsOverviewResponse, TrendResponseData } from './types.ts'

/**
 * 创建站点统计接口模块，依赖注入的 http 客户端由调用方提供。
 */
export function createStatsAPI(request: KyInstance) {
  return {
    // 获取站点统计概览。
    getOverview(): Promise<StatsOverviewResponse> {
      return request.post('admin/stats/overview', { json: {} }).json()
    },

    // 获取文章浏览量趋势，days 取值范围 1-90，默认 7。
    getArticleViewsTrend(days = 7): Promise<TrendResponseData> {
      return request.post('admin/stats/articles', { json: { days } }).json()
    }
  }
}

export type StatsAPI = ReturnType<typeof createStatsAPI>

export type { StatsOverviewResponse, TrendResponseData }
