<script lang="ts">
import HeroSection from '$lib/components/layout/HeroSection.svelte'
import ContentSection from '$lib/components/layout/ContentSection.svelte'
import type { PageProps } from './$types'
import { onMount } from 'svelte'

let { data }: PageProps = $props()
let currentSection = $state(0)
let isScrolling = $state(false)
let sections: HTMLElement[] = []

onMount(() => {
  const handleWheel = (e: WheelEvent) => {
    if (isScrolling) return

    const scrollY = window.scrollY
    const windowHeight = window.innerHeight
    const direction = e.deltaY > 0 ? 1 : -1

    // 只在以下情况下阻止默认滚动并实现整屏滚动：
    // 1. 在Hero区域向下滚动
    // 2. 在Content区域顶部向上滚动回Hero区域
    const isInHeroArea = scrollY < windowHeight * 0.1
    const isAtContentTop = scrollY >= windowHeight * 0.9 && scrollY < windowHeight * 1.2

    if ((isInHeroArea && direction > 0) || (isAtContentTop && direction < 0)) {
      e.preventDefault()
      isScrolling = true

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

      // 防止连续滚动
      setTimeout(() => {
        isScrolling = false
      }, 800)
    }
    // 在Content区域内部正常滚动，不阻止默认行为
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
  window.addEventListener('scroll', handleScroll)

  return () => {
    window.removeEventListener('wheel', handleWheel)
    window.removeEventListener('scroll', handleScroll)
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
