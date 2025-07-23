<script lang="ts">
import { Button } from '$ui'
import Header from '$lib/components/layout/Header.svelte'
import { ModeWatcher } from 'mode-watcher'
import { onMount } from 'svelte'
import astronaut404 from '$lib/assets/images/404.svg'
import '../app.css'

let { status }: { error: App.Error; status: number } = $props()

// 根据状态码决定显示内容
let is500Error = $state(false)
let errorTitle = $state('')
let errorSubtitle = $state('')

// 响应式更新状态
$effect(() => {
  is500Error = status === 500
  errorTitle = is500Error ? '实验出现意外结果' : '探索进入未知领域'
  errorSubtitle = is500Error ? '我们的服务器正在经历技术性阵痛' : '你寻找的页面已消失在数字宇宙中'
})

// 粒子系统
let particles: Array<{ x: number; y: number; vx: number; vy: number; opacity: number }> = []
let canvas: HTMLCanvasElement
let ctx: CanvasRenderingContext2D | null

onMount(() => {
  if (canvas && typeof window !== 'undefined') {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
    ctx = canvas.getContext('2d')
    initParticles()
    animate()

    // 窗口大小改变时重新初始化
    const handleResize = () => {
      if (canvas) {
        canvas.width = window.innerWidth
        canvas.height = window.innerHeight
        initParticles()
      }
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }

  return () => {}
})

function initParticles() {
  particles = []
  for (let i = 0; i < 50; i++) {
    particles.push({
      x: Math.random() * window.innerWidth,
      y: Math.random() * window.innerHeight,
      vx: (Math.random() - 0.5) * 0.5,
      vy: (Math.random() - 0.5) * 0.5,
      opacity: Math.random() * 0.5 + 0.2
    })
  }
}

function animate() {
  if (!ctx || !canvas) return

  ctx.clearRect(0, 0, canvas.width, canvas.height)
  ctx.fillStyle = 'white'

  particles.forEach(particle => {
    particle.x += particle.vx
    particle.y += particle.vy

    // 边界检测
    if (particle.x < 0 || particle.x > canvas.width) particle.vx *= -1
    if (particle.y < 0 || particle.y > canvas.height) particle.vy *= -1

    ctx!.globalAlpha = particle.opacity
    ctx!.fillRect(particle.x, particle.y, 2, 2)
  })

  requestAnimationFrame(animate)
}

function createConstellation() {
  // 彩蛋：点击背景生成星座图案
  particles = particles.map(p => ({
    ...p,
    x: Math.random() * window.innerWidth,
    y: Math.random() * window.innerHeight,
    opacity: Math.random() * 0.8 + 0.2
  }))
}

// 最近访问页面（模拟数据）
const recentPages = [
  { title: '博客首页', url: '/', icon: '🏠' },
  { title: '关于我', url: '/about', icon: '👨‍💻' },
  { title: '项目展示', url: '/projects', icon: '🚀' }
]
</script>

<svelte:head>
  <title>{status} - MyBlog</title>
</svelte:head>

<!-- 主题监听器 -->
<ModeWatcher />

<!-- 保留导航栏 -->
<Header />

<div class="error-container relative min-h-screen overflow-hidden">
  <!-- 深空背景 -->
  <div class="absolute inset-0 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900"></div>

  <!-- 粒子画布 -->
  <canvas
    bind:this={canvas}
    class="pointer-events-none absolute inset-0"
    width={typeof window !== 'undefined' ? window.innerWidth : 1920}
    height={typeof window !== 'undefined' ? window.innerHeight : 1080}
    onclick={createConstellation}
  ></canvas>

  <!-- 主内容区 -->
  <div class="relative z-10 flex min-h-screen items-center justify-center px-4 pt-20">
    <div class="max-w-2xl text-center">
      <!-- 发光404 -->
      <div class="mb-8">
        <h1 class="glow-text animate-pulse text-9xl font-bold text-blue-400">
          {status}
        </h1>
      </div>

      <!-- 宇航员插画 -->
      <div class="relative mb-8">
        <div class="astronaut-container inline-block">
          <!-- 404宇航员SVG -->
          <img src={astronaut404} alt="404宇航员" class="animate-float mx-auto h-48 w-48" />

          <!-- 断裂的绳子 -->
          <div class="animate-swing absolute top-16 -right-8">
            <svg class="h-24 w-16" viewBox="0 0 64 96" fill="none">
              <path
                d="M32 0 Q36 8 32 16 Q28 24 32 32 Q36 40 32 48 Q28 56 32 64 Q36 72 32 80 Q28 88 32 96"
                stroke="#9ca3af"
                stroke-width="2"
                fill="none"
              />
            </svg>
            <div class="absolute -bottom-2 left-1/2 -translate-x-1/2 transform">
              <a href="/" class="text-sm text-blue-400 hover:text-blue-300">🏠 首页</a>
            </div>
          </div>
        </div>
      </div>

      <!-- 文案 -->
      <div class="mb-8 space-y-4">
        <h2 class="text-3xl font-bold text-white">{errorTitle}</h2>
        <p class="text-lg text-gray-300">{errorSubtitle}</p>
        {#if !is500Error}
          <p class="text-sm text-gray-400">带我回家 →</p>
        {:else}
          <p class="text-sm text-gray-400">维修团队正在紧急处理中...</p>
        {/if}
      </div>

      <!-- 主要操作按钮 -->
      <div class="mb-12">
        {#if !is500Error}
          <Button
            href="/"
            class="glow-button rounded-lg bg-blue-600 px-8 py-4 text-lg font-semibold text-white transition-all duration-300 hover:bg-blue-500 hover:shadow-lg hover:shadow-blue-500/25"
          >
            返回安全基地
          </Button>
        {:else}
          <div class="flex flex-col gap-4 sm:flex-row">
            <Button
              onclick={() => window.location.reload()}
              class="glow-button rounded-lg bg-blue-600 px-6 py-3 text-white transition-all duration-300 hover:bg-blue-500 hover:shadow-lg"
            >
              重试实验
            </Button>
            <Button
              href="/"
              variant="outline"
              class="rounded-lg border-white/20 px-6 py-3 text-white transition-all duration-300 hover:bg-white/10"
            >
              返回首页
            </Button>
          </div>
        {/if}
      </div>

      <!-- 星图导航 -->
      <div class="space-y-4">
        <p class="text-sm text-gray-400">或探索这些星球</p>
        <div class="flex justify-center space-x-6">
          {#each recentPages as planet, index (index)}
            <a
              href={planet.url}
              class="group flex flex-col items-center rounded-lg p-3 transition-colors hover:bg-white/5"
            >
              <span class="mb-1 text-2xl transition-transform group-hover:scale-110"
                >{planet.icon}</span
              >
              <span class="text-xs text-gray-400 group-hover:text-gray-300">{planet.title}</span>
            </a>
          {/each}
        </div>
      </div>
    </div>
  </div>
</div>

<style>
.glow-text {
  text-shadow:
    0 0 10px #4a6cf7,
    0 0 20px #4a6cf7,
    0 0 40px #4a6cf7;
}

.glow-button:hover {
  box-shadow:
    0 0 20px rgba(74, 108, 247, 0.4),
    0 4px 20px rgba(0, 0, 0, 0.3);
}

@keyframes float {
  0%,
  100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes swing {
  0%,
  100% {
    transform: rotate(-5deg);
  }
  50% {
    transform: rotate(5deg);
  }
}

.animate-float {
  animation: float 3s ease-in-out infinite;
}

.animate-swing {
  transform-origin: top center;
  animation: swing 2s ease-in-out infinite;
}

.error-container {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
}
</style>
