<script lang="ts">
import * as Dialog from '$lib/components/ui/dialog/index.js'
import { cn, type WithoutChildrenOrChild } from '$lib/utils'
import Command from './command.svelte'
import type { Command as CommandPrimitive, Dialog as DialogPrimitive } from 'bits-ui'
import type { Snippet } from 'svelte'

let {
  open = $bindable(false),
  ref = $bindable(null),
  value = $bindable(''),
  title = 'Command Palette',
  description = 'Search for a command to run...',
  showCloseButton = false,
  portalProps,
  children,
  class: className,
  ...restProps
}: WithoutChildrenOrChild<DialogPrimitive.RootProps> &
  WithoutChildrenOrChild<CommandPrimitive.RootProps> & {
    portalProps?: DialogPrimitive.PortalProps
    children: Snippet
    title?: string
    description?: string
    showCloseButton?: boolean
    class?: string
  } = $props()
</script>

<Dialog.Root bind:open {...restProps}>
  <Dialog.Header class="sr-only">
    <Dialog.Title>{title}</Dialog.Title>
    <Dialog.Description>{description}</Dialog.Description>
  </Dialog.Header>
  <!-- portalProps 仅在调用方显式传入时展开，undefined 不能赋给非可选属性。 -->
  <Dialog.Content
    class={cn('top-1/3 translate-y-0 overflow-hidden rounded-xl! p-0', className)}
    {showCloseButton}
    {...portalProps ? { portalProps } : {}}
  >
    <Command {...restProps} bind:value bind:ref {children} />
  </Dialog.Content>
</Dialog.Root>
