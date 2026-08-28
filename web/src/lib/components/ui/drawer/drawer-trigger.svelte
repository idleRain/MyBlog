<script lang="ts">
import { Drawer as DrawerPrimitive } from 'vaul-svelte'

let {
  ref = $bindable(null),
  ...restProps
}: DrawerPrimitive.TriggerProps & { ref?: HTMLButtonElement | null } = $props()
</script>

<!-- 元素引用经 bind:ref 双向绑定，与 bits-ui 组件标准用法保持一致。 -->
<DrawerPrimitive.Trigger bind:ref data-slot="drawer-trigger" {...restProps} />
