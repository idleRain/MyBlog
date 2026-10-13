import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Article } from '@myblog/api/modules/article/types'

import { ARTICLE_PAGE_SIZE } from '$lib/constants/article'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const articleApiMock = {
  list: vi.fn(),
  delete: vi.fn(),
  publish: vi.fn(),
  unpublish: vi.fn(),
  archive: vi.fn(),
  private: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn()
}

vi.mock('$lib/api', () => ({ ArticleAPI: articleApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { emptyPostsPageData, postsPageState } = await import('../posts-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 搜索防抖时长，与页面状态模块内的取值一致。
const SEARCH_DEBOUNCE_MS = 300

// 构造一篇文章替身，字段取值只覆盖被测逻辑关心的部分。
function makeArticle(id: number, status = 'draft'): Article {
  return { id, title: `文章${id}`, status } as unknown as Article
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  postsPageState.initialize(emptyPostsPageData())
  postsPageState.search = ''
  postsPageState.status = ''
  postsPageState.sortBy = 'created_at'
  postsPageState.order = 'desc'
  postsPageState.currentPage = 1
  postsPageState.isLoading = false
}

describe('文章管理页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果填充列表、总数与权限标记', () => {
    const articles = [makeArticle(1), makeArticle(2, 'published')]

    postsPageState.initialize({ articles, total: 2, canManageAllArticles: true })

    expect(postsPageState.articles).toEqual(articles)
    expect(postsPageState.total).toBe(2)
    expect(postsPageState.canManageAllArticles).toBe(true)
  })

  it('空数据下无生效筛选条件且总页数至少为 1', () => {
    expect(postsPageState.articles).toEqual([])
    expect(postsPageState.canManageAllArticles).toBe(false)
    expect(postsPageState.hasActiveFilters).toBe(false)
    expect(postsPageState.totalPages).toBe(1)
  })
})

describe('文章管理页状态：列表加载', () => {
  beforeEach(resetState)

  it('默认携带分页、空状态与默认排序参数', async () => {
    articleApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { articles: [makeArticle(1)], total: 1 }
    })

    await postsPageState.loadArticles()

    expect(articleApiMock.list).toHaveBeenCalledWith({
      page: 1,
      pageSize: ARTICLE_PAGE_SIZE,
      status: '',
      sortBy: 'created_at',
      order: 'desc'
    })
    expect(postsPageState.articles).toHaveLength(1)
    expect(postsPageState.total).toBe(1)
  })

  it('搜索词去除首尾空白后作为查询条件', async () => {
    articleApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { articles: [], total: 0 } })
    postsPageState.search = '  博客  '
    postsPageState.status = 'published'

    await postsPageState.loadArticles()

    expect(articleApiMock.list).toHaveBeenCalledWith({
      page: 1,
      pageSize: ARTICLE_PAGE_SIZE,
      status: 'published',
      sortBy: 'created_at',
      order: 'desc',
      search: '博客'
    })
  })

  it('搜索词为纯空白时不携带查询条件', async () => {
    articleApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { articles: [], total: 0 } })
    postsPageState.search = '   '

    await postsPageState.loadArticles()

    expect(articleApiMock.list).toHaveBeenCalledWith({
      page: 1,
      pageSize: ARTICLE_PAGE_SIZE,
      status: '',
      sortBy: 'created_at',
      order: 'desc'
    })
  })

  it('业务码非成功时提示后端文案且不清空现有列表', async () => {
    articleApiMock.list.mockResolvedValue({ code: 500, message: '查询失败' })
    postsPageState.initialize({
      articles: [makeArticle(1)],
      total: 1,
      canManageAllArticles: false
    })

    await postsPageState.loadArticles()

    expect(toastMock.error).toHaveBeenCalledWith('查询失败')
    expect(postsPageState.articles).toHaveLength(1)
    expect(postsPageState.isLoading).toBe(false)
  })

  it('请求抛错时提示网络错误并复位加载态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    articleApiMock.list.mockRejectedValue(new Error('断网'))

    await postsPageState.loadArticles()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(postsPageState.isLoading).toBe(false)
    consoleError.mockRestore()
  })
})

