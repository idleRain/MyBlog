import type { ApiResponse } from '@myblog/shared'

// 站点统计概览，字段与后端 StatsOverview 一致。
export interface StatsOverview {
  articleCount: number
  publishedCount: number
  totalViews: number
  totalLikes: number
  commentCount: number
  userCount: number
  categoryCount: number
  tagCount: number
}

// 文章浏览量趋势响应，缺失日期后端已补零。
export interface TrendResponse {
  dates: string[]
  values: number[]
}

// 各响应类型
export type StatsOverviewResponse = ApiResponse<StatsOverview>
export type TrendResponseData = ApiResponse<TrendResponse>
