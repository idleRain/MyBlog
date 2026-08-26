<script lang="ts">
import { Drawer as DrawerPrimitive } from 'vaul-svelte'

let {
  ref = $bindable(null),
  ...restProps
}: DrawerPrimitive.CloseProps & { ref?: HTMLButtonElement | null } = $props()
</script>

<!-- vaul-svelte 当前版本包装旧版 bits-ui，元素引用经 el 属性转发。 -->
<DrawerPrimitive.Close el={ref ?? undefined} data-slot="drawer-close" {...restProps} />
