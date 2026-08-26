<script lang="ts">
import { Drawer as DrawerPrimitive } from 'vaul-svelte'
import { cn } from '$lib/utils'

let {
  ref = $bindable(null),
  class: className,
  ...restProps
}: DrawerPrimitive.TitleProps & { ref?: HTMLHeadingElement | null } = $props()
</script>

<!-- vaul-svelte 当前版本包装旧版 bits-ui，元素引用经 el 属性转发。 -->
<DrawerPrimitive.Title
  el={ref ?? undefined}
  data-slot="drawer-title"
  class={cn('text-base font-medium text-foreground', className)}
  {...restProps}
/>
