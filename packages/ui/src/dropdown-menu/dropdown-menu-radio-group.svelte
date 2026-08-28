<script lang="ts">
import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui'

let {
  ref = $bindable(null),
  value = $bindable(),
  ...restProps
}: DropdownMenuPrimitive.RadioGroupProps = $props()
</script>

<!-- value 在未选中时为 undefined，与判别联合的 bindable 约束冲突，使用 never 收束。 -->
<DropdownMenuPrimitive.RadioGroup
  bind:ref
  bind:value={value as never}
  data-slot="dropdown-menu-radio-group"
  {...restProps}
/>
