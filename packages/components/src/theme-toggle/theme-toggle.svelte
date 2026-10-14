<script lang="ts">
import {
  resolveToggleDecoration,
  resolveToggleLabel,
  resolveToggleVariant,
  type ToggleVariant
} from './toggle-variant.js'
import { Button } from '@myblog/ui/button'
import { Sun, Moon } from '@lucide/svelte'
import { toggleMode } from 'mode-watcher'
import { cn } from '@myblog/shared'

// 定位类不内置于组件：独立浮动场景（如登录页）由调用方传入 fixed 定位类，
// 内嵌布局场景（如导航栏）保持文档流内定位，避免与页面元素重叠或遮挡关闭按钮。
const props = $props<{ class?: string; label?: string; variant?: ToggleVariant }>()

// 形态与名称经纯函数解析，降级规则集中在 toggle-variant.ts 并有单测覆盖。
const variant = $derived(resolveToggleVariant(props.variant))
const label = $derived(resolveToggleLabel(props.label))
const decoration = $derived(resolveToggleDecoration(variant))
</script>

<Button
  {variant}
  size="icon"
  onclick={toggleMode}
  class={cn(props.class, decoration)}
  title={label}
  aria-label={label}
>
  <Sun
    class="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-transform dark:scale-0 dark:-rotate-90"
  />
  <Moon
    class="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-transform dark:scale-100 dark:rotate-0"
  />
  <span class="sr-only">{label}</span>
</Button>
