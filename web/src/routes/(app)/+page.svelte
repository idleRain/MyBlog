<script lang="ts">
import HeroSection from '$lib/components/layout/HeroSection.svelte'
import ContentSection from '$lib/components/layout/ContentSection.svelte'
import type { PageProps } from './$types'
import { onMount } from 'svelte'

let { data }: PageProps = $props()
let currentSection = $state(0)
let isScrolling = $state(false)

onMount(() => {
  let wheelTimeout: number | undefined
  let accumulatedDelta = 0
  const SCROLL_THRESHOLD = 50 // 降低触发阈值，使其更敏感

  const handleWheel = (e: WheelEvent) => {
    if (isScrolling) return

    const scrollY = window.scrollY
    const windowHeight = window.innerHeight
    const direction = e.deltaY > 0 ? 1 : -1

    // 累积滚动量，使小幅滚动也能触发
    accumulatedDelta += Math.abs(e.deltaY)

    // 清除之前的超时，重新设置
    if (wheelTimeout) {
      clearTimeout(wheelTimeout)
    }

    // 更精确的区域判断
    const isInHeroArea = scrollY < windowHeight * 0.2 // Hero区域范围
    const isAtContentTop = scrollY >= windowHeight * 1.0 && scrollY < windowHeight * 1.2 // 只有刚进入Content区域的很小范围

    // 判断是否应该触发整屏滚动
    const shouldTriggerSnap =
      (isInHeroArea && direction > 0) || // Hero区域向下滚动到Content
      (isAtContentTop && direction < 0) // 只有在Content区域顶部很小范围内向上滚动才回到Hero

    if (shouldTriggerSnap && accumulatedDelta > SCROLL_THRESHOLD) {
      e.preventDefault()
      isScrolling = true
      accumulatedDelta = 0 // 重置累积值

      if (direction > 0) {
        // 向下滚动到Content区域
        currentSection = 1
        const targetSection = document.getElementById('content-section')
        if (targetSection) {
          targetSection.scrollIntoView({ behavior: 'smooth' })
        }
      } else {
        // 向上滚动到Hero区域
        currentSection = 0
        window.scrollTo({ top: 0, behavior: 'smooth' })
      }

      // 防止连续滚动，缩短等待时间
      setTimeout(() => {
        isScrolling = false
      }, 600)
    } else {
      // 设置超时来重置累积值
      wheelTimeout = window.setTimeout(() => {
        accumulatedDelta = 0
      }, 150)
    }
  }

  // 优化的键盘导航支持
  const handleKeyDown = (e: KeyboardEvent) => {
    if (isScrolling) return

    const scrollY = window.scrollY
    const windowHeight = window.innerHeight

    if (e.key === 'ArrowDown' || e.key === 'PageDown') {
      if (scrollY < windowHeight * 0.5) {
        e.preventDefault()
        isScrolling = true
        currentSection = 1
        const targetSection = document.getElementById('content-section')
        if (targetSection) {
          targetSection.scrollIntoView({ behavior: 'smooth' })
        }
        setTimeout(() => {
          isScrolling = false
        }, 600)
      }
    } else if (e.key === 'ArrowUp' || e.key === 'PageUp') {
      if (scrollY > windowHeight * 0.5) {
        e.preventDefault()
        isScrolling = true
        currentSection = 0
        window.scrollTo({ top: 0, behavior: 'smooth' })
        setTimeout(() => {
          isScrolling = false
        }, 600)
      }
    }
  }

  // 监听滚动事件来更新当前区域
  const handleScroll = () => {
    const scrollY = window.scrollY
    const windowHeight = window.innerHeight

    if (scrollY < windowHeight / 2) {
      currentSection = 0
    } else {
      currentSection = 1
    }
  }

  window.addEventListener('wheel', handleWheel, { passive: false })
  window.addEventListener('keydown', handleKeyDown)
  window.addEventListener('scroll', handleScroll)

  return () => {
    window.removeEventListener('wheel', handleWheel)
    window.removeEventListener('keydown', handleKeyDown)
    window.removeEventListener('scroll', handleScroll)
    if (wheelTimeout) {
      clearTimeout(wheelTimeout)
    }
  }
})
</script>

<svelte:head>
  <title>MyBlog - 用代码编织创意</title>
  <meta
    name="description"
    content="一个专注于技术分享和创意设计的个人博客，探索现代Web开发的无限可能。"
  />
  <meta name="keywords" content="博客,技术,开发,设计,SvelteKit,Go,全栈开发" />

  <!-- Open Graph -->
  <meta property="og:title" content="MyBlog - 用代码编织创意" />
  <meta property="og:description" content="一个专注于技术分享和创意设计的个人博客" />
  <meta property="og:type" content="website" />
  <meta property="og:url" content="https://myblog.example.com" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="MyBlog - 用代码编织创意" />
  <meta name="twitter:description" content="一个专注于技术分享和创意设计的个人博客" />
</svelte:head>

<!-- 首屏英雄区域 -->
<div id="hero-section" class="relative">
  <HeroSection />
</div>

<!-- 内容推荐区域 -->
<div id="content-section" class="relative">
  <ContentSection />
</div>
