import { describe, expect, it } from 'vitest'

import { dictItemLabel, findDictItem } from './helpers.ts'
import type { EnabledDictGroup } from './types.ts'

// 构造标签状态字典分组的测试夹具。
function buildTagStatusGroup(): EnabledDictGroup {
  return {
    id: 1,
    code: 'tag_status',
    name: '标签状态',
    description: '',
    status: 1,
    sortOrder: 1,
    extra: null,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    items: [
      { id: 1, typeId: 1, value: '1', label: '启用', description: '', status: 1, sortOrder: 1, extra: null, createdAt: '', updatedAt: '' },
      { id: 2, typeId: 1, value: '0', label: '隐藏', description: '', status: 1, sortOrder: 2, extra: null, createdAt: '', updatedAt: '' }
    ]
  }
}

describe('findDictItem', () => {
  it('按存储值命中字典项', () => {
    const group = buildTagStatusGroup()
    expect(findDictItem(group, '1')?.label).toBe('启用')
    expect(findDictItem(group, '0')?.label).toBe('隐藏')
  })

  it('值未命中时返回 undefined', () => {
    expect(findDictItem(buildTagStatusGroup(), '2')).toBeUndefined()
  })

  it('分组缺失时安全返回 undefined', () => {
    expect(findDictItem(null, '1')).toBeUndefined()
    expect(findDictItem(undefined, '1')).toBeUndefined()
  })
})

describe('dictItemLabel', () => {
  it('命中时返回字典项显示名', () => {
    expect(dictItemLabel(buildTagStatusGroup(), '1', '未知')).toBe('启用')
  })

  it('未命中或分组缺失时回退缺省文案', () => {
    expect(dictItemLabel(buildTagStatusGroup(), '9', '未知')).toBe('未知')
    expect(dictItemLabel(null, '1', '未知')).toBe('未知')
  })
})
