<script lang="ts">
import { Toaster } from '$ui/sonner/index.js'
import { ModeWatcher } from 'mode-watcher'
import '@/app.css'

let { children }: { children: import('svelte').Snippet } = $props()
</script>

<ModeWatcher />
<Toaster position="top-center" />

{@render children()}
