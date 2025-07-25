<script lang="ts">
import { Button } from '$ui/button'
import { LogOut } from '@lucide/svelte'
import { performLogout } from '$lib/utils/jwt'
import { toast } from 'svelte-sonner'

export let variant: 'default' | 'outline' | 'secondary' | 'ghost' | 'link' | 'destructive' = 'ghost'
export let size: 'default' | 'sm' | 'lg' | 'icon' = 'default'
export let showIcon = true
export let showText = true
export let disabled = false

let isLoggingOut = false

async function handleLogout() {
  if (isLoggingOut) return

  isLoggingOut = true

  try {
    await performLogout()
    toast.success('已成功登出')
  } catch (error) {
    console.error('登出失败:', error)
    toast.error('登出失败，请重试')
  } finally {
    isLoggingOut = false
  }
}
</script>

<Button {variant} {size} {disabled} onclick={handleLogout} class="gap-2 {$$props.class || ''}">
  {#if isLoggingOut}
    <div
      class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
    ></div>
    {#if showText}
      登出中...
    {/if}
  {:else}
    {#if showIcon}
      <LogOut class="h-4 w-4" />
    {/if}
    {#if showText}
      登出
    {/if}
  {/if}
</Button>
