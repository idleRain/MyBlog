<script lang="ts">
import { Command as CommandPrimitive, useId } from 'bits-ui'
import { cn } from '$lib/utils'

let {
  ref = $bindable(null),
  class: className,
  children,
  heading,
  value,
  ...restProps
}: CommandPrimitive.GroupProps & {
  heading?: string
} = $props()
</script>

<CommandPrimitive.Group
  bind:ref
  data-slot="command-group"
  class={cn(
    'overflow-hidden p-1 text-foreground **:[[cmdk-group-heading]]:px-2 **:[[cmdk-group-heading]]:py-1.5 **:[[cmdk-group-heading]]:text-xs **:[[cmdk-group-heading]]:font-medium **:[[cmdk-group-heading]]:text-muted-foreground',
    className
  )}
  value={value ?? heading ?? `----${useId()}`}
  {...restProps}
>
  {#if heading}
    <CommandPrimitive.GroupHeading class="px-2 py-1.5 text-xs font-medium text-muted-foreground">
      {heading}
    </CommandPrimitive.GroupHeading>
  {/if}
  <!-- children 仅在调用方显式传入时展开，undefined 不能赋给非可选属性。 -->
  <CommandPrimitive.GroupItems {...children ? { children } : {}} />
</CommandPrimitive.Group>
