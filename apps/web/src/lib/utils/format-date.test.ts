import { describe, expect, it } from 'vitest'
import { formatDate } from './format-date'

describe('目录日期格式化', () => {
  it('截取 ISO 字符串的日期段，输出 YYYY-MM-DD', () => {
    expect(formatDate('2026-10-08 09:30:00')).toBe('2026-10-08')
    expect(formatDate('2026-10-08T09:30:00Z')).toBe('2026-10-08')
  })

  // 该函数刻意不做时区换算，直接截取原始字符串。
  // 若改为经 Date 解析后再格式化，跨时区的正午前后会整体偏移一天，
  // 此用例把「不换算」这一有意行为固定下来。
  it('不做时区换算，给定日期段原样保留', () => {
    expect(formatDate('2026-01-01T23:59:59+08:00')).toBe('2026-01-01')
    expect(formatDate('2026-12-31T00:00:00-05:00')).toBe('2026-12-31')
  })

  it('长度不足时按实际长度返回，不补位也不抛错', () => {
    expect(formatDate('2026')).toBe('2026')
    expect(formatDate('')).toBe('')
  })
})
