import type {
  CreateTagRequest,
  Tag,
  TagStatus,
  UpdateTagRequest
} from '@myblog/api/modules/tag/types'
import type { EnabledDictGroup } from '@myblog/api/modules/dict/types'
import { DICT_TYPE_CODE_TAG_STATUS } from '@myblog/api'
import { debounce } from '@myblog/shared'
import { toast } from 'svelte-sonner'

import { TAG_PAGE_SIZE } from '$lib/constants/tag'
import { DictAPI, TagAPI } from '$lib/api'

// 搜索输入停顿时长，避免每次按键都触发请求。
const SEARCH_DEBOUNCE_MS = 300

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 标签状态筛选项的降级来源，字典不可用时使用。
const FALLBACK_STATUS_OPTIONS = [
  { value: '1', label: '启用' },
  { value: '0', label: '隐藏' }
]

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface TagPageData {
  tags: Tag[]
  total: number
  statusDict: EnabledDictGroup | null
}

// 空页面数据，load 失败时使用。
export function emptyTagPageData(): TagPageData {
  return { tags: [], total: 0, statusDict: null }
}

// 标签管理页状态：集中持有列表、筛选条件与全部读写操作，页面组件只做渲染。
class TagPageState {
  tags = $state<Tag[]>([])
  total = $state(0)
  currentPage = $state(1)
  search = $state('')
  // 状态筛选使用字符串值：空串为全部，其余为字典项值，与标签整型状态的字符串形式对齐。
  statusFilter = $state('')
  // 热门筛选使用字符串值：空串为全部，'1' 为仅热门。
  hotFilter = $state('')
  statusDict = $state<EnabledDictGroup | null>(null)
  isLoading = $state(false)

  // 总页数，分页控件与页码文案共用同一口径。
  get totalPages(): number {
    return Math.max(1, Math.ceil(this.total / TAG_PAGE_SIZE))
  }

  // 状态筛选项按字典项排序生成；停用字典项仍参与筛选，保证存量数据可被过滤。
  get statusFilterOptions(): { value: string; label: string }[] {
    if (!this.statusDict) {
      return FALLBACK_STATUS_OPTIONS
    }
    return [...this.statusDict.items]
      .sort((left, right) => left.sortOrder - right.sortOrder)
      .map(item => ({ value: item.value, label: item.label }))
  }

  // 是否存在生效中的筛选条件，用于空态区分无数据与无匹配。
  get hasActiveFilters(): boolean {
    return this.search.trim() !== '' || this.statusFilter !== '' || this.hotFilter !== ''
  }

  // 用 load 结果初始化页面状态，重复挂载时以最新数据覆盖。
  initialize(data: TagPageData) {
    this.tags = data.tags
    this.total = data.total
    this.statusDict = data.statusDict
  }

  // 加载标签列表，携带搜索、状态与热门筛选。
  async loadTags() {
    this.isLoading = true
    try {
      const response = await TagAPI.adminList({
        page: this.currentPage,
        pageSize: TAG_PAGE_SIZE,
        ...(this.statusFilter !== '' ? { status: Number(this.statusFilter) as TagStatus } : {}),
        ...(this.hotFilter === '1' ? { isHot: true } : {}),
        ...(this.search.trim() ? { search: this.search.trim() } : {})
      })
      if (response.code === 200 && response.data) {
        this.tags = response.data.tags ?? []
        this.total = response.data.total ?? 0
      } else {
        toast.error(response.message || '加载标签失败')
      }
    } catch (error) {
      console.error('加载标签失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 搜索词防抖应用，连续输入只在停顿后请求一次。
  applySearchDebounced = debounce(() => {
    this.currentPage = 1
    void this.loadTags()
  }, SEARCH_DEBOUNCE_MS)

  // 状态筛选变更后回到第一页并重新加载。
  setStatusFilter(value: string) {
    this.statusFilter = value
    this.currentPage = 1
    void this.loadTags()
  }

  // 热门筛选变更后回到第一页并重新加载。
  setHotFilter(value: string) {
    this.hotFilter = value
    this.currentPage = 1
    void this.loadTags()
  }

  // 重置全部筛选条件并回到第一页。
  resetFilters() {
    this.search = ''
    this.statusFilter = ''
    this.hotFilter = ''
    this.currentPage = 1
    void this.loadTags()
  }

  // 切换页码并重新加载。
  changePage(page: number) {
    this.currentPage = page
    void this.loadTags()
  }

  // 创建标签，成功后刷新当前页。
  async createTag(payload: CreateTagRequest): Promise<boolean> {
    try {
      const response = await TagAPI.create(payload)
      if (response.code === 200 && response.data) {
        toast.success('标签创建成功')
        await this.loadTags()
        return true
      }
      toast.error(response.message || '保存标签失败')
    } catch (error) {
      console.error('保存标签失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 更新标签，成功后刷新当前页。
  async updateTag(payload: UpdateTagRequest): Promise<boolean> {
    try {
      const response = await TagAPI.update(payload)
      if (response.code === 200 && response.data) {
        toast.success('标签更新成功')
        await this.loadTags()
        return true
      }
      toast.error(response.message || '保存标签失败')
    } catch (error) {
      console.error('保存标签失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 删除标签，已挂载文章的标签处理由后端裁决。
  async removeTag(id: number): Promise<boolean> {
    try {
      const response = await TagAPI.delete(id)
      if (response.code === 200) {
        toast.success('标签删除成功')
        await this.loadTags()
        return true
      }
      toast.error(response.message || '删除标签失败')
    } catch (error) {
      console.error('删除标签失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 加载标签状态字典，失败时静默回退内置配置，不阻断列表使用。
  async loadStatusDict() {
    try {
      const response = await DictAPI.getByType(DICT_TYPE_CODE_TAG_STATUS)
      if (response.code === 200 && response.data) {
        this.statusDict = response.data
      }
    } catch (error) {
      console.error('加载标签状态字典失败:', error)
    }
  }
}

// 页面级共享状态实例。
export const tagPageState = new TagPageState()
