import type { DictItem, EnabledDictGroup } from './types.ts'

// 字典消费辅助：业务侧以字符串形式的存储值定位字典项，
// 标签状态等整型状态经 String(status) 转换后即可直接匹配。

// findDictItem 在字典分组内按存储值定位字典项，未命中返回 undefined。
export function findDictItem(
  group: EnabledDictGroup | null | undefined,
  value: string
): DictItem | undefined {
  return group?.items.find(item => item.value === value)
}

// dictItemLabel 返回字典项的显示名，分组缺失或值未命中时回退传入的缺省文案。
export function dictItemLabel(
  group: EnabledDictGroup | null | undefined,
  value: string,
  fallback: string
): string {
  return findDictItem(group, value)?.label ?? fallback
}
