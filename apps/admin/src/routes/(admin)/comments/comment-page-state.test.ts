import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { AdminComment } from '@myblog/api/modules/comment/types'

import { COMMENT_PAGE_SIZE } from '$lib/constants/comment'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const commentApiMock = {
  adminList: vi.fn(),
  approve: vi.fn(),
  reject: vi.fn(),
  markSpam: vi.fn(),
  trash: vi.fn(),
  delete: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn()
}

vi.mock('$lib/api', () => ({ CommentAPI: commentApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { emptyCommentPageData, commentPageState } = await import('./comment-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 搜索防抖时长，与页面状态模块内的取值一致。
const SEARCH_DEBOUNCE_MS = 300

// 构造一条评论替身，字段取值只覆盖被测逻辑关心的部分。
function makeComment(id: number, status = 'pending'): AdminComment {
  return {
    id,
    status,
    content: `评论内容 ${id}`,
    articleId: 9,
    authorName: `游客${id}`,
    createdAt: '2024-01-01T00:00:00.000Z'
  } as unknown as AdminComment
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  commentPageState.initialize(emptyCommentPageData())
  commentPageState.keyword = ''
  commentPageState.statusFilter = ''
  commentPageState.currentPage = 1
}

describe('评论页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果填充列表与总数', () => {
    const comments = [makeComment(1), makeComment(2)]

    commentPageState.initialize({ comments, total: 2 })

    expect(commentPageState.comments).toEqual(comments)
    expect(commentPageState.total).toBe(2)
  })

  it('空数据下无生效筛选条件且总页数按一页展示', () => {
    expect(commentPageState.hasActiveFilters).toBe(false)
    expect(commentPageState.totalPages).toBe(1)
  })

  // 空白关键词不构成筛选，避免只输入空格时误判为空态文案分支。
  it('关键词仅含空白字符时不视为生效筛选条件', () => {
    commentPageState.keyword = '   '

    expect(commentPageState.hasActiveFilters).toBe(false)
  })
})

describe('评论页状态：列表加载', () => {
  beforeEach(resetState)

  it('默认只携带分页参数', async () => {
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [makeComment(1)], total: 1 }
    })

    await commentPageState.loadComments()

    expect(commentApiMock.adminList).toHaveBeenCalledWith({ page: 1, pageSize: COMMENT_PAGE_SIZE })
    expect(commentPageState.comments).toHaveLength(1)
    expect(commentPageState.total).toBe(1)
  })

  it('筛选条件生效时按状态与去空白关键词查询', async () => {
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 0 }
    })
    commentPageState.statusFilter = 'trash'
    commentPageState.keyword = '  测试  '
    commentPageState.currentPage = 3

    await commentPageState.loadComments()

    expect(commentApiMock.adminList).toHaveBeenCalledWith({
      page: 3,
      pageSize: COMMENT_PAGE_SIZE,
      status: 'trash',
      keyword: '测试'
    })
  })

  it('业务码非成功时提示后端文案且不清空现有列表', async () => {
    commentApiMock.adminList.mockResolvedValue({ code: 500, message: '查询失败' })
    commentPageState.initialize({ comments: [makeComment(1)], total: 1 })

    await commentPageState.loadComments()

    expect(toastMock.error).toHaveBeenCalledWith('查询失败')
    expect(commentPageState.comments).toHaveLength(1)
  })

  it('请求抛错时提示网络错误并结束加载态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    commentApiMock.adminList.mockRejectedValue(new Error('断网'))

    await commentPageState.loadComments()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(commentPageState.isLoading).toBe(false)
    consoleError.mockRestore()
  })
})

describe('评论页状态：筛选与分页', () => {
  beforeEach(resetState)

  it('状态筛选变更后回到第一页并重新加载', async () => {
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 0 }
    })
    commentPageState.currentPage = 4

    commentPageState.setStatusFilter('spam')

    await vi.waitFor(() => expect(commentApiMock.adminList).toHaveBeenCalledTimes(1))
    expect(commentPageState.statusFilter).toBe('spam')
    expect(commentPageState.currentPage).toBe(1)
  })

  it('重置筛选清空全部条件并回到第一页', async () => {
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 0 }
    })
    commentPageState.keyword = '评论'
    commentPageState.statusFilter = 'approved'
    commentPageState.currentPage = 5

    commentPageState.resetFilters()

    expect(commentPageState.keyword).toBe('')
    expect(commentPageState.statusFilter).toBe('')
    expect(commentPageState.currentPage).toBe(1)
    expect(commentPageState.hasActiveFilters).toBe(false)
  })

  it('切换页码后按目标页加载并刷新总页数', async () => {
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 50 }
    })

    commentPageState.changePage(2)

    await vi.waitFor(() =>
      expect(commentApiMock.adminList).toHaveBeenCalledWith({
        page: 2,
        pageSize: COMMENT_PAGE_SIZE
      })
    )
    expect(commentPageState.totalPages).toBe(5)
  })
})

