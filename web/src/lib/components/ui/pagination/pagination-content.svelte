<script lang="ts">
import type { HTMLAttributes } from 'svelte/elements'
import { cn, type WithElementRef } from '$lib/utils'

let {
  ref = $bindable(null),
  class: className,
  children,
  ...restProps
}: WithElementRef<HTMLAttributes<HTMLUListElement>> = $props()
</script>

<ul
  bind:this={ref}
  data-slot="pagination-content"
  class={cn('flex flex-row items-center gap-1', className)}
  {...restProps}
>
  {@render children?.()}
</ul>
