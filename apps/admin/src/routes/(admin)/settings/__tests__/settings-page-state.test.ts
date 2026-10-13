import type { Setting } from '@myblog/api/modules/setting/types'
import { beforeEach, describe, expect, it, vi } from 'vitest'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const settingApiMock = {
  list: vi.fn(),
  update: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn(),
  info: vi.fn()
}

vi.mock('$lib/api', () => ({ SettingAPI: settingApiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { emptySettingsPageData, settingsPageState } = await import('../settings-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 构造一个设置项替身，字段取值只覆盖被测逻辑关心的部分。
function makeSetting(keyName: string, value: string, groupName = 'site'): Setting {
  return {
    keyName,
    value,
    label: keyName,
    defaultValue: '',
    description: '',
    type: 'string',
    groupName,
    isPublic: false,
    isReadonly: false,
    isSensitive: false,
    validationRule: '',
    sortOrder: 0
  } as unknown as Setting
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  settingsPageState.initialize(emptySettingsPageData())
  settingsPageState.isLoading = false
  settingsPageState.isSaving = false
}

describe('系统设置页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果展开表单当前值与初值', () => {
    const settings = [makeSetting('site_title', '我的博客'), makeSetting('site_desc', '简介')]

    settingsPageState.initialize({ settings })

    expect(settingsPageState.settings).toEqual(settings)
    expect(settingsPageState.values).toEqual({ site_title: '我的博客', site_desc: '简介' })
    expect(settingsPageState.originalValues).toEqual({
      site_title: '我的博客',
      site_desc: '简介'
    })
  })

  it('初值与当前值互不共享引用', () => {
    settingsPageState.initialize({ settings: [makeSetting('site_title', '我的博客')] })

    settingsPageState.setValue('site_title', '新标题')

    expect(settingsPageState.values.site_title).toBe('新标题')
    expect(settingsPageState.originalValues.site_title).toBe('我的博客')
  })

  it('空数据下不产生分组与改动项', () => {
    expect(settingsPageState.groups).toEqual([])
    expect(settingsPageState.collectChangedItems()).toEqual([])
    expect(settingsPageState.isLoading).toBe(false)
    expect(settingsPageState.isSaving).toBe(false)
  })
})

describe('系统设置页状态：分组', () => {
  beforeEach(resetState)

  it('按分组名聚合设置项并保持原始顺序', () => {
    settingsPageState.initialize({
      settings: [
        makeSetting('site_title', '标题', 'site'),
        makeSetting('sec_token', '令牌', 'security'),
        makeSetting('site_desc', '简介', 'site')
      ]
    })

    const groups = settingsPageState.groups

    expect(groups.map(group => group.groupName)).toEqual(['site', 'security'])
    expect(groups[0]?.items.map(item => item.keyName)).toEqual(['site_title', 'site_desc'])
    expect(groups[1]?.items.map(item => item.keyName)).toEqual(['sec_token'])
  })
})

describe('系统设置页状态：改动汇总', () => {
  beforeEach(resetState)

  it('只汇总取值发生变化的设置项', () => {
    settingsPageState.initialize({
      settings: [makeSetting('site_title', '标题'), makeSetting('site_desc', '简介')]
    })

    settingsPageState.setValue('site_desc', '新简介')

    expect(settingsPageState.collectChangedItems()).toEqual([
      { keyName: 'site_desc', value: '新简介' }
    ])
  })

  it('值被改回初值后不再计入改动', () => {
    settingsPageState.initialize({ settings: [makeSetting('site_title', '标题')] })

    settingsPageState.setValue('site_title', '临时值')
    settingsPageState.setValue('site_title', '标题')

    expect(settingsPageState.collectChangedItems()).toEqual([])
  })
})

describe('系统设置页状态：保存', () => {
  beforeEach(resetState)

  it('没有改动时仅提示且不发起请求', async () => {
    settingsPageState.initialize({ settings: [makeSetting('site_title', '标题')] })

    await settingsPageState.saveSettings()

    expect(settingApiMock.update).not.toHaveBeenCalled()
    expect(toastMock.info).toHaveBeenCalledWith('没有需要保存的修改')
  })

  it('保存成功后提示改动条数并重新加载', async () => {
    settingApiMock.update.mockResolvedValue({ code: SUCCESS_CODE, data: { settings: [] } })
    settingApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { settings: [makeSetting('site_title', '新标题')] }
    })
    settingsPageState.initialize({ settings: [makeSetting('site_title', '标题')] })
    settingsPageState.setValue('site_title', '新标题')

    await settingsPageState.saveSettings()

    expect(settingApiMock.update).toHaveBeenCalledWith({
      items: [{ keyName: 'site_title', value: '新标题' }]
    })
    expect(toastMock.success).toHaveBeenCalledWith('已保存 1 项设置')
    expect(settingApiMock.list).toHaveBeenCalledTimes(1)
    expect(settingsPageState.values.site_title).toBe('新标题')
    expect(settingsPageState.isSaving).toBe(false)
  })

  it('保存进行中重复提交不发起第二次请求', async () => {
    settingApiMock.update.mockResolvedValue({ code: SUCCESS_CODE, data: { settings: [] } })
    settingApiMock.list.mockResolvedValue({ code: SUCCESS_CODE, data: { settings: [] } })
    settingsPageState.initialize({ settings: [makeSetting('site_title', '标题')] })
    settingsPageState.setValue('site_title', '新标题')
    settingsPageState.isSaving = true

    await settingsPageState.saveSettings()

    expect(settingApiMock.update).not.toHaveBeenCalled()
  })

  it('业务码非成功时提示后端文案且保留改动', async () => {
    settingApiMock.update.mockResolvedValue({ code: 400, message: '校验未通过' })
    settingsPageState.initialize({ settings: [makeSetting('site_title', '标题')] })
    settingsPageState.setValue('site_title', '新标题')

    await settingsPageState.saveSettings()

    expect(toastMock.error).toHaveBeenCalledWith('校验未通过')
    expect(settingsPageState.values.site_title).toBe('新标题')
    expect(settingsPageState.originalValues.site_title).toBe('标题')
    expect(settingsPageState.isSaving).toBe(false)
  })

  it('请求抛错时提示网络错误并复位保存态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    settingApiMock.update.mockRejectedValue(new Error('断网'))
    settingsPageState.initialize({ settings: [makeSetting('site_title', '标题')] })
    settingsPageState.setValue('site_title', '新标题')

    await settingsPageState.saveSettings()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(settingsPageState.isSaving).toBe(false)
    consoleError.mockRestore()
  })
})

