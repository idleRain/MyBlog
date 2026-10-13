import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { User } from '@myblog/api/modules/user/types'

import { USER_PAGE_SIZE } from '$lib/constants/user'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const userApiMock = {
  getUserList: vi.fn(),
  createUser: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
  batchDeleteUsers: vi.fn(),
  batchUpdateUserStatus: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn()
}

// 登录态以最小替身提供，可分配角色由权限工具按当前用户实时推导。
const authState: { user: User | null } = { user: null }

vi.mock('$lib/api', () => ({ UserAPI: userApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))
vi.mock('$lib/stores/auth', () => ({
  authStore: { getCurrentState: () => authState }
}))

const { emptyUserPageData, userPageState } = await import('./user-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 后端统一响应的失败业务码，用于校验失败分支的提示文案。
const ERROR_CODE = 500

// 当前登录用户的标识，与列表中的用户区分开。
const CURRENT_USER_ID = 100

// 搜索防抖时长，与页面状态模块内的取值一致。
const SEARCH_DEBOUNCE_MS = 300

// 构造一个用户替身，字段取值只覆盖被测逻辑关心的部分。
function makeUser(id: number, overrides: Partial<User> = {}): User {
  return {
    id,
    username: `user${id}`,
    email: `user${id}@blog.test`,
    nickname: `用户${id}`,
    avatar: '',
    birthday: null,
    role: 'user',
    status: 1,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
    ...overrides
  }
}

// 列表加载成功的统一响应替身。
const LIST_SUCCESS = { code: SUCCESS_CODE, data: { users: [], total: 0 } }

// 先等待上一用例触发的列表刷新结束，再清理替身与页面状态，避免用例之间互相污染。
async function resetState() {
  await new Promise(resolve => setTimeout(resolve, 0))
  vi.clearAllMocks()
  authState.user = makeUser(CURRENT_USER_ID, { role: 'superadmin' })
  userPageState.initialize(emptyUserPageData())
  userPageState.keyword = ''
  userPageState.currentPage = 1
  userPageState.selectedIds = []
  userPageState.isLoading = false
}

describe('用户页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果填充列表与总数', () => {
    const users = [makeUser(1), makeUser(2)]

    userPageState.initialize({ users, total: 2 })

    expect(userPageState.users).toEqual(users)
    expect(userPageState.total).toBe(2)
  })

  it('按当前登录用户推导可分配角色', () => {
    authState.user = makeUser(CURRENT_USER_ID, { role: 'superadmin' })
    userPageState.initialize(emptyUserPageData())
    expect(userPageState.assignableRoles).toEqual(['user', 'editor', 'admin', 'superadmin'])

    authState.user = makeUser(CURRENT_USER_ID, { role: 'admin' })
    userPageState.initialize(emptyUserPageData())
    expect(userPageState.assignableRoles).toEqual(['user', 'editor'])
  })

  it('空数据下派生状态取默认值', () => {
    userPageState.initialize(emptyUserPageData())

    expect(userPageState.totalPages).toBe(1)
    expect(userPageState.hasActiveFilters).toBe(false)
    expect(userPageState.isAllSelected).toBe(false)
  })
})

describe('用户页状态：列表加载', () => {
  beforeEach(resetState)

  it('默认只携带第一页与既定分页容量', async () => {
    userApiMock.getUserList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { users: [makeUser(1)], total: 1 }
    })

    await userPageState.loadUsers()

    expect(userApiMock.getUserList).toHaveBeenCalledWith(1, USER_PAGE_SIZE, '')
    expect(userPageState.users).toHaveLength(1)
    expect(userPageState.total).toBe(1)
  })

  it('搜索词去除首尾空白后按当前页码查询', async () => {
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)
    userPageState.keyword = '  博客  '
    userPageState.currentPage = 3

    await userPageState.loadUsers()

    expect(userApiMock.getUserList).toHaveBeenCalledWith(3, USER_PAGE_SIZE, '博客')
  })

  it('业务码非成功时提示后端文案且保留现有列表', async () => {
    userApiMock.getUserList.mockResolvedValue({ code: ERROR_CODE, message: '查询失败' })
    userPageState.initialize({ users: [makeUser(1)], total: 1 })

    await userPageState.loadUsers()

    expect(toastMock.error).toHaveBeenCalledWith('查询失败')
    expect(userPageState.users).toHaveLength(1)
    expect(userPageState.isLoading).toBe(false)
  })

  it('请求抛错时提示网络错误并结束加载态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    userApiMock.getUserList.mockRejectedValue(new Error('断网'))

    await userPageState.loadUsers()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(userPageState.isLoading).toBe(false)
    consoleError.mockRestore()
  })

  it('加载成功后清空批量选择', async () => {
    userApiMock.getUserList.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { users: [makeUser(1), makeUser(2)], total: 2 }
    })
    userPageState.selectedIds = [1, 2]

    await userPageState.loadUsers()

    expect(userPageState.selectedIds).toEqual([])
  })
})

