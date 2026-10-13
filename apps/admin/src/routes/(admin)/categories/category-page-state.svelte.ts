import type {
  Category,
  CategoryTreeNode,
  CreateCategoryRequest,
  UpdateCategoryRequest
} from '@myblog/api/modules/category/types'
import { toast } from 'svelte-sonner'

import { CATEGORY_STATUS_FILTER_ALL, type CategoryRow } from '$lib/constants/category'
import { CategoryAPI } from '$lib/api'

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface CategoryPageData {
  tree: CategoryTreeNode[]
}

// 空页面数据，load 失败或无数据时使用。
export function emptyCategoryPageData(): CategoryPageData {
  return { tree: [] }
}

// 将分类树递归扁平化为带缩进深度的行序列，深度由渲染层换算为缩进量。
function flattenTree(nodes: CategoryTreeNode[], depth = 0): CategoryRow[] {
  return nodes.flatMap(node => [
    { category: node, depth },
    ...flattenTree(node.children ?? [], depth + 1)
  ])
}

// 分类管理页状态：集中持有分类树行、筛选条件与全部读写操作，页面组件只做渲染。
class CategoryPageState {
  rows = $state<CategoryRow[]>([])
  searchQuery = $state('')
  // 状态筛选使用字符串值：空串为全部，其余为分类状态值的字符串形式。
  statusFilter = $state<string>(CATEGORY_STATUS_FILTER_ALL)
  // 初始数据由 load 提供，首次渲染不需要加载态。
  isLoading = $state(false)

  // 按名称与 URL 标识匹配后再按显示状态过滤，树数据为全量，无需服务端分页。
  get filteredRows(): CategoryRow[] {
    const query = this.searchQuery.trim().toLowerCase()
    return this.rows.filter(row => {
      const nameHit = !query || row.category.name.toLowerCase().includes(query)
      const slugHit = !query || (row.category.slug ?? '').toLowerCase().includes(query)
      const statusHit =
        this.statusFilter === CATEGORY_STATUS_FILTER_ALL ||
        row.category.status === Number(this.statusFilter)
      return (nameHit || slugHit) && statusHit
    })
  }

  // 全量分类列表，供表单弹窗选择父分类，由扁平行去掉缩进深度得到。
  get flatCategories(): Category[] {
    return this.rows.map(row => row.category)
  }

  // 树中是否已存在分类，用于空态区分无数据与无匹配。
  get hasCategories(): boolean {
    return this.rows.length > 0
  }

  // 分类总数，取全量树行的数量，客户端筛选不改变该口径。
  get total(): number {
    return this.rows.length
  }

  // 用 load 结果初始化页面状态，重复挂载时以最新数据覆盖。
  initialize(data: CategoryPageData) {
    this.rows = flattenTree(data.tree)
  }

  // 重置筛选条件并清空搜索结果。
  resetFilters() {
    this.searchQuery = ''
    this.statusFilter = CATEGORY_STATUS_FILTER_ALL
  }

  // 加载分类树并重新生成扁平行，树为全量数据，刷新后保留当前筛选条件。
  async loadCategories() {
    this.isLoading = true
    try {
      const response = await CategoryAPI.getTree()
      if (response.code === 200 && response.data) {
        this.rows = flattenTree(response.data.tree ?? [])
      } else {
        toast.error(response.message || '加载分类失败')
      }
    } catch (error) {
      console.error('加载分类失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 创建分类，成功后重新加载分类树。
  async createCategory(payload: CreateCategoryRequest): Promise<boolean> {
    try {
      const response = await CategoryAPI.create(payload)
      if (response.code === 200 && response.data) {
        toast.success('分类创建成功')
        await this.loadCategories()
        return true
      }
      toast.error(response.message || '保存分类失败')
    } catch (error) {
      console.error('保存分类失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 更新分类，成功后重新加载分类树。
  async updateCategory(payload: UpdateCategoryRequest): Promise<boolean> {
    try {
      const response = await CategoryAPI.update(payload)
      if (response.code === 200 && response.data) {
        toast.success('分类更新成功')
        await this.loadCategories()
        return true
      }
      toast.error(response.message || '保存分类失败')
    } catch (error) {
      console.error('保存分类失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 删除分类，存在子分类时后端拒绝，拒绝原因沿用后端文案提示。
  async removeCategory(id: number): Promise<boolean> {
    try {
      const response = await CategoryAPI.delete(id)
      if (response.code === 200) {
        toast.success('分类删除成功')
        await this.loadCategories()
        return true
      }
      toast.error(response.message || '删除分类失败')
    } catch (error) {
      console.error('删除分类失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }
}

// 页面级共享状态实例。
export const categoryPageState = new CategoryPageState()
