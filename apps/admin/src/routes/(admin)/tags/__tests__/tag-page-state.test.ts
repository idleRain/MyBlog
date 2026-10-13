import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { EnabledDictGroup } from '@myblog/api/modules/dict/types'
import type { Tag } from '@myblog/api/modules/tag/types'

import { TAG_PAGE_SIZE } from '$lib/constants/tag'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const tagApiMock = {
  adminList: vi.fn(),
  create: vi.fn(),
  update: vi.fn(),
  delete: vi.fn()
}

const dictApiMock = {
  getByType: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn()
}

vi.mock('$lib/api', () => ({ TagAPI: tagApiMock, DictAPI: dictApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { emptyTagPageData, tagPageState } = await import('../tag-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 搜索防抖时长，与页面状态模块内的取值一致。
const SEARCH_DEBOUNCE_MS = 300

// 构造一个标签替身，字段取值只覆盖被测逻辑关心的部分。
function makeTag(id: number, name = `标签${id}`): Tag {
  return { id, name, slug: `tag-${id}`, status: 1, usageCount: 0, isHot: false } as unknown as Tag
}

// 构造标签状态字典替身。
function makeStatusDict(): EnabledDictGroup {
  return {
    code: 'tag_status',
    items: [{ value: '1', label: '启用', sortOrder: 1 }]
  } as unknown as EnabledDictGroup
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  tagPageState.initialize(emptyTagPageData())
  tagPageState.search = ''
  tagPageState.statusFilter = ''
  tagPageState.hotFilter = ''
  tagPageState.currentPage = 1
}

describe('标签页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果填充列表、总数与状态字典', () => {
    const tags = [makeTag(1), makeTag(2)]
    const statusDict = makeStatusDict()

    tagPageState.initialize({ tags, total: 2, statusDict })

    expect(tagPageState.tags).toEqual(tags)
    expect(tagPageState.total).toBe(2)
    expect(tagPageState.statusDict).toEqual(statusDict)
  })

  it('空数据下无生效筛选条件', () => {
    expect(tagPageState.hasActiveFilters).toBe(false)
    expect(tagPageState.totalPages).toBe(1)
  })
})

describe('标签页状态：列表加载', () => {
  beforeEach(resetState)

  it('默认只携带分页参数', async () => {
    tagApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { tags: [makeTag(1)], total: 1 }
    })

    await tagPageState.loadTags()

    expect(tagApiMock.adminList).toHaveBeenCalledWith({ page: 1, pageSize: TAG_PAGE_SIZE })
    expect(tagPageState.tags).toHaveLength(1)
    expect(tagPageState.total).toBe(1)
  })

  it('筛选条件生效时按状态、热门与搜索词查询', async () => {
    tagApiMock.adminList.mockResolvedValue({ code: SUCCESS_CODE, data: { tags: [], total: 0 } })
    tagPageState.statusFilter = '0'
    tagPageState.hotFilter = '1'
    tagPageState.search = '  博客  '
    tagPageState.currentPage = 3

    await tagPageState.loadTags()

    expect(tagApiMock.adminList).toHaveBeenCalledWith({
      page: 3,
      pageSize: TAG_PAGE_SIZE,
      status: 0,
      isHot: true,
      search: '博客'
    })
  })

  it('业务码非成功时提示后端文案且不清空现有列表', async () => {
    tagApiMock.adminList.mockResolvedValue({ code: 500, message: '查询失败' })
    tagPageState.initialize({ tags: [makeTag(1)], total: 1, statusDict: null })

    await tagPageState.loadTags()

    expect(toastMock.error).toHaveBeenCalledWith('查询失败')
    expect(tagPageState.tags).toHaveLength(1)
  })

  it('请求抛错时提示网络错误', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    tagApiMock.adminList.mockRejectedValue(new Error('断网'))

    await tagPageState.loadTags()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(tagPageState.isLoading).toBe(false)
    consoleError.mockRestore()
  })
})