describe('用户页状态：筛选与分页', () => {
  beforeEach(resetState)

  it('切换页码后按目标页加载，总页数按分页容量向上取整', () => {
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)

    userPageState.changePage(2)

    expect(userPageState.currentPage).toBe(2)
    expect(userApiMock.getUserList).toHaveBeenCalledWith(2, USER_PAGE_SIZE, '')

    userPageState.total = 25
    expect(userPageState.totalPages).toBe(3)
  })

  it('重置搜索清空关键词并回到第一页', () => {
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)
    userPageState.keyword = '博客'
    userPageState.currentPage = 5

    userPageState.resetAndReload()

    expect(userPageState.keyword).toBe('')
    expect(userPageState.currentPage).toBe(1)
    expect(userPageState.hasActiveFilters).toBe(false)
    expect(userApiMock.getUserList).toHaveBeenCalledWith(1, USER_PAGE_SIZE, '')
  })

  it('关键词仅含空白时不视为生效筛选', () => {
    userPageState.keyword = '   '

    expect(userPageState.hasActiveFilters).toBe(false)
  })
})

describe('用户页状态：搜索防抖', () => {
  beforeEach(async () => {
    await resetState()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  // 连续输入必须在停顿后只请求一次，且回到第一页，避免翻页后搜索命中空页。
  it('连续输入只在停顿后触发一次请求并回到第一页', () => {
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)
    userPageState.currentPage = 3

    userPageState.applySearchDebounced()
    userPageState.applySearchDebounced()
    expect(userApiMock.getUserList).not.toHaveBeenCalled()

    vi.advanceTimersByTime(SEARCH_DEBOUNCE_MS)

    expect(userApiMock.getUserList).toHaveBeenCalledTimes(1)
    expect(userPageState.currentPage).toBe(1)
  })
})

describe('用户页状态：批量选择', () => {
  beforeEach(resetState)

  it('全选与取消全选按当前页用户切换', () => {
    userPageState.users = [makeUser(1), makeUser(2)]

    userPageState.toggleSelectAll()
    expect(userPageState.selectedIds).toEqual([1, 2])
    expect(userPageState.isAllSelected).toBe(true)

    userPageState.toggleSelectAll()
    expect(userPageState.selectedIds).toEqual([])
    expect(userPageState.isAllSelected).toBe(false)
  })

  it('单项选择支持追加已选与移除已选', () => {
    userPageState.toggleSelect(1)
    userPageState.toggleSelect(2)
    expect(userPageState.selectedIds).toEqual([1, 2])

    userPageState.toggleSelect(1)
    expect(userPageState.selectedIds).toEqual([2])
  })

  it('清空选择移除全部标识', () => {
    userPageState.selectedIds = [1, 2]

    userPageState.clearSelection()

    expect(userPageState.selectedIds).toEqual([])
  })
})

describe('用户页状态：增删改', () => {
  beforeEach(resetState)

  it('创建用户成功时提示并刷新列表，失败时提示后端文案且不刷新', async () => {
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)
    userApiMock.createUser.mockResolvedValueOnce({ code: SUCCESS_CODE, data: makeUser(9) })

    const created = await userPageState.createUser({
      username: '新用户',
      email: 'new@blog.test',
      password: 'secret123'
    })

    expect(created).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('用户创建成功')
    expect(userApiMock.getUserList).toHaveBeenCalledTimes(1)

    userApiMock.getUserList.mockClear()
    userApiMock.createUser.mockResolvedValueOnce({ code: ERROR_CODE, message: '用户名已存在' })

    const duplicated = await userPageState.createUser({
      username: '重复',
      email: 'dup@blog.test',
      password: 'secret123'
    })

    expect(duplicated).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('用户名已存在')
    expect(userApiMock.getUserList).not.toHaveBeenCalled()
  })

  it('更新用户成功时提示并刷新列表，异常时提示网络错误', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)
    userApiMock.updateUser.mockResolvedValueOnce({ code: SUCCESS_CODE, data: makeUser(1) })

    const updated = await userPageState.updateUser({
      id: 1,
      username: 'user1',
      email: 'user1@blog.test'
    })

    expect(updated).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('用户更新成功')
    expect(userApiMock.getUserList).toHaveBeenCalledTimes(1)

    userApiMock.getUserList.mockClear()
    userApiMock.updateUser.mockRejectedValueOnce(new Error('断网'))

    const failed = await userPageState.updateUser({
      id: 1,
      username: 'user1',
      email: 'user1@blog.test'
    })

    expect(failed).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(userApiMock.getUserList).not.toHaveBeenCalled()
    consoleError.mockRestore()
  })

  it('删除用户成功时提示并刷新列表，失败时提示后端文案且不刷新', async () => {
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)
    userApiMock.deleteUser.mockResolvedValueOnce({ code: SUCCESS_CODE })

    const removed = await userPageState.removeUser(1)

    expect(removed).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('用户删除成功')
    expect(userApiMock.getUserList).toHaveBeenCalledTimes(1)

    userApiMock.getUserList.mockClear()
    userApiMock.deleteUser.mockResolvedValueOnce({ code: ERROR_CODE, message: '不能删除当前用户' })

    const refused = await userPageState.removeUser(2)

    expect(refused).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('不能删除当前用户')
    expect(userApiMock.getUserList).not.toHaveBeenCalled()
  })
})