describe('系统设置页状态：加载', () => {
  beforeEach(resetState)

  it('加载成功重置表单当前值与初值', async () => {
    settingApiMock.list.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { settings: [makeSetting('site_title', '服务端标题')] }
    })

    await settingsPageState.loadSettings()

    expect(settingsPageState.settings).toHaveLength(1)
    expect(settingsPageState.values).toEqual({ site_title: '服务端标题' })
    expect(settingsPageState.originalValues).toEqual({ site_title: '服务端标题' })
    expect(settingsPageState.isLoading).toBe(false)
  })

  it('加载失败时提示后端文案且不改动现有表单', async () => {
    settingApiMock.list.mockResolvedValue({ code: 500, message: '查询失败' })
    settingsPageState.initialize({ settings: [makeSetting('site_title', '标题')] })

    await settingsPageState.loadSettings()

    expect(toastMock.error).toHaveBeenCalledWith('查询失败')
    expect(settingsPageState.values).toEqual({ site_title: '标题' })
    expect(settingsPageState.isLoading).toBe(false)
  })

  it('请求抛错时提示网络错误并复位加载态', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    settingApiMock.list.mockRejectedValue(new Error('断网'))

    await settingsPageState.loadSettings()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(settingsPageState.isLoading).toBe(false)
    consoleError.mockRestore()
  })
})
