import type { CategoryTreeNode } from '@myblog/api/modules/category/types'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { CATEGORY_STATUS_FILTER_ALL } from '$lib/constants/category'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const categoryApiMock = {
  getTree: vi.fn(),
  create: vi.fn(),
  update: vi.fn(),
  delete: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn()
}

vi.mock('$lib/api', () => ({ CategoryAPI: categoryApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { emptyCategoryPageData, categoryPageState } = await import('../category-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 网络异常的统一提示文案，与页面状态模块内的取值一致。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 构造分类树节点替身，字段取值只覆盖被测逻辑关心的部分。
function makeNode(
  id: number,
  options: { name?: string; slug?: string; status?: number; children?: CategoryTreeNode[] } = {}
): CategoryTreeNode {
  const { name = `分类${id}`, slug = `category-${id}`, status = 1, children = [] } = options
  return { id, name, slug, status, children } as unknown as CategoryTreeNode
}

// 两级分类树替身：顶级分类 1 带子分类 11，顶级分类 2 为隐藏分类。
function makeTree(): CategoryTreeNode[] {
  return [
    makeNode(1, {
      name: '技术',
      slug: 'tech',
      children: [makeNode(11, { name: '前端', slug: 'fe' })]
    }),
    makeNode(2, { name: '生活', slug: 'life', status: 0 })
  ]
}

// 当前筛选结果中的分类 ID 序列，用于断言筛选后的行集合与顺序。
function filteredIds(): number[] {
  return categoryPageState.filteredRows.map(row => row.category.id)
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  categoryPageState.initialize(emptyCategoryPageData())
  categoryPageState.searchQuery = ''
  categoryPageState.statusFilter = CATEGORY_STATUS_FILTER_ALL
  categoryPageState.isLoading = false
}

describe('分类页状态：初始化', () => {
  beforeEach(resetState)

  it('按树层级扁平化 load 结果并记录缩进深度', () => {
    categoryPageState.initialize({ tree: makeTree() })

    expect(categoryPageState.rows.map(row => [row.category.id, row.depth])).toEqual([
      [1, 0],
      [11, 1],
      [2, 0]
    ])
    expect(categoryPageState.flatCategories.map(category => category.id)).toEqual([1, 11, 2])
    expect(categoryPageState.hasCategories).toBe(true)
    expect(categoryPageState.total).toBe(3)
  })

  it('空数据下无分类且筛选结果为空', () => {
    categoryPageState.initialize(emptyCategoryPageData())

    expect(categoryPageState.hasCategories).toBe(false)
    expect(categoryPageState.filteredRows).toEqual([])
  })

  it('重复初始化时以最新的树覆盖既有行', () => {
    categoryPageState.initialize({ tree: makeTree() })
    categoryPageState.initialize({ tree: [makeNode(5, { name: '随笔' })] })

    expect(filteredIds()).toEqual([5])
  })
})

describe('分类页状态：树加载', () => {
  beforeEach(resetState)

  it('加载成功后替换扁平行且不携带参数', async () => {
    categoryApiMock.getTree.mockResolvedValue({ code: SUCCESS_CODE, data: { tree: makeTree() } })

    await categoryPageState.loadCategories()

    expect(categoryApiMock.getTree).toHaveBeenCalledWith()
    expect(filteredIds()).toEqual([1, 11, 2])
    expect(categoryPageState.isLoading).toBe(false)
  })

  it('加载失败时提示后端文案或默认文案且保留现有行', async () => {
    categoryPageState.initialize({ tree: makeTree() })
    categoryApiMock.getTree.mockResolvedValue({ code: 500, message: '查询失败' })

    await categoryPageState.loadCategories()
    expect(toastMock.error).toHaveBeenCalledWith('查询失败')

    // 响应缺少数据且未携带文案时回退到模块内的默认失败文案。
    categoryApiMock.getTree.mockResolvedValue({ code: SUCCESS_CODE })
    await categoryPageState.loadCategories()
    expect(toastMock.error).toHaveBeenLastCalledWith('加载分类失败')
    expect(categoryPageState.rows).toHaveLength(3)
  })

  it('请求抛错时提示网络错误并复位加载态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    categoryApiMock.getTree.mockRejectedValue(new Error('断网'))

    await categoryPageState.loadCategories()

    expect(toastMock.error).toHaveBeenCalledWith(NETWORK_ERROR_MESSAGE)
    expect(categoryPageState.isLoading).toBe(false)
    expect(consoleError).toHaveBeenCalled()
    consoleError.mockRestore()
  })

  it('请求未完成期间处于加载态', async () => {
    let resolveTree: (value: unknown) => void = () => {}
    categoryApiMock.getTree.mockReturnValue(
      new Promise(resolve => {
        resolveTree = resolve
      })
    )

    const loading = categoryPageState.loadCategories()
    expect(categoryPageState.isLoading).toBe(true)

    resolveTree({ code: SUCCESS_CODE, data: { tree: [] } })
    await loading

    expect(categoryPageState.isLoading).toBe(false)
  })
})

describe('分类页状态：客户端筛选', () => {
  beforeEach(() => {
    resetState()
    categoryPageState.initialize({ tree: makeTree() })
  })

  it('按名称匹配且忽略大小写', () => {
    categoryPageState.searchQuery = '技术'

    expect(filteredIds()).toEqual([1])
  })

  it('按 URL 标识匹配', () => {
    categoryPageState.searchQuery = 'LIFE'

    expect(filteredIds()).toEqual([2])
  })

  it('按显示状态匹配', () => {
    categoryPageState.statusFilter = '0'

    expect(filteredIds()).toEqual([2])
  })

  it('名称或标识命中后再与状态取交集', () => {
    categoryPageState.searchQuery = '前端'
    categoryPageState.statusFilter = '1'
    expect(filteredIds()).toEqual([11])

    categoryPageState.statusFilter = '0'
    expect(filteredIds()).toEqual([])

    // 名称命中但状态不符的行不展示。
    categoryPageState.searchQuery = '生活'
    categoryPageState.statusFilter = '1'
    expect(filteredIds()).toEqual([])
  })

  it('重置筛选清空条件并恢复全量行', () => {
    categoryPageState.searchQuery = '技术'
    categoryPageState.statusFilter = '0'

    categoryPageState.resetFilters()

    expect(categoryPageState.searchQuery).toBe('')
    expect(categoryPageState.statusFilter).toBe(CATEGORY_STATUS_FILTER_ALL)
    expect(filteredIds()).toEqual([1, 11, 2])
  })

  it('筛选无命中时仍保留全量行，供空态区分无数据与无匹配', () => {
    categoryPageState.searchQuery = '不存在的分类'

    expect(filteredIds()).toEqual([])
    expect(categoryPageState.hasCategories).toBe(true)
  })
})

describe('分类页状态：写操作', () => {
  beforeEach(() => {
    resetState()
    categoryPageState.initialize({ tree: makeTree() })
    categoryApiMock.getTree.mockResolvedValue({ code: SUCCESS_CODE, data: { tree: makeTree() } })
  })

  it('创建成功后提示并重新加载分类树', async () => {
    categoryApiMock.create.mockResolvedValue({ code: SUCCESS_CODE, data: makeNode(9) })

    const created = await categoryPageState.createCategory({ name: '新分类' })

    expect(created).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('分类创建成功')
    expect(categoryApiMock.getTree).toHaveBeenCalledTimes(1)
  })

  it('创建失败时提示后端文案且不刷新分类树', async () => {
    categoryApiMock.create.mockResolvedValue({ code: 400, message: '分类名称已存在' })

    const created = await categoryPageState.createCategory({ name: '重复' })

    expect(created).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('分类名称已存在')
    expect(categoryApiMock.getTree).not.toHaveBeenCalled()
  })

  it('更新成功后提示并重新加载分类树', async () => {
    categoryApiMock.update.mockResolvedValue({ code: SUCCESS_CODE, data: makeNode(1) })

    const updated = await categoryPageState.updateCategory({ id: 1, name: '改名' })

    expect(updated).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('分类更新成功')
    expect(categoryApiMock.getTree).toHaveBeenCalledTimes(1)
  })

  it('更新失败时提示后端文案且不刷新分类树', async () => {
    categoryApiMock.update.mockResolvedValue({ code: 400, message: '分类不存在' })

    const updated = await categoryPageState.updateCategory({ id: 1, name: '改名' })

    expect(updated).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('分类不存在')
    expect(categoryApiMock.getTree).not.toHaveBeenCalled()
  })

  it('删除成功后提示并重新加载分类树', async () => {
    categoryApiMock.delete.mockResolvedValue({ code: SUCCESS_CODE })

    const removed = await categoryPageState.removeCategory(2)

    expect(removed).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('分类删除成功')
    expect(categoryApiMock.getTree).toHaveBeenCalledTimes(1)
  })

  // 存在子分类的删除由后端拒绝，前端必须原样转达后端文案且不刷新列表。
  it('删除被后端拒绝时提示后端文案且不刷新分类树', async () => {
    categoryApiMock.delete.mockResolvedValue({ code: 400, message: '该分类下存在子分类，无法删除' })

    const removed = await categoryPageState.removeCategory(1)

    expect(removed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('该分类下存在子分类，无法删除')
    expect(categoryApiMock.getTree).not.toHaveBeenCalled()
  })

  it('写操作请求抛错时提示网络错误且不刷新分类树', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    categoryApiMock.create.mockRejectedValue(new Error('断网'))

    const created = await categoryPageState.createCategory({ name: '新分类' })

    expect(created).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith(NETWORK_ERROR_MESSAGE)
    expect(categoryApiMock.getTree).not.toHaveBeenCalled()
    consoleError.mockRestore()
  })
})
