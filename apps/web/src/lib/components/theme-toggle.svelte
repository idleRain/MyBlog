<script lang="ts">
import { Sun, Moon } from '@lucide/svelte'
import { toggleMode } from 'mode-watcher'
import { cn } from '@myblog/shared'
import { Button } from '$ui/button'

// 未传入 label 时使用的默认无障碍名称。
const DEFAULT_TOGGLE_LABEL = '切换主题'

// outline 保留边框阴影供独立浮动场景使用，ghost 供导航栏与相邻图标按钮统一视觉。
type ToggleVariant = 'ghost' | 'outline'

// 定位类不内置于组件：独立浮动场景（如登录页）由调用方传入 fixed 定位类，
// 内嵌布局场景（如导航栏）保持文档流内定位，避免与页面元素重叠或遮挡关闭按钮。
const props = $props<{ class?: string; label?: string; variant?: ToggleVariant }>()

// variant 响应式派生，保证运行时切换形态时装饰类与按钮变体同步更新。
const variant = $derived(props.variant ?? 'outline')
// 边框与阴影属于 outline 专属装饰，ghost 形态下由 Button 自身的悬停反馈承担视觉。
const isOutlineVariant = $derived(variant === 'outline')

function handleToggle() {
  toggleMode()
}
</script>

<Button
  {variant}
  size="icon"
  onclick={handleToggle}
  class={cn(
    props.class,
    isOutlineVariant && 'border bg-background/80 shadow-md backdrop-blur-sm hover:shadow-lg'
  )}
  title={props.label ?? DEFAULT_TOGGLE_LABEL}
  aria-label={props.label ?? DEFAULT_TOGGLE_LABEL}
>
  <Sun
    class="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-transform dark:scale-0 dark:-rotate-90"
  />
  <Moon
    class="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-transform dark:scale-100 dark:rotate-0"
  />
  <span class="sr-only">{props.label ?? DEFAULT_TOGGLE_LABEL}</span>
</Button>
