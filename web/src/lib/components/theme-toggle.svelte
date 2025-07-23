<script lang="ts">
import { Sun, Moon } from '@lucide/svelte'
import { toggleMode } from 'mode-watcher'
import { Button } from '$ui/button'
import { cn } from '$lib/utils'

const props = $props<{ class?: string }>()

function handleToggle() {
  console.log('Theme toggle clicked')
  toggleMode()
  console.log('Theme toggled')
}
</script>

<Button
  variant="outline"
  size="icon"
  onclick={handleToggle}
  class={cn(
    props.class,
    'fixed top-4 right-4 z-50 border bg-background/80 shadow-md backdrop-blur-sm transition-all hover:shadow-lg'
  )}
  title="切换主题"
>
  <Sun
    class="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-all dark:scale-0 dark:-rotate-90"
  />
  <Moon
    class="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-all dark:scale-100 dark:rotate-0"
  />
  <span class="sr-only">切换主题</span>
</Button>
