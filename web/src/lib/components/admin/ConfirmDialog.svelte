<script lang="ts">
import { AlertDialog } from '$ui'
import { Spinner } from '$ui/spinner'
import { cn } from '$lib/utils'

interface Props {
  /** 对话框开合状态，由调用方持有并在动作完成后置回 false。 */
  open?: boolean
  title: string
  description: string
  confirmText?: string | undefined
  cancelText?: string | undefined
  /** 危险操作（删除、禁用等）时确认按钮呈现 destructive 语义。 */
  destructive?: boolean | undefined
  /** 确认动作：执行完成后组件自动关闭对话框。 */
  onConfirm?: (() => void | Promise<void>) | undefined
}

let {
  open = $bindable(false),
  title,
  description,
  confirmText = '确认',
  cancelText = '取消',
  destructive = false,
  onConfirm
}: Props = $props()

let isConfirming = $state(false)

// 顶部状态线颜色随操作语义切换：普通动作用 signal，危险动作用 destructive。
const topLineColor = $derived(destructive ? 'bg-destructive' : 'bg-signal')

// 确认点击时阻止 AlertDialog 的默认关闭，等待动作完成后自行关闭，
// 以便在提交期间呈现按钮禁用与加载状态。
async function handleConfirm(event: MouseEvent) {
  event.preventDefault()
  if (isConfirming) return
  isConfirming = true
  try {
    await onConfirm?.()
    open = false
  } finally {
    isConfirming = false
  }
}
</script>

<!-- 危险确认对话框：替代原生 confirm，规格书直角加状态线样式。 -->
<AlertDialog.Root bind:open>
  <AlertDialog.Content class="rounded-none border border-border ring-0 sm:max-w-md">
    <!-- 状态线：直角内容顶部的语义色短线，对应规格书的页边标注线。 -->
    <div class={cn('h-0.5 w-10', topLineColor)} aria-hidden="true"></div>
    <AlertDialog.Header class="pt-2">
      <AlertDialog.Title>{title}</AlertDialog.Title>
      <AlertDialog.Description>{description}</AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <AlertDialog.Cancel variant="outline" class="rounded-none" disabled={isConfirming}>
        {cancelText}
      </AlertDialog.Cancel>
      <AlertDialog.Action
        variant={destructive ? 'destructive' : 'default'}
        class="rounded-none"
        disabled={isConfirming}
        onclick={handleConfirm}
      >
        {#if isConfirming}
          <Spinner data-icon="inline-start" />
        {/if}
        {confirmText}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
