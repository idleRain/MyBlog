<script lang="ts">
import { Button } from '$lib/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '$lib/components/ui/avatar'
import { ChevronDown } from '@lucide/svelte'
import { onMount } from 'svelte'

let particlesContainer: HTMLElement
let particles: Array<{
  x: number
  y: number
  vx: number
  vy: number
  size: number
  alpha: number
}> = []

onMount(() => {
  if (!particlesContainer) return

  // 初始化粒子 - 响应式数量调整
  const isMobile = window.innerWidth < 768
  const particleCount = isMobile ? 40 : 120 // 移动端减少粒子数量
  const particleSpeed = isMobile ? 0.2 : 0.4 // 移动端降低速度
  const particleSize = isMobile ? 2 : 4 // 移动端减小尺寸

  for (let i = 0; i < particleCount; i++) {
    particles.push({
      x: Math.random() * window.innerWidth,
      y: Math.random() * window.innerHeight,
      vx: (Math.random() - 0.5) * particleSpeed,
      vy: (Math.random() - 0.5) * particleSpeed,
      size: Math.random() * particleSize + 1,
      alpha: Math.random() * 0.6 + 0.2
    })
  }

  // 动画循环
  function animate() {
    particles.forEach(particle => {
      particle.x += particle.vx
      particle.y += particle.vy

      // 边界检测
      if (particle.x < 0 || particle.x > window.innerWidth) particle.vx *= -1
      if (particle.y < 0 || particle.y > window.innerHeight) particle.vy *= -1

      // 保持在边界内
      particle.x = Math.max(0, Math.min(window.innerWidth, particle.x))
      particle.y = Math.max(0, Math.min(window.innerHeight, particle.y))
    })

    // 更新粒子位置
    if (particlesContainer) {
      const canvas = particlesContainer.querySelector('canvas') as HTMLCanvasElement
      if (canvas) {
        const ctx = canvas.getContext('2d')
        if (ctx) {
          ctx.clearRect(0, 0, canvas.width, canvas.height)

          particles.forEach((particle, index) => {
            // 为粒子添加隐约的连线效果
            // 移动端减少连线效果以提升性能
            if (!isMobile) {
              particles.forEach((otherParticle, otherIndex) => {
                if (index !== otherIndex) {
                  const dx = particle.x - otherParticle.x
                  const dy = particle.y - otherParticle.y
                  const distance = Math.sqrt(dx * dx + dy * dy)

                  if (distance < 150) {
                    ctx.strokeStyle = `rgba(74, 108, 247, ${(1 - distance / 150) * 0.15})`
                    ctx.lineWidth = 1
                    ctx.beginPath()
                    ctx.moveTo(particle.x, particle.y)
                    ctx.lineTo(otherParticle.x, otherParticle.y)
                    ctx.stroke()
                  }
                }
              })
            }

            // 绘制粒子，加上发光效果
            const gradient = ctx.createRadialGradient(
              particle.x,
              particle.y,
              0,
              particle.x,
              particle.y,
              particle.size * 2
            )
            gradient.addColorStop(0, `rgba(74, 108, 247, ${particle.alpha})`)
            gradient.addColorStop(0.5, `rgba(147, 51, 234, ${particle.alpha * 0.7})`)
            gradient.addColorStop(1, `rgba(74, 108, 247, 0)`)

            ctx.fillStyle = gradient
            ctx.beginPath()
            ctx.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2)
            ctx.fill()
          })
        }
      }
    }

    requestAnimationFrame(animate)
  }

  // 创建canvas
  const canvas = document.createElement('canvas')
  canvas.width = window.innerWidth
  canvas.height = window.innerHeight
  canvas.className = 'absolute inset-0 pointer-events-none'
  particlesContainer.appendChild(canvas)

  animate()

  // 监听窗口大小变化
  const handleResize = () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
  }
  window.addEventListener('resize', handleResize)

  return () => {
    window.removeEventListener('resize', handleResize)
  }
})
</script>

