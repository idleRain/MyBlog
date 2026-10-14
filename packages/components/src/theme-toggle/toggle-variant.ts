// 主题切换按钮的变体与无障碍名称解析。
// 与组件文件分离，使降级规则可被单测直接覆盖。

// outline 保留边框阴影供独立浮动场景使用，ghost 供导航栏与相邻图标按钮统一视觉。
export type ToggleVariant = 'ghost' | 'outline'

// 未传入 label 时使用的默认无障碍名称。
export const DEFAULT_TOGGLE_LABEL = '切换主题'

// 缺省形态取 outline，供登录页等独立浮动场景直接使用。
export const DEFAULT_TOGGLE_VARIANT: ToggleVariant = 'outline'

// 边框与阴影属于 outline 专属装饰，ghost 形态下由 Button 自身的悬停反馈承担视觉。
const OUTLINE_DECORATION = 'border bg-background/80 shadow-md backdrop-blur-sm hover:shadow-lg'

/** 解析无障碍名称，未传入时回退默认名称。 */
export function resolveToggleLabel(label?: string): string {
  return label ?? DEFAULT_TOGGLE_LABEL
}

/** 解析按钮形态，未传入时回退默认形态。 */
export function resolveToggleVariant(variant?: ToggleVariant): ToggleVariant {
  return variant ?? DEFAULT_TOGGLE_VARIANT
}

/** 解析形态专属装饰类，ghost 形态返回 undefined 以交由调用方的 cn 忽略。 */
export function resolveToggleDecoration(variant: ToggleVariant): string | undefined {
  return variant === 'outline' ? OUTLINE_DECORATION : undefined
}
