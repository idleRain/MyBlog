import type { Setting, UpdateSettingItem } from '@myblog/api/modules/setting/types'
import { toast } from 'svelte-sonner'
import { SettingAPI } from '$lib/api'

// 网络异常的统一提示文案。
const NETWORK_ERROR_MESSAGE = '网络错误，请稍后重试'

// 初始页面数据形状，由 +page.ts 的 load 函数提供。
export interface SettingsPageData {
  settings: Setting[]
}

// 空页面数据，load 失败时使用。
export function emptySettingsPageData(): SettingsPageData {
  return { settings: [] }
}

// 按 keyName 展开设置项键值，表单初值与当前值共用同一口径。
function collectValues(settings: Setting[]): Record<string, string> {
  const values: Record<string, string> = {}
  for (const setting of settings) {
    values[setting.keyName] = setting.value
  }
  return values
}

// 设置项分组，分组名与成员按设置项的原始顺序聚合。
interface SettingGroup {
  groupName: string
  items: Setting[]
}

// 系统设置页状态：集中持有设置项、表单值与全部读写操作，页面组件只做渲染。
class SettingsPageState {
  settings = $state<Setting[]>([])
  values = $state<Record<string, string>>({})
  originalValues = $state<Record<string, string>>({})
  isLoading = $state(false)
  isSaving = $state(false)

  // 按分组名聚合设置项并保持原始顺序。
  get groups(): SettingGroup[] {
    const groups: Record<string, Setting[]> = {}
    for (const setting of this.settings) {
      const group = groups[setting.groupName] ?? []
      group.push(setting)
      groups[setting.groupName] = group
    }
    return Object.entries(groups).map(([groupName, items]) => ({ groupName, items }))
  }

  // 用设置项初始化页面状态，表单当前值与初值同时归位。
  initialize(data: SettingsPageData) {
    this.settings = data.settings
    const values = collectValues(data.settings)
    this.values = { ...values }
    this.originalValues = values
  }

  // 记录单个设置项的编辑值。
  setValue(keyName: string, value: string) {
    this.values[keyName] = value
  }

  // 加载全部设置项并重置表单当前值与初值。
  async loadSettings() {
    this.isLoading = true
    try {
      const response = await SettingAPI.list()
      if (response.code === 200 && response.data) {
        this.initialize({ settings: response.data.settings ?? [] })
      } else {
        toast.error(response.message || '加载设置失败')
      }
    } catch (error) {
      console.error('加载设置失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isLoading = false
    }
  }

  // 汇总有改动的设置项，敏感掩码未改动时不会进入提交载荷。
  collectChangedItems(): UpdateSettingItem[] {
    return Object.keys(this.values)
      .filter(keyName => this.originalValues[keyName] !== this.values[keyName])
      .map(keyName => ({ keyName, value: this.values[keyName] ?? '' }))
  }

  // 批量保存有改动的设置项，成功后重新加载以对齐服务端值。
  async saveSettings() {
    if (this.isSaving) return

    const items = this.collectChangedItems()
    if (items.length === 0) {
      toast.info('没有需要保存的修改')
      return
    }

    this.isSaving = true
    try {
      const response = await SettingAPI.update({ items })
      if (response.code === 200 && response.data) {
        toast.success(`已保存 ${items.length} 项设置`)
        await this.loadSettings()
      } else {
        toast.error(response.message || '保存设置失败')
      }
    } catch (error) {
      console.error('保存设置失败:', error)
      toast.error(NETWORK_ERROR_MESSAGE)
    } finally {
      this.isSaving = false
    }
  }
}

// 页面级共享状态实例。
export const settingsPageState = new SettingsPageState()