<section class="relative flex min-h-screen items-center justify-center overflow-hidden">
  <!-- 动态粒子背景 -->
  <div bind:this={particlesContainer} class="absolute inset-0 opacity-80"></div>

  <!-- 渐变背景 -->
  <div
    class="absolute inset-0 bg-gradient-to-br from-blue-50/50 via-purple-50/30 to-pink-50/50 dark:from-gray-950 dark:via-blue-950/20 dark:to-purple-950/20"
  ></div>

  <!-- 内容区域 -->
  <div class="relative z-10 mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
    <div class="flex flex-col items-center justify-between gap-12 lg:flex-row">
      <!-- 左侧：个人头像 -->
      <div class="order-2 lg:order-1">
        <div class="group relative">
          <!-- 微光边框 -->
          <div
            class="absolute -inset-1 animate-pulse rounded-full bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 opacity-75 transition-opacity group-hover:opacity-100"
          ></div>
          <Avatar class="relative h-32 w-32 sm:h-48 sm:w-48">
            <AvatarImage src="/avatar.jpg" alt="头像" />
            <AvatarFallback
              class="bg-gradient-to-br from-blue-500 to-purple-600 text-3xl font-bold text-white sm:text-4xl"
            >
              M
            </AvatarFallback>
          </Avatar>
        </div>
      </div>

      <!-- 右侧：个人介绍 -->
      <div class="order-1 text-center lg:order-2 lg:text-left">
        <!-- 姓名标题 -->
        <h1
          class="mb-4 text-4xl font-bold text-gray-900 sm:text-5xl lg:text-6xl xl:text-7xl dark:text-white"
        >
          <span
            class="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent"
          >
            个人博客
          </span>
        </h1>

        <!-- 职业描述 -->
        <p class="mb-4 text-xl text-gray-600 sm:text-2xl dark:text-gray-300">
          前端/全栈开发工程师 · 创意设计师
        </p>

        <!-- 个性标语 -->
        <p class="mb-8 max-w-2xl text-lg text-gray-500 sm:text-xl dark:text-gray-400">
          用代码编织创意，用技术改变世界。<br class="hidden sm:block" />
          专注于构建优雅的数字体验。
        </p>

        <!-- 行动按钮 -->
        <div class="flex flex-col justify-center gap-4 sm:flex-row lg:justify-start">
          <Button
            size="lg"
            class="transform bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg transition-all duration-300 hover:-translate-y-1 hover:from-blue-700 hover:to-purple-700 hover:shadow-xl"
          >
            查看作品集
          </Button>
          <Button
            variant="outline"
            size="lg"
            class="border-gray-300 hover:bg-gray-50 dark:border-gray-600 dark:hover:bg-gray-800"
          >
            联系我
          </Button>
        </div>

        <!-- 统计数据 -->
        <div class="mt-12 flex flex-wrap justify-center gap-8 lg:justify-start">
          <div class="text-center">
            <div class="text-2xl font-bold text-gray-900 sm:text-3xl dark:text-white">20+</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">项目经验</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-gray-900 sm:text-3xl dark:text-white">5年+</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">开发经验</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-gray-900 sm:text-3xl dark:text-white">0</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">博客文章</div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- 底部滚动指示器 -->
  <div class="absolute bottom-8 left-1/2 -translate-x-1/2 transform">
    <div class="flex flex-col items-center space-y-2 text-gray-400 dark:text-gray-500">
      <span class="text-sm">向下滚动探索更多</span>
      <ChevronDown class="h-6 w-6 animate-bounce" />
    </div>
  </div>
</section>

<style>
/* 确保粒子动画流畅 */
section {
  will-change: transform;
}

/* 优化动画性能 */
@media (prefers-reduced-motion: reduce) {
  .animate-pulse,
  .animate-bounce {
    animation: none;
  }
}
</style>