describe('用户页状态：状态切换', () => {
  beforeEach(resetState)

  it('按目标状态提交完整用户资料，锁定态归入启用方向', async () => {
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)
    userApiMock.updateUser.mockResolvedValue({ code: SUCCESS_CODE, data: makeUser(1) })

    await userPageState.toggleUserStatus(makeUser(1, { status: 1 }))

    expect(userApiMock.updateUser).toHaveBeenCalledWith({
      id: 1,
      username: 'user1',
      email: 'user1@blog.test',
      nickname: '用户1',
      role: 'user',
      birthday: '',
      status: 0
    })
    expect(toastMock.success).toHaveBeenCalledWith('用户禁用成功')
    expect(userApiMock.getUserList).toHaveBeenCalledTimes(1)

    await userPageState.toggleUserStatus(makeUser(2, { status: 2 }))

    expect(userApiMock.updateUser).toHaveBeenLastCalledWith(
      expect.objectContaining({ id: 2, status: 1 })
    )
    expect(toastMock.success).toHaveBeenCalledWith('用户启用成功')
  })

  it('状态切换异常时提示网络错误且不刷新列表', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    userApiMock.updateUser.mockRejectedValue(new Error('断网'))

    await userPageState.toggleUserStatus(makeUser(1))

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(userApiMock.getUserList).not.toHaveBeenCalled()
    consoleError.mockRestore()
  })
})

describe('用户页状态：批量操作', () => {
  beforeEach(resetState)

  it('批量删除与批量状态切换均按选中标识提交', async () => {
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)
    userApiMock.batchDeleteUsers.mockResolvedValue({ code: SUCCESS_CODE })
    userApiMock.batchUpdateUserStatus.mockResolvedValue({ code: SUCCESS_CODE })

    userPageState.selectedIds = [1, 2]
    await expect(userPageState.executeBatch('delete')).resolves.toBe(true)
    expect(userApiMock.batchDeleteUsers).toHaveBeenCalledWith({ ids: [1, 2] })
    expect(toastMock.success).toHaveBeenCalledWith('批量删除成功')

    userPageState.selectedIds = [3]
    await expect(userPageState.executeBatch('enable')).resolves.toBe(true)
    expect(userApiMock.batchUpdateUserStatus).toHaveBeenCalledWith({ ids: [3], status: 1 })
    expect(toastMock.success).toHaveBeenCalledWith('批量启用成功')

    userPageState.selectedIds = [4]
    await expect(userPageState.executeBatch('disable')).resolves.toBe(true)
    expect(userApiMock.batchUpdateUserStatus).toHaveBeenCalledWith({ ids: [4], status: 0 })
    expect(toastMock.success).toHaveBeenCalledWith('批量禁用成功')
  })

  // 批量操作无论成败均在 finally 中刷新列表，避免选中项与后端状态脱节。
  it('批量操作失败与异常均提示错误并刷新当前页', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    userApiMock.getUserList.mockResolvedValue(LIST_SUCCESS)

    userPageState.selectedIds = [1]
    userApiMock.batchDeleteUsers.mockResolvedValue({
      code: ERROR_CODE,
      message: '存在不可删除的用户'
    })

    await expect(userPageState.executeBatch('delete')).resolves.toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('存在不可删除的用户')
    expect(userApiMock.getUserList).toHaveBeenCalledTimes(1)

    userPageState.selectedIds = [2]
    userApiMock.batchUpdateUserStatus.mockRejectedValue(new Error('断网'))

    await expect(userPageState.executeBatch('enable')).resolves.toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(userApiMock.getUserList).toHaveBeenCalledTimes(2)
    consoleError.mockRestore()
  })
})
