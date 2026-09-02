import type { KyInstance } from 'ky'
import type { SettingsResponse, UpdateSettingsRequest } from './types.ts'

/**
 * 创建设置接口模块，依赖注入的 http 客户端由调用方提供。
 */
export function createSettingAPI(request: KyInstance) {
  return {
    // 管理端：获取全部设置项，敏感值已掩码。
    list(): Promise<SettingsResponse> {
      return request.post('admin/settings/list', { json: {} }).json()
    },

    // 管理端：批量更新设置项，只读项后端拒绝修改。
    update(params: UpdateSettingsRequest): Promise<SettingsResponse> {
      return request.post('admin/settings/update', { json: params }).json()
    }
  }
}

export type SettingAPI = ReturnType<typeof createSettingAPI>

export type { SettingsResponse, UpdateSettingsRequest }
