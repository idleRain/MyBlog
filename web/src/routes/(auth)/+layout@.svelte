<script>
import { Toaster } from '$ui/sonner'
import { ModeWatcher } from 'mode-watcher'
import '../../app.css'

let { children } = $props()
</script>

<ModeWatcher />
<Toaster position="top-center" />

<!-- 主内容区域 -->
<main class="min-h-screen">
  {@render children()}
</main>
