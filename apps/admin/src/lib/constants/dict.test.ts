import { describe, expect, it } from 'vitest'

import { DICT_STATUS_DISABLED, DICT_STATUS_ENABLED, dictStatusOf } from './dict'

describe('字典状态展示配置', () => {
  it('生效状态返回主色样式与生效文案', () => {
    expect(dictStatusOf(DICT_STATUS_ENABLED)).toEqual({ label: '生效', variant: 'default' })
  })

  it('停用状态返回次级样式与停用文案', () => {
    expect(dictStatusOf(DICT_STATUS_DISABLED)).toEqual({ label: '停用', variant: 'secondary' })
  })

  // 后端新增状态取值而前端未同步时，列表必须保留可见的状态标注而不是空白。
  it('未知状态回退为未知文案与次级样式', () => {
    expect(dictStatusOf(99)).toEqual({ label: '未知', variant: 'secondary' })
  })
})
