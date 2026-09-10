<script lang="ts">
import { gsap, ScrollTrigger, MOTION } from '$lib/motion/gsap-setup'
import { ArrowUp } from '@lucide/svelte'

// 滚动越过该阈值后显示回到顶部按钮，单位像素。
const SHOW_THRESHOLD = 600

let button = $state<HTMLElement>()

/** 平滑滚动回页面顶部，由按钮点击触发。 */
function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

$effect(() => {
  const element = button
  if (!element) return

  gsap.set(element, { autoAlpha: 0, y: 12 })

  // 出现与隐藏使用非对称时长：进入更从容，退出更干脆。
  const trigger = ScrollTrigger.create({
    start: SHOW_THRESHOLD,
    end: 'max',
    onToggle: self => {
      gsap.to(element, {
        autoAlpha: self.isActive ? 1 : 0,
        y: self.isActive ? 0 : 12,
        duration: self.isActive ? 0.3 : 0.2,
        ease: self.isActive ? MOTION.ease.out : 'power2.in'
      })
    }
  })

  return () => trigger.kill()
})
</script>

<button
  bind:this={button}
  type="button"
  class="fixed right-6 bottom-6 z-40 flex h-11 w-11 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-lg transition-[background-color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal hover:text-signal-foreground active:scale-[0.95]"
  aria-label="回到顶部"
  onclick={scrollToTop}
>
  <ArrowUp class="h-5 w-5" aria-hidden="true" />
</button>
