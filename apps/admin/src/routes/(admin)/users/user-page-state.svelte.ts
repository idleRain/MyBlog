import type {
  CreateUserRequest,
  UpdateUserRequest,
  User,
  UserRole
} from '@myblog/api/modules/user/types'
import { USER_PAGE_SIZE, USER_STATUS_DISABLED, USER_STATUS_ENABLED } from '$lib/constants/user'
import { getAssignableRoles } from '$lib/utils/permissions'
import { authStore } from '$lib/stores/auth'
import { debounce } from '@myblog/shared'
import { toast } from 'svelte-sonner'
import { UserAPI } from '$lib/api'

// 搜索输入停顿时长，避免每次按键都触发请求。
const SEARCH_DEBOUNCE_MS = 300

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 登录态尚未就绪时的角色选项降级值，保留完整角色集合。
const DEFAULT_ASSIGNABLE_ROLES: UserRole[] = ['user', 'editor', 'admin', 'superadmin']

// 启用与禁用两态的动作文案，单条状态切换的提示语与批量操作共用。
const STATUS_ACTION_LABELS: Record<
  typeof USER_STATUS_DISABLED | typeof USER_STATUS_ENABLED,
  string
> = {
  [USER_STATUS_DISABLED]: '禁用',
  [USER_STATUS_ENABLED]: '启用'
}

// 批量操作动作，与二次确认弹窗的语义一一对应。
export type BatchAction = 'enable' | 'disable' | 'delete'

// 批量操作的动作文案，用于成功与失败提示语的拼装。
const BATCH_ACTION_LABELS: Record<BatchAction, string> = {
  enable: STATUS_ACTION_LABELS[USER_STATUS_ENABLED],
  disable: STATUS_ACTION_LABELS[USER_STATUS_DISABLED],
  delete: '删除'
}

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface UserPageData {
  users: User[]
  total: number
}

// 空页面数据，load 失败时使用。
export function emptyUserPageData(): UserPageData {
  return { users: [], total: 0 }
}

// 用户管理页状态：集中持有列表、筛选、批量选择与全部读写操作，页面组件只做渲染。
class UserPageState {
  users = $state<User[]>([])
  total = $state(0)
  currentPage = $state(1)
  keyword = $state('')
  // 批量选择的目标用户标识，列表重新加载成功后清空。
  selectedIds = $state<number[]>([])
  isLoading = $state(false)
  // 当前用户可分配角色，由登录态经权限工具推导，作为表单角色选项的数据源。
  assignableRoles = $state<UserRole[]>(DEFAULT_ASSIGNABLE_ROLES)

  // 总页数，分页控件与页码文案共用同一口径。
  get totalPages(): number {
    return Math.max(1, Math.ceil(this.total / USER_PAGE_SIZE))
  }

  // 是否存在生效中的搜索条件，用于空态区分无数据与无匹配。
  get hasActiveFilters(): boolean {
    return this.keyword.trim() !== ''
  }

  // 当前页是否已全选，页内无数据时恒为未全选。
  get isAllSelected(): boolean {
    return this.users.length > 0 && this.selectedIds.length === this.users.length
  }

  // 用 load 结果初始化页面状态，可分配角色按当前登录态重新推导。
  initialize(data: UserPageData) {
    this.users = data.users
    this.total = data.total
    this.assignableRoles = getAssignableRoles(authStore.getCurrentState().user)
  }

