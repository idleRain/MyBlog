<script>
import { ModeWatcher } from 'mode-watcher'
import { Toaster } from '$ui/sonner/index.js'
import '@/app.css'

let { children } = $props()
</script>

<ModeWatcher />
<Toaster position="top-center" />

{@render children()}
