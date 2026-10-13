import type {
  CreateFriendlyLinkRequest,
  FriendlyLink,
  LinkStatus,
  UpdateFriendlyLinkRequest
} from '@myblog/api/modules/friendlyLink/types'
import { toast } from 'svelte-sonner'

import { LINK_PAGE_SIZE, LINK_STATUS_ACTIONS, type LinkStatusAction } from '$lib/constants/link'
import { FriendlyLinkAPI } from '$lib/api'

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface LinkPageData {
  links: FriendlyLink[]
  total: number
}

// 空页面数据，load 失败时使用。
export function emptyLinkPageData(): LinkPageData {
  return { links: [], total: 0 }
}

// 友链管理页状态：集中持有列表、筛选条件与全部读写操作，页面组件只做渲染。
class LinkPageState {
  links = $state<FriendlyLink[]>([])
  total = $state(0)
  currentPage = $state(1)
  // 状态筛选使用字符串值：空串为全部，其余为友链状态枚举值。
  statusFilter = $state<LinkStatus | ''>('')
  // 初始数据由 load 提供，首次渲染不需要加载态。
  isLoading = $state(false)

  // 总页数，分页控件与页码文案共用同一口径。
  get totalPages(): number {
    return Math.max(1, Math.ceil(this.total / LINK_PAGE_SIZE))
  }

  // 是否存在生效中的筛选条件，用于空态区分无数据与无匹配。
  get hasActiveFilters(): boolean {
    return this.statusFilter !== ''
  }

  // 用 load 结果初始化页面状态，重复挂载时以最新数据覆盖。
  initialize(data: LinkPageData) {
    this.links = data.links
    this.total = data.total
  }

  // 加载友链列表，携带当前页码与状态筛选，未筛选时不传 status 字段。
  async loadLinks() {
    this.isLoading = true
    try {
      const response = await FriendlyLinkAPI.adminList({
        page: this.currentPage,
        pageSize: LINK_PAGE_SIZE,
        ...(this.statusFilter ? { status: this.statusFilter } : {})
      })
      if (response.code === 200 && response.data) {
        this.links = response.data.links ?? []
        this.total = response.data.total ?? 0
      } else {
        toast.error(response.message || '加载友链失败')
      }
    } catch (error) {
      console.error('加载友链失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 状态筛选变更后回到第一页并重新加载。
  setStatusFilter(value: string) {
    this.statusFilter = value as LinkStatus | ''
    this.currentPage = 1
    void this.loadLinks()
  }

  // 重置筛选条件并回到第一页。
  resetAndReload() {
    this.statusFilter = ''
    this.currentPage = 1
    void this.loadLinks()
  }

  // 切换页码并重新加载。
  changePage(page: number) {
    this.currentPage = page
    void this.loadLinks()
  }

  // 创建友链，成功后刷新当前页。
  async createLink(payload: CreateFriendlyLinkRequest): Promise<boolean> {
    try {
      const response = await FriendlyLinkAPI.create(payload)
      if (response.code === 200 && response.data) {
        toast.success('友链添加成功')
        await this.loadLinks()
        return true
      }
      toast.error(response.message || '保存友链失败')
    } catch (error) {
      console.error('保存友链失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 更新友链，成功后刷新当前页。
  async updateLink(payload: UpdateFriendlyLinkRequest): Promise<boolean> {
    try {
      const response = await FriendlyLinkAPI.update(payload)
      if (response.code === 200 && response.data) {
        toast.success('友链更新成功')
        await this.loadLinks()
        return true
      }
      toast.error(response.message || '保存友链失败')
    } catch (error) {
      console.error('保存友链失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 执行审核动作，成功提示由该动作在当前状态下的文案拼接，随后刷新列表。
  async reviewLink(link: FriendlyLink, key: LinkStatusAction): Promise<boolean> {
    try {
      const response = await FriendlyLinkAPI[key](link.id)
      if (response.code === 200) {
        const actionLabel = LINK_STATUS_ACTIONS[link.status].find(item => item.key === key)?.label
        toast.success(`${actionLabel}成功`)
        await this.loadLinks()
        return true
      }
      toast.error(response.message || '操作失败')
    } catch (error) {
      console.error('友链操作失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 删除友链，成功判定只要求业务码，响应无数据同样视为完成。
  async removeLink(id: number): Promise<boolean> {
    try {
      const response = await FriendlyLinkAPI.delete(id)
      if (response.code === 200) {
        toast.success('友链删除成功')
        await this.loadLinks()
        return true
      }
      toast.error(response.message || '删除友链失败')
    } catch (error) {
      console.error('删除友链失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }
}

// 页面级共享状态实例。
export const linkPageState = new LinkPageState()