describe('评论页状态：搜索防抖', () => {
  beforeEach(() => {
    resetState()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  // 连续输入必须在停顿后只请求一次，且回到第一页，避免翻页后搜索命中空页。
  it('连续输入只在停顿后触发一次请求并回到第一页', () => {
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 0 }
    })
    commentPageState.currentPage = 3

    commentPageState.applyKeywordDebounced()
    commentPageState.applyKeywordDebounced()
    expect(commentApiMock.adminList).not.toHaveBeenCalled()

    vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS)

    expect(commentApiMock.adminList).toHaveBeenCalledTimes(1)
    expect(commentPageState.currentPage).toBe(1)
  })
})

describe('评论页状态：审核操作', () => {
  beforeEach(resetState)

  it('通过动作调用 approve 并提示成功、刷新列表', async () => {
    commentApiMock.approve.mockResolvedValue({ code: SUCCESS_CODE })
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 0 }
    })

    await commentPageState.runAction(makeComment(1), 'approve')

    expect(commentApiMock.approve).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('通过成功')
    await vi.waitFor(() => expect(commentApiMock.adminList).toHaveBeenCalledTimes(1))
  })

  it('拒绝动作调用 reject 并提示成功', async () => {
    commentApiMock.reject.mockResolvedValue({ code: SUCCESS_CODE })
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 0 }
    })

    await commentPageState.runAction(makeComment(1), 'reject')

    expect(commentApiMock.reject).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('拒绝成功')
  })

  it('标记垃圾动作走 markSpam 方法名', async () => {
    commentApiMock.markSpam.mockResolvedValue({ code: SUCCESS_CODE })
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 0 }
    })

    await commentPageState.runAction(makeComment(1), 'spam')

    expect(commentApiMock.markSpam).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('标记垃圾成功')
  })

  it('移入回收站动作调用 trash 并提示成功', async () => {
    commentApiMock.trash.mockResolvedValue({ code: SUCCESS_CODE })
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 0 }
    })

    await commentPageState.runAction(makeComment(1), 'trash')

    expect(commentApiMock.trash).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('移入回收站成功')
  })

  it('业务码非成功时提示后端文案且不刷新列表', async () => {
    commentApiMock.approve.mockResolvedValue({ code: 403, message: '无权操作该评论' })

    await commentPageState.runAction(makeComment(1), 'approve')

    expect(toastMock.error).toHaveBeenCalledWith('无权操作该评论')
    expect(toastMock.success).not.toHaveBeenCalled()
    expect(commentApiMock.adminList).not.toHaveBeenCalled()
  })

  it('请求抛错时提示网络错误且不刷新列表', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    commentApiMock.approve.mockRejectedValue(new Error('断网'))

    await commentPageState.runAction(makeComment(1), 'approve')

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(commentApiMock.adminList).not.toHaveBeenCalled()
    consoleError.mockRestore()
  })
})

describe('评论页状态：删除', () => {
  beforeEach(resetState)

  it('删除成功后提示并刷新列表', async () => {
    commentApiMock.delete.mockResolvedValue({ code: SUCCESS_CODE })
    commentApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { comments: [], total: 0 }
    })

    const removed = await commentPageState.deleteComment(1)

    expect(removed).toBe(true)
    expect(commentApiMock.delete).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('评论删除成功')
    await vi.waitFor(() => expect(commentApiMock.adminList).toHaveBeenCalledTimes(1))
  })

  it('删除失败时提示后端文案且不刷新列表', async () => {
    commentApiMock.delete.mockResolvedValue({ code: 404, message: '评论不存在' })

    const removed = await commentPageState.deleteComment(1)

    expect(removed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('评论不存在')
    expect(commentApiMock.adminList).not.toHaveBeenCalled()
  })

  it('删除请求抛错时提示网络错误并返回失败', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    commentApiMock.delete.mockRejectedValue(new Error('断网'))

    const removed = await commentPageState.deleteComment(1)

    expect(removed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    consoleError.mockRestore()
  })
})
