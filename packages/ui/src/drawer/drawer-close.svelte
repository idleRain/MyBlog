<script lang="ts">
import { Drawer as DrawerPrimitive } from 'vaul-svelte'

let {
  ref = $bindable(null),
  ...restProps
}: DrawerPrimitive.CloseProps & { ref?: HTMLButtonElement | null } = $props()
</script>

<!-- 元素引用经 bind:ref 双向绑定，与 bits-ui 组件标准用法保持一致。 -->
<DrawerPrimitive.Close bind:ref data-slot="drawer-close" {...restProps} />
