import type {
  CreateDictItemRequest,
  CreateDictTypeRequest,
  DictItem,
  DictType,
  UpdateDictItemRequest,
  UpdateDictTypeRequest
} from '@myblog/api/modules/dict/types'
import { toast } from 'svelte-sonner'

import { DictAPI } from '$lib/api'

// 字典管理页单页容量：字典为小体量配置数据，整页加载即可覆盖常规场景。
export const DICT_PAGE_SIZE = 50

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface DictPageData {
  types: DictType[]
  typesTotal: number
  selectedTypeId: number | null
  items: DictItem[]
  itemsTotal: number
}

// 空页面数据，load 失败或无数据时使用。
export function emptyDictPageData(): DictPageData {
  return { types: [], typesTotal: 0, selectedTypeId: null, items: [], itemsTotal: 0 }
}

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 字典管理页状态：集中持有类型与字典项列表及全部读写操作，页面组件只做渲染。
class DictPageState {
  types = $state<DictType[]>([])
  typesTotal = $state(0)
  selectedTypeId = $state<number | null>(null)
  items = $state<DictItem[]>([])
  itemsTotal = $state(0)
  isLoadingTypes = $state(false)
  isLoadingItems = $state(false)

  // 当前选中的字典类型，未选中时返回 null。
  get selectedType(): DictType | null {
    return this.types.find(type => type.id === this.selectedTypeId) ?? null
  }

  // 用 load 结果初始化页面状态，重复挂载时以最新数据覆盖。
  initialize(data: DictPageData) {
    this.types = data.types
    this.typesTotal = data.typesTotal
    this.selectedTypeId = data.selectedTypeId
    this.items = data.items
    this.itemsTotal = data.itemsTotal
  }

  // 重新加载类型列表，选中类型被删除或缺失时回退到首个类型。
  async loadTypes() {
    this.isLoadingTypes = true
    try {
      const response = await DictAPI.adminListTypes({ page: 1, pageSize: DICT_PAGE_SIZE })
      if (response.code === 200 && response.data) {
        this.types = response.data.types ?? []
        this.typesTotal = response.data.total ?? 0

        // 选中类型仍在列表中则保持不动，否则回退首个类型或清空字典项。
        if (this.types.length === 0) {
          this.selectedTypeId = null
          this.items = []
          this.itemsTotal = 0
        } else if (!this.selectedType) {
          await this.selectType(this.types[0]!.id)
        }
      } else {
        toast.error(response.message || '加载字典类型失败')
      }
    } catch (error) {
      console.error('加载字典类型失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoadingTypes = false
    }
  }

  // 选中字典类型并加载其字典项，重复选中当前类型时跳过重复加载。
  async selectType(id: number) {
    if (this.selectedTypeId === id && !this.isLoadingItems) return
    this.selectedTypeId = id
    await this.loadItems()
  }

  // 加载当前选中类型的字典项。
  async loadItems() {
    if (this.selectedTypeId == null) {
      this.items = []
      this.itemsTotal = 0
      return
    }

    this.isLoadingItems = true
    try {
      const response = await DictAPI.adminListItems({
        typeId: this.selectedTypeId,
        page: 1,
        pageSize: DICT_PAGE_SIZE
      })
      if (response.code === 200 && response.data) {
        this.items = response.data.items ?? []
        this.itemsTotal = response.data.total ?? 0
      } else {
        toast.error(response.message || '加载字典项失败')
      }
    } catch (error) {
      console.error('加载字典项失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoadingItems = false
    }
  }

  // 创建字典类型，成功后刷新类型列表。
  async createType(payload: CreateDictTypeRequest): Promise<boolean> {
    try {
      const response = await DictAPI.adminCreateType(payload)
      if (response.code === 200) {
        toast.success('字典类型创建成功')
        await this.loadTypes()
        return true
      }
      toast.error(response.message || '字典类型创建失败')
    } catch (error) {
      console.error('字典类型创建失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 更新字典类型，成功后刷新类型列表。
  async updateType(payload: UpdateDictTypeRequest): Promise<boolean> {
    try {
      const response = await DictAPI.adminUpdateType(payload)
      if (response.code === 200) {
        toast.success('字典类型更新成功')
        await this.loadTypes()
        return true
      }
      toast.error(response.message || '字典类型更新失败')
    } catch (error) {
      console.error('字典类型更新失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 删除字典类型，成功后清空选中并刷新类型列表。
  async removeType(id: number): Promise<boolean> {
    try {
      const response = await DictAPI.adminDeleteType(id)
      if (response.code === 200) {
        toast.success('字典类型删除成功')
        this.selectedTypeId = null
        this.items = []
        this.itemsTotal = 0
        await this.loadTypes()
        return true
      }
      toast.error(response.message || '字典类型删除失败')
    } catch (error) {
      console.error('字典类型删除失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 创建字典项，成功后刷新当前类型的字典项。
  async createItem(payload: CreateDictItemRequest): Promise<boolean> {
    try {
      const response = await DictAPI.adminCreateItem(payload)
      if (response.code === 200) {
        toast.success('字典项创建成功')
        await this.loadItems()
        return true
      }
      toast.error(response.message || '字典项创建失败')
    } catch (error) {
      console.error('字典项创建失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 更新字典项，成功后刷新当前类型的字典项。
  async updateItem(payload: UpdateDictItemRequest): Promise<boolean> {
    try {
      const response = await DictAPI.adminUpdateItem(payload)
      if (response.code === 200) {
        toast.success('字典项更新成功')
        await this.loadItems()
        return true
      }
      toast.error(response.message || '字典项更新失败')
    } catch (error) {
      console.error('字典项更新失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }

  // 删除字典项，成功后刷新当前类型的字典项。
  async removeItem(id: number): Promise<boolean> {
    try {
      const response = await DictAPI.adminDeleteItem(id)
      if (response.code === 200) {
        toast.success('字典项删除成功')
        await this.loadItems()
        return true
      }
      toast.error(response.message || '字典项删除失败')
    } catch (error) {
      console.error('字典项删除失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    }
    return false
  }
}

// 页面级共享状态实例。
export const dictPageState = new DictPageState()