describe('标签页状态：筛选项与分页', () => {
  beforeEach(resetState)

  it('字典缺失时状态筛选项回退内置配置', () => {
    expect(tagPageState.statusFilterOptions).toEqual([
      { value: '1', label: '启用' },
      { value: '0', label: '隐藏' }
    ])
  })

  it('字典可用时按排序值生成状态筛选项', () => {
    tagPageState.statusDict = {
      items: [
        { value: '0', label: '隐藏', sortOrder: 2 },
        { value: '1', label: '启用', sortOrder: 1 }
      ]
    } as unknown as EnabledDictGroup

    expect(tagPageState.statusFilterOptions).toEqual([
      { value: '1', label: '启用' },
      { value: '0', label: '隐藏' }
    ])
  })

  it('状态与热门筛选变更后回到第一页并重新加载', async () => {
    tagApiMock.adminList.mockResolvedValue({ code: SUCCESS_CODE, data: { tags: [], total: 0 } })
    tagPageState.currentPage = 4

    tagPageState.setStatusFilter('0')
    await vi.waitFor(() => expect(tagApiMock.adminList).toHaveBeenCalledTimes(1))
    expect(tagPageState.currentPage).toBe(1)

    tagPageState.setHotFilter('1')
    await vi.waitFor(() => expect(tagApiMock.adminList).toHaveBeenCalledTimes(2))
    expect(tagPageState.hotFilter).toBe('1')
  })

  it('重置筛选清空全部条件并回到第一页', async () => {
    tagApiMock.adminList.mockResolvedValue({ code: SUCCESS_CODE, data: { tags: [], total: 0 } })
    tagPageState.search = '博客'
    tagPageState.statusFilter = '1'
    tagPageState.hotFilter = '1'
    tagPageState.currentPage = 5

    tagPageState.resetFilters()

    expect(tagPageState.search).toBe('')
    expect(tagPageState.statusFilter).toBe('')
    expect(tagPageState.hotFilter).toBe('')
    expect(tagPageState.currentPage).toBe(1)
    expect(tagPageState.hasActiveFilters).toBe(false)
  })

  it('切换页码后按目标页加载', async () => {
    tagApiMock.adminList.mockResolvedValue({ code: SUCCESS_CODE, data: { tags: [], total: 50 } })

    tagPageState.changePage(2)

    await vi.waitFor(() =>
      expect(tagApiMock.adminList).toHaveBeenCalledWith({ page: 2, pageSize: TAG_PAGE_SIZE })
    )
    expect(tagPageState.totalPages).toBe(5)
  })
})

describe('标签页状态：搜索防抖', () => {
  beforeEach(() => {
    resetState()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  // 连续输入必须在停顿后只请求一次，且回到第一页，避免翻页后搜索命中空页。
  it('连续输入只在停顿后触发一次请求并回到第一页', () => {
    tagApiMock.adminList.mockResolvedValue({ code: SUCCESS_CODE, data: { tags: [], total: 0 } })
    tagPageState.currentPage = 3

    tagPageState.applySearchDebounced()
    tagPageState.applySearchDebounced()
    expect(tagApiMock.adminList).not.toHaveBeenCalled()

    vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS)

    expect(tagApiMock.adminList).toHaveBeenCalledTimes(1)
    expect(tagPageState.currentPage).toBe(1)
  })
})

describe('标签页状态：写操作', () => {
  beforeEach(resetState)

  it('创建成功后提示并刷新列表', async () => {
    tagApiMock.create.mockResolvedValue({ code: SUCCESS_CODE, data: makeTag(9) })
    tagApiMock.adminList.mockResolvedValue({ code: SUCCESS_CODE, data: { tags: [], total: 0 } })

    const created = await tagPageState.createTag({ name: '新标签' } as never)

    expect(created).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('标签创建成功')
    expect(tagApiMock.adminList).toHaveBeenCalledTimes(1)
  })

  it('创建失败时提示后端文案且不刷新列表', async () => {
    tagApiMock.create.mockResolvedValue({ code: 400, message: '标签名称已存在' })

    const created = await tagPageState.createTag({ name: '重复' } as never)

    expect(created).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('标签名称已存在')
    expect(tagApiMock.adminList).not.toHaveBeenCalled()
  })

  it('更新成功后提示并刷新列表', async () => {
    tagApiMock.update.mockResolvedValue({ code: SUCCESS_CODE, data: makeTag(1) })
    tagApiMock.adminList.mockResolvedValue({ code: SUCCESS_CODE, data: { tags: [], total: 0 } })

    const updated = await tagPageState.updateTag({ id: 1, name: '改名' } as never)

    expect(updated).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('标签更新成功')
  })

  it('删除成功后提示并刷新列表', async () => {
    tagApiMock.delete.mockResolvedValue({ code: SUCCESS_CODE })
    tagApiMock.adminList.mockResolvedValue({ code: SUCCESS_CODE, data: { tags: [], total: 0 } })

    const removed = await tagPageState.removeTag(1)

    expect(removed).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('标签删除成功')
    expect(tagApiMock.adminList).toHaveBeenCalledTimes(1)
  })

  it('删除失败时不刷新列表', async () => {
    tagApiMock.delete.mockResolvedValue({ code: 400, message: '标签仍被引用' })

    const removed = await tagPageState.removeTag(1)

    expect(removed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('标签仍被引用')
    expect(tagApiMock.adminList).not.toHaveBeenCalled()
  })
})

describe('标签页状态：状态字典', () => {
  beforeEach(resetState)

  it('加载成功后覆盖状态字典', async () => {
    const statusDict = makeStatusDict()
    dictApiMock.getByType.mockResolvedValue({ code: SUCCESS_CODE, data: statusDict })

    await tagPageState.loadStatusDict()

    expect(tagPageState.statusDict).toEqual(statusDict)
  })

  // 字典是展示增强而非列表前提，加载失败必须静默降级。
  it('加载失败时保持原有字典且不抛错', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    dictApiMock.getByType.mockRejectedValue(new Error('断网'))

    await expect(tagPageState.loadStatusDict()).resolves.toBeUndefined()
    expect(tagPageState.statusDict).toBeNull()
    expect(toastMock.error).not.toHaveBeenCalled()
    consoleError.mockRestore()
  })
})
