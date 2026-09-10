<script lang="ts">
import { gsap, MOTION } from '$lib/motion/gsap-setup'

let bar = $state<HTMLElement>()

$effect(() => {
  if (!bar) return

  // 阅读进度条与整页滚动挂钩，scrub 模式下以线性插值跟随滚动比例。
  const tween = gsap.to(bar, {
    scaleX: 1,
    ease: MOTION.ease.linear,
    scrollTrigger: { start: 0, end: 'max', scrub: 0.3 }
  })

  return () => {
    tween.scrollTrigger?.kill()
    tween.kill()
  }
})
</script>

<!-- 阅读进度条：固定在视口顶端的朱红细线，宽度随滚动比例延展 -->
<div
  bind:this={bar}
  class="fixed inset-x-0 top-0 z-[60] h-0.5 origin-left scale-x-0 bg-signal"
  aria-hidden="true"
></div>
