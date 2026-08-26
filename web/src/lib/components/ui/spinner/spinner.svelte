<script lang="ts">
import Loader2Icon from '@lucide/svelte/icons/loader-2'
import { cn } from '$lib/utils'
import type { SVGAttributes } from 'svelte/elements'

let {
  class: className,
  role = 'status',
  // we add name, color, and stroke for compatibility with different icon libraries props
  name,
  color,
  stroke,
  'aria-label': ariaLabel = 'Loading',
  ...restProps
}: SVGAttributes<SVGSVGElement> = $props()

// name/color/stroke 仅在明确提供时传入，null 与 undefined 均不能赋给 Lucide 的非可选属性。
// 使用 $derived 保持对响应式 props 的追踪，避免仅捕获初始值。
const optionalIconProps = $derived({
  ...(name !== null && name !== undefined ? { name } : {}),
  ...(color !== null && color !== undefined ? { color } : {}),
  ...(stroke !== null && stroke !== undefined ? { stroke } : {})
})
</script>

<Loader2Icon
  {role}
  {...optionalIconProps}
  aria-label={ariaLabel}
  class={cn('size-4 animate-spin', className)}
  {...restProps}
/>
