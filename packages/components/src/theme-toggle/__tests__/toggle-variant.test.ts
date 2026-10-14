// 主题切换按钮的形态与名称解析单元测试。
// 覆盖默认值回退与 outline 专属装饰两条分支。
import {
  DEFAULT_TOGGLE_LABEL,
  DEFAULT_TOGGLE_VARIANT,
  resolveToggleDecoration,
  resolveToggleLabel,
  resolveToggleVariant
} from '../toggle-variant'
import { describe, expect, it } from 'vitest'

describe('无障碍名称解析', () => {
  it('未传入名称时回退默认名称', () => {
    expect(resolveToggleLabel()).toBe(DEFAULT_TOGGLE_LABEL)
  })

  it('传入名称时原样返回', () => {
    expect(resolveToggleLabel('Toggle theme')).toBe('Toggle theme')
  })
})

describe('形态解析', () => {
  it('未传入形态时回退默认形态', () => {
    expect(resolveToggleVariant()).toBe(DEFAULT_TOGGLE_VARIANT)
  })

  it('传入形态时原样返回', () => {
    expect(resolveToggleVariant('ghost')).toBe('ghost')
  })
})

describe('形态专属装饰', () => {
  it('outline 形态附带边框与阴影装饰', () => {
    const decoration = resolveToggleDecoration('outline')

    expect(decoration).toContain('border')
    expect(decoration).toContain('shadow-md')
  })

  it('ghost 形态不附带装饰，交由按钮自身悬停反馈', () => {
    expect(resolveToggleDecoration('ghost')).toBeUndefined()
  })
})
