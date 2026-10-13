import type { FriendlyLink, LinkStatus } from '@myblog/api/modules/friendlyLink/types'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { LINK_PAGE_SIZE } from '$lib/constants/link'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const friendlyLinkApiMock = {
  adminList: vi.fn(),
  create: vi.fn(),
  update: vi.fn(),
  delete: vi.fn(),
  approve: vi.fn(),
  hide: vi.fn(),
  reject: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn()
}

vi.mock('$lib/api', () => ({ FriendlyLinkAPI: friendlyLinkApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { emptyLinkPageData, linkPageState } = await import('./link-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 构造一个友链替身，字段取值只覆盖被测逻辑关心的部分。
function makeLink(id: number, status: LinkStatus = 'pending', name = `友链${id}`): FriendlyLink {
  return {
    id,
    name,
    status,
    url: `https://link-${id}.example.com`
  } as unknown as FriendlyLink
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  linkPageState.initialize(emptyLinkPageData())
  linkPageState.statusFilter = ''
  linkPageState.currentPage = 1
}

describe('友链页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果填充列表与总数', () => {
    const links = [makeLink(1), makeLink(2, 'active')]

    linkPageState.initialize({ links, total: 2 })

    expect(linkPageState.links).toEqual(links)
    expect(linkPageState.total).toBe(2)
  })

  it('空数据下无生效筛选条件且总页数为 1', () => {
    expect(linkPageState.hasActiveFilters).toBe(false)
    expect(linkPageState.totalPages).toBe(1)
  })
})

describe('友链页状态：列表加载', () => {
  beforeEach(resetState)

  it('默认只携带分页参数', async () => {
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [makeLink(1)], total: 1 }
    })

    await linkPageState.loadLinks()

    expect(friendlyLinkApiMock.adminList).toHaveBeenCalledWith({
      page: 1,
      pageSize: LINK_PAGE_SIZE
    })
    expect(linkPageState.links).toHaveLength(1)
    expect(linkPageState.total).toBe(1)
  })

  it('筛选生效时携带状态与当前页码', async () => {
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 0 }
    })
    linkPageState.statusFilter = 'hidden'
    linkPageState.currentPage = 2

    await linkPageState.loadLinks()

    expect(friendlyLinkApiMock.adminList).toHaveBeenCalledWith({
      page: 2,
      pageSize: LINK_PAGE_SIZE,
      status: 'hidden'
    })
  })

  it('业务码非成功时提示后端文案且不清空现有列表', async () => {
    friendlyLinkApiMock.adminList.mockResolvedValue({ code: 500, message: '查询失败' })
    linkPageState.initialize({ links: [makeLink(1)], total: 1 })

    await linkPageState.loadLinks()

    expect(toastMock.error).toHaveBeenCalledWith('查询失败')
    expect(linkPageState.links).toHaveLength(1)
  })

  it('请求抛错时提示网络错误并复位加载态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    friendlyLinkApiMock.adminList.mockRejectedValue(new Error('断网'))

    await linkPageState.loadLinks()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(linkPageState.isLoading).toBe(false)
    consoleError.mockRestore()
  })
})

describe('友链页状态：筛选与分页', () => {
  beforeEach(resetState)

  it('状态筛选变更后回到第一页并重新加载', async () => {
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 0 }
    })
    linkPageState.currentPage = 4

    linkPageState.setStatusFilter('active')

    await vi.waitFor(() =>
      expect(friendlyLinkApiMock.adminList).toHaveBeenCalledWith({
        page: 1,
        pageSize: LINK_PAGE_SIZE,
        status: 'active'
      })
    )
    expect(linkPageState.currentPage).toBe(1)
    expect(linkPageState.hasActiveFilters).toBe(true)
  })

  it('重置筛选清空条件并回到第一页', async () => {
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 0 }
    })
    linkPageState.statusFilter = 'rejected'
    linkPageState.currentPage = 5

    linkPageState.resetAndReload()

    expect(linkPageState.statusFilter).toBe('')
    expect(linkPageState.currentPage).toBe(1)
    expect(linkPageState.hasActiveFilters).toBe(false)
    await vi.waitFor(() =>
      expect(friendlyLinkApiMock.adminList).toHaveBeenCalledWith({
        page: 1,
        pageSize: LINK_PAGE_SIZE
      })
    )
  })

  it('切换页码后按目标页加载且总页数向上取整', async () => {
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 21 }
    })

    linkPageState.changePage(3)

    await vi.waitFor(() =>
      expect(friendlyLinkApiMock.adminList).toHaveBeenCalledWith({
        page: 3,
        pageSize: LINK_PAGE_SIZE
      })
    )
    expect(linkPageState.totalPages).toBe(3)
  })
})

