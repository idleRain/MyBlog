<script lang="ts">
import * as AlertDialog from '$ui/alert-dialog'
import { Button } from '$ui'

interface Props {
  title: string
  description?: string
  confirmText?: string
  cancelText?: string
  destructive?: boolean
  isLoading?: boolean
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}

let {
  title,
  description,
  confirmText = '确认',
  cancelText = '取消',
  destructive = false,
  isLoading = false,
  open,
  onOpenChange,
  onConfirm
}: Props = $props()
</script>

<AlertDialog.Root {open} {onOpenChange}>
  <AlertDialog.Content class="sm:max-w-md">
    <AlertDialog.Header>
      <AlertDialog.Title>{title}</AlertDialog.Title>
      {#if description}
        <AlertDialog.Description>{description}</AlertDialog.Description>
      {/if}
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button variant="outline" onclick={() => onOpenChange(false)} disabled={isLoading}>
        {cancelText}
      </Button>
      <Button
        variant={destructive ? 'destructive' : 'default'}
        onclick={onConfirm}
        disabled={isLoading}
      >
        {#if isLoading}
          <span
            class="mr-2 inline-block size-3.5 animate-spin rounded-full border-2 border-current border-t-transparent"
          ></span>
        {/if}
        {confirmText}
      </Button>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