describe('文章管理页状态：筛选与分页', () => {
  beforeEach(resetState)

  it('筛选变更后回到第一页并重新加载', async () => {
    articleApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { articles: [], total: 0 } })
    postsPageState.currentPage = 4

    postsPageState.applyFilterChange()

    expect(postsPageState.currentPage).toBe(1)
    await vi.waitFor(() => expect(articleApiMock.list).toHaveBeenCalledTimes(1))
  })

  it('重置筛选清空全部条件并回到第一页', async () => {
    articleApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { articles: [], total: 0 } })
    postsPageState.search = '博客'
    postsPageState.status = 'archived'
    postsPageState.sortBy = 'view_count'
    postsPageState.order = 'asc'
    postsPageState.currentPage = 5

    postsPageState.resetFilters()

    expect(postsPageState.search).toBe('')
    expect(postsPageState.status).toBe('')
    expect(postsPageState.sortBy).toBe('created_at')
    expect(postsPageState.order).toBe('desc')
    expect(postsPageState.currentPage).toBe(1)
    expect(postsPageState.hasActiveFilters).toBe(false)
    await vi.waitFor(() => expect(articleApiMock.list).toHaveBeenCalledTimes(1))
  })

  it('切换页码后按目标页加载', async () => {
    articleApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { articles: [], total: ARTICLE_PAGE_SIZE * 5 }
    })

    postsPageState.changePage(2)

    await vi.waitFor(() =>
      expect(articleApiMock.list).toHaveBeenCalledWith({
        page: 2,
        pageSize: ARTICLE_PAGE_SIZE,
        status: '',
        sortBy: 'created_at',
        order: 'desc'
      })
    )
    expect(postsPageState.totalPages).toBe(5)
  })
})

describe('文章管理页状态：搜索防抖', () => {
  beforeEach(() => {
    resetState()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  // 连续输入必须在停顿后只请求一次，且回到第一页，避免翻页后搜索命中空页。
  it('连续输入只在停顿后触发一次请求并回到第一页', () => {
    articleApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { articles: [], total: 0 } })
    postsPageState.currentPage = 3

    postsPageState.applySearchDebounced()
    postsPageState.applySearchDebounced()
    expect(articleApiMock.list).not.toHaveBeenCalled()

    vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS)

    expect(articleApiMock.list).toHaveBeenCalledTimes(1)
    expect(postsPageState.currentPage).toBe(1)
  })
})

describe('文章管理页状态：状态流转与删除', () => {
  beforeEach(resetState)

  it('状态流转成功后提示后端消息并刷新列表', async () => {
    articleApiMock.publish.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { message: '文章已发布' }
    })
    articleApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { articles: [], total: 0 } })

    await postsPageState.applyStatusAction(makeArticle(1), 'publish')

    expect(articleApiMock.publish).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('文章已发布')
    expect(articleApiMock.list).toHaveBeenCalledTimes(1)
  })

  it('状态流转缺少业务消息时使用默认提示', async () => {
    articleApiMock.archive.mockResolvedValue({ code: SUCCESS_CODE, data: null })
    articleApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { articles: [], total: 0 } })

    await postsPageState.applyStatusAction(makeArticle(1), 'archive')

    expect(toastMock.success).toHaveBeenCalledWith('操作成功')
  })

  it('状态流转失败时提示后端文案且不刷新列表', async () => {
    articleApiMock.private.mockResolvedValue({ code: 403, message: '无权操作' })

    await postsPageState.applyStatusAction(makeArticle(1), 'private')

    expect(toastMock.error).toHaveBeenCalledWith('无权操作')
    expect(articleApiMock.list).not.toHaveBeenCalled()
  })

  it('状态流转请求抛错时提示网络错误', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    articleApiMock.unpublish.mockRejectedValue(new Error('断网'))

    await postsPageState.applyStatusAction(makeArticle(1, 'published'), 'unpublish')

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    consoleError.mockRestore()
  })

  it('删除成功后提示并刷新列表', async () => {
    articleApiMock.delete.mockResolvedValue({ code: SUCCESS_CODE })
    articleApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { articles: [], total: 0 } })

    await postsPageState.removeArticle(makeArticle(1))

    expect(articleApiMock.delete).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('文章删除成功')
    expect(articleApiMock.list).toHaveBeenCalledTimes(1)
  })

  it('删除失败时提示后端文案且不刷新列表', async () => {
    articleApiMock.delete.mockResolvedValue({ code: 404, message: '文章不存在' })

    await postsPageState.removeArticle(makeArticle(1))

    expect(toastMock.error).toHaveBeenCalledWith('文章不存在')
    expect(articleApiMock.list).not.toHaveBeenCalled()
  })
})