describe('友链页状态：写操作', () => {
  beforeEach(resetState)

  it('创建成功后提示并刷新列表', async () => {
    friendlyLinkApiMock.create.mockResolvedValue({ code: SUCCESS_CODE, data: makeLink(9) })
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 0 }
    })

    const created = await linkPageState.createLink({ name: '新友链' } as never)

    expect(created).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('友链添加成功')
    expect(friendlyLinkApiMock.adminList).toHaveBeenCalledTimes(1)
  })

  it('创建失败时提示后端文案且不刷新列表', async () => {
    friendlyLinkApiMock.create.mockResolvedValue({ code: 400, message: '站点链接已存在' })

    const created = await linkPageState.createLink({ name: '重复' } as never)

    expect(created).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('站点链接已存在')
    expect(friendlyLinkApiMock.adminList).not.toHaveBeenCalled()
  })

  it('更新成功后提示并刷新列表', async () => {
    friendlyLinkApiMock.update.mockResolvedValue({ code: SUCCESS_CODE, data: makeLink(1) })
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 0 }
    })

    const updated = await linkPageState.updateLink({ id: 1, name: '改名' } as never)

    expect(updated).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('友链更新成功')
    expect(friendlyLinkApiMock.adminList).toHaveBeenCalledTimes(1)
  })

  it('更新抛错时提示网络错误且不刷新列表', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    friendlyLinkApiMock.update.mockRejectedValue(new Error('断网'))

    const updated = await linkPageState.updateLink({ id: 1, name: '改名' } as never)

    expect(updated).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(friendlyLinkApiMock.adminList).not.toHaveBeenCalled()
    consoleError.mockRestore()
  })

  it('删除只要求业务码，成功时提示并刷新列表', async () => {
    friendlyLinkApiMock.delete.mockResolvedValue({ code: SUCCESS_CODE })
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 0 }
    })

    const removed = await linkPageState.removeLink(1)

    expect(removed).toBe(true)
    expect(friendlyLinkApiMock.delete).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('友链删除成功')
    expect(friendlyLinkApiMock.adminList).toHaveBeenCalledTimes(1)
  })

  it('删除失败时提示后端文案且不刷新列表', async () => {
    friendlyLinkApiMock.delete.mockResolvedValue({ code: 400, message: '删除友链失败' })

    const removed = await linkPageState.removeLink(1)

    expect(removed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('删除友链失败')
    expect(friendlyLinkApiMock.adminList).not.toHaveBeenCalled()
  })
})

describe('友链页状态：审核操作', () => {
  beforeEach(resetState)

  it('待审核友链通过后提示通过并刷新列表', async () => {
    friendlyLinkApiMock.approve.mockResolvedValue({ code: SUCCESS_CODE })
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 0 }
    })

    const reviewed = await linkPageState.reviewLink(makeLink(1, 'pending'), 'approve')

    expect(reviewed).toBe(true)
    expect(friendlyLinkApiMock.approve).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('通过成功')
    expect(friendlyLinkApiMock.adminList).toHaveBeenCalledTimes(1)
  })

  it('已隐藏友链重新上架后提示重新上架', async () => {
    friendlyLinkApiMock.approve.mockResolvedValue({ code: SUCCESS_CODE })
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 0 }
    })

    await linkPageState.reviewLink(makeLink(1, 'hidden'), 'approve')

    expect(toastMock.success).toHaveBeenCalledWith('重新上架成功')
  })

  it('展示中友链下架后提示下架', async () => {
    friendlyLinkApiMock.hide.mockResolvedValue({ code: SUCCESS_CODE })
    friendlyLinkApiMock.adminList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { links: [], total: 0 }
    })

    const reviewed = await linkPageState.reviewLink(makeLink(1, 'active'), 'hide')

    expect(reviewed).toBe(true)
    expect(friendlyLinkApiMock.hide).toHaveBeenCalledWith(1)
    expect(toastMock.success).toHaveBeenCalledWith('下架成功')
  })

  it('审核业务码非成功时提示后端文案且不刷新列表', async () => {
    friendlyLinkApiMock.reject.mockResolvedValue({ code: 400, message: '状态不允许该操作' })

    const reviewed = await linkPageState.reviewLink(makeLink(1, 'pending'), 'reject')

    expect(reviewed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('状态不允许该操作')
    expect(friendlyLinkApiMock.adminList).not.toHaveBeenCalled()
  })

  it('审核请求抛错时提示网络错误', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    friendlyLinkApiMock.approve.mockRejectedValue(new Error('断网'))

    const reviewed = await linkPageState.reviewLink(makeLink(1, 'pending'), 'approve')

    expect(reviewed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    consoleError.mockRestore()
  })
})