  // 加载用户列表，搜索词经 keyword 参数交由后端模糊匹配。
  // 写操作成功后经此方法刷新且不等待刷新结果，保证弹窗与确认框即时关闭。
  async loadUsers() {
    this.isLoading = true
    try {
      const response = await UserAPI.getUserList(
        this.currentPage,
        USER_PAGE_SIZE,
        this.keyword.trim()
      )
      if (response.code === 200 && response.data) {
        this.users = response.data.users ?? []
        this.total = response.data.total ?? 0
        this.selectedIds = []
      } else {
        toast.error(response.message || '加载用户列表失败')
      }
    } catch (error) {
      console.error('加载用户列表失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 搜索词防抖应用，连续输入只在停顿后请求一次，并回到第一页。
  applySearchDebounced = debounce(() => {
    this.currentPage = 1
    void this.loadUsers()
  }, SEARCH_DEBOUNCE_MS)

  // 清空搜索并回到第一页。
  resetAndReload() {
    this.keyword = ''
    this.currentPage = 1
    void this.loadUsers()
  }

  // 切换页码并重新加载。
  changePage(page: number) {
    this.currentPage = page
    void this.loadUsers()
  }

  // 全选当前页；已全选时取消全部选择。
  toggleSelectAll() {
    this.selectedIds =
      this.selectedIds.length === this.users.length ? [] : this.users.map(user => user.id)
  }

  // 切换单个用户的选择状态。
  toggleSelect(id: number) {
    this.selectedIds = this.selectedIds.includes(id)
      ? this.selectedIds.filter(item => item !== id)
      : [...this.selectedIds, id]
  }

  // 清空批量选择。
  clearSelection() {
    this.selectedIds = []
  }

  // 创建用户，成功后刷新列表。
  async createUser(payload: CreateUserRequest): Promise<boolean> {
    try {
      const response = await UserAPI.createUser(payload)
      if (response.code === 200 && response.data) {
        toast.success('用户创建成功')
        void this.loadUsers()
        return true
      }
      toast.error(response.message || '保存用户失败')
    } catch (error) {
      console.error('保存用户失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 更新用户，成功后刷新列表。
  async updateUser(payload: UpdateUserRequest): Promise<boolean> {
    try {
      const response = await UserAPI.updateUser(payload)
      if (response.code === 200 && response.data) {
        toast.success('用户更新成功')
        void this.loadUsers()
        return true
      }
      toast.error(response.message || '保存用户失败')
    } catch (error) {
      console.error('保存用户失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 切换用户启用与禁用状态，状态 2 为锁定态，按原口径归入启用方向。
  async toggleUserStatus(user: User): Promise<void> {
    const nextStatus =
      user.status === USER_STATUS_ENABLED ? USER_STATUS_DISABLED : USER_STATUS_ENABLED
    const action = STATUS_ACTION_LABELS[nextStatus]
    try {
      const response = await UserAPI.updateUser({
        id: user.id,
        username: user.username,
        email: user.email,
        nickname: user.nickname || '',
        role: user.role || 'user',
        birthday: user.birthday || '',
        status: nextStatus
      })
      if (response.code === 200) {
        toast.success(`用户${action}成功`)
        void this.loadUsers()
      } else {
        toast.error(response.message || `${action}用户失败`)
      }
    } catch (error) {
      console.error(`切换用户状态失败:`, error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
  }

  // 删除单个用户，成功后刷新列表。
  async removeUser(id: number): Promise<boolean> {
    try {
      const response = await UserAPI.deleteUser(id)
      if (response.code === 200) {
        toast.success('用户删除成功')
        void this.loadUsers()
        return true
      }
      toast.error(response.message || '删除用户失败')
    } catch (error) {
      console.error('删除用户失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 执行确认后的批量操作，单次请求交由后端原子处理。
  // 后端在事务内完成，任一目标违规整批拒绝，前端仅接收整体成败。
  // 列表刷新沿用原实现的 finally 语义，成功与失败均重新加载当前页。
  async executeBatch(action: BatchAction): Promise<boolean> {
    const ids = [...this.selectedIds]
    try {
      const response =
        action === 'delete'
          ? await UserAPI.batchDeleteUsers({ ids })
          : await UserAPI.batchUpdateUserStatus({
              ids,
              status: action === 'enable' ? USER_STATUS_ENABLED : USER_STATUS_DISABLED
            })
      if (response.code === 200) {
        toast.success(`批量${BATCH_ACTION_LABELS[action]}成功`)
        return true
      }
      toast.error(response.message || `批量${BATCH_ACTION_LABELS[action]}失败`)
    } catch (error) {
      console.error('批量操作失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      void this.loadUsers()
    }
    return false
  }
}

// 页面级共享状态实例。
export const userPageState = new UserPageState()
