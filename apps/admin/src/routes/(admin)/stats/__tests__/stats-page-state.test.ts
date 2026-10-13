import type { StatsOverview } from '@myblog/api/modules/stats/types'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { DEFAULT_TREND_DAYS } from '$lib/constants/stats'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const statsApiMock = {
  getOverview: vi.fn(),
  getArticleViewsTrend: vi.fn()
}

const toastMock = {
  error: vi.fn()
}

vi.mock('$lib/api', () => ({ StatsAPI: statsApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { emptyStatsPageData, statsPageState } = await import('../stats-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 构造统计概览替身，字段取值只覆盖被测逻辑关心的部分。
function makeOverview(articleCount = 3): StatsOverview {
  return { articleCount } as unknown as StatsOverview
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  statsPageState.initialize(emptyStatsPageData())
  statsPageState.trendDays = DEFAULT_TREND_DAYS
  statsPageState.isLoading = false
}

describe('站点统计页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果填充概览与趋势数据', () => {
    const overview = makeOverview(8)

    statsPageState.initialize({ overview, trendDates: ['2026-01-01'], trendValues: [12] })

    expect(statsPageState.overview).toEqual(overview)
    expect(statsPageState.trendDates).toEqual(['2026-01-01'])
    expect(statsPageState.trendValues).toEqual([12])
  })

  it('空数据下概览为空且趋势天数取默认值', () => {
    expect(statsPageState.overview).toBeNull()
    expect(statsPageState.trendDays).toBe(DEFAULT_TREND_DAYS)
    expect(statsPageState.isLoading).toBe(false)
  })
})

describe('站点统计页状态：加载', () => {
  beforeEach(resetState)

  it('概览与趋势按当前天数并行取数', async () => {
    statsApiMock.getOverview.mockResolvedValue({ code: SUCCESS_CODE, data: makeOverview() })
    statsApiMock.getArticleViewsTrend.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { dates: [], values: [] }
    })

    await statsPageState.loadStats()

    expect(statsApiMock.getOverview).toHaveBeenCalledTimes(1)
    expect(statsApiMock.getArticleViewsTrend).toHaveBeenCalledWith(Number(DEFAULT_TREND_DAYS))
  })

  it('加载成功更新概览与趋势数据', async () => {
    statsApiMock.getOverview.mockResolvedValue({ code: SUCCESS_CODE, data: makeOverview(9) })
    statsApiMock.getArticleViewsTrend.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { dates: ['2026-01-01'], values: [5] }
    })

    await statsPageState.loadStats()

    expect(statsPageState.overview?.articleCount).toBe(9)
    expect(statsPageState.trendDates).toEqual(['2026-01-01'])
    expect(statsPageState.trendValues).toEqual([5])
    expect(statsPageState.isLoading).toBe(false)
  })

  it('趋势响应缺失字段时按空数组处理', async () => {
    statsApiMock.getOverview.mockResolvedValue({ code: SUCCESS_CODE, data: makeOverview() })
    statsApiMock.getArticleViewsTrend.mockResolvedValue({ code: SUCCESS_CODE, data: {} })

    await statsPageState.loadStats()

    expect(statsPageState.trendDates).toEqual([])
    expect(statsPageState.trendValues).toEqual([])
  })

  it('业务码非成功时保留现有数据且不提示', async () => {
    statsApiMock.getOverview.mockResolvedValue({ code: 500, message: '查询失败' })
    statsApiMock.getArticleViewsTrend.mockResolvedValue({ code: 500, message: '查询失败' })
    statsPageState.initialize({ overview: makeOverview(2), trendDates: ['d'], trendValues: [1] })

    await statsPageState.loadStats()

    expect(statsPageState.overview?.articleCount).toBe(2)
    expect(statsPageState.trendDates).toEqual(['d'])
    expect(toastMock.error).not.toHaveBeenCalled()
  })

  it('请求抛错时提示网络错误并复位加载态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    statsApiMock.getOverview.mockRejectedValue(new Error('断网'))
    statsApiMock.getArticleViewsTrend.mockRejectedValue(new Error('断网'))

    await statsPageState.loadStats()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(statsPageState.isLoading).toBe(false)
    consoleError.mockRestore()
  })
})

describe('站点统计页状态：趋势天数切换', () => {
  beforeEach(resetState)

  it('切换天数后按新天数重新加载', async () => {
    statsApiMock.getOverview.mockResolvedValue({ code: SUCCESS_CODE, data: makeOverview() })
    statsApiMock.getArticleViewsTrend.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { dates: [], values: [] }
    })

    statsPageState.setTrendDays('30')

    expect(statsPageState.trendDays).toBe('30')
    await vi.waitFor(() => expect(statsApiMock.getArticleViewsTrend).toHaveBeenCalledWith(30))
  })
})
