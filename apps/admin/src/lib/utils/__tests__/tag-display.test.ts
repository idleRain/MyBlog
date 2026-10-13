import type { EnabledDictGroup } from '@myblog/api/modules/dict/types'
import { describe, expect, it } from 'vitest'

import { resolveTagStatusLabel, resolveTagStatusVariant } from '../tag-display'

// 构造最小可用的状态字典分组，仅覆盖解析所需的字段。
function statusDictWithVariant(variant: string, label: string): EnabledDictGroup {
  return {
    id: 1,
    code: 'tag_status',
    name: '标签状态',
    description: '',
    status: 1,
    sortOrder: 0,
    extra: null,
    createdAt: '',
    updatedAt: '',
    items: [
      {
        id: 1,
        typeId: 1,
        value: '1',
        label,
        description: '',
        sortOrder: 0,
        status: 1,
        extra: { variant },
        createdAt: '',
        updatedAt: ''
      }
    ]
  }
}

describe('标签状态展示解析', () => {
  it('字典缺失时回退内置配置', () => {
    expect(resolveTagStatusLabel(null, 1)).toBe('启用')
    expect(resolveTagStatusVariant(null, 1)).toBe('default')
  })

  it('字典命中时以字典文案与样式为准', () => {
    const dict = statusDictWithVariant('outline', '已上线')
    expect(resolveTagStatusLabel(dict, 1)).toBe('已上线')
    expect(resolveTagStatusVariant(dict, 1)).toBe('outline')
  })

  // 运维在字典里填入非法样式时必须回退，否则 Badge 会丢掉全部样式。
  it('字典样式取值非法时回退内置配置', () => {
    const dict = statusDictWithVariant('primary', '已上线')
    expect(resolveTagStatusVariant(dict, 1)).toBe('default')
  })

  it('未知状态取值返回未知文案与次级样式', () => {
    expect(resolveTagStatusLabel(null, 99)).toBe('未知')
    expect(resolveTagStatusVariant(null, 99)).toBe('secondary')
  })
})
