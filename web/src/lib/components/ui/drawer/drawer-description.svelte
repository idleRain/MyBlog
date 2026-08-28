<script lang="ts">
import { Drawer as DrawerPrimitive } from 'vaul-svelte'
import { cn } from '$lib/utils'

let {
  ref = $bindable(null),
  class: className,
  ...restProps
}: DrawerPrimitive.DescriptionProps & { ref?: HTMLDivElement | null } = $props()
</script>

<!-- 元素引用经 bind:ref 双向绑定，与 bits-ui 组件标准用法保持一致。 -->
<DrawerPrimitive.Description
  bind:ref
  data-slot="drawer-description"
  class={cn('text-sm text-muted-foreground', className)}
  {...restProps}
/>
