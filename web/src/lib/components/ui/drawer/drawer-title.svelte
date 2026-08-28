<script lang="ts">
import { Drawer as DrawerPrimitive } from 'vaul-svelte'
import { cn } from '$lib/utils'

let {
  ref = $bindable(null),
  class: className,
  ...restProps
}: DrawerPrimitive.TitleProps & { ref?: HTMLHeadingElement | null } = $props()
</script>

<!-- 元素引用经 bind:ref 双向绑定，与 bits-ui 组件标准用法保持一致。 -->
<DrawerPrimitive.Title
  bind:ref
  data-slot="drawer-title"
  class={cn('text-base font-medium text-foreground', className)}
  {...restProps}
/>
