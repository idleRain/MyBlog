<script lang="ts">
import {
  BackToTop,
  FeaturedStory,
  HeroSection,
  PostIndexSection,
  ReadingProgress,
  TimelineSection
} from '$lib/components/home/index.ts'
</script>

<svelte:head>
  <title>MyBlog - 把想法，编译成界面</title>
  <meta
    name="description"
    content="一个专注于技术分享和创意设计的个人博客，探索现代Web开发的无限可能。"
  />
  <meta name="keywords" content="博客,技术,开发,设计,SvelteKit,Go,全栈开发" />

  <!-- Open Graph -->
  <meta property="og:title" content="MyBlog - 把想法，编译成界面" />
  <meta property="og:description" content="一个专注于技术分享和创意设计的个人博客" />
  <meta property="og:type" content="website" />
  <meta property="og:url" content="https://myblog.example.com" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="MyBlog - 把想法，编译成界面" />
  <meta name="twitter:description" content="一个专注于技术分享和创意设计的个人博客" />
</svelte:head>

<!-- 阅读进度条：视口顶端的朱红细线 -->
<ReadingProgress />

<!-- 首页正文：编辑杂志版式，叠加印刷网格纸纹 -->
<div class="texture-grid">
  <HeroSection />
  <FeaturedStory />
  <PostIndexSection />
  <TimelineSection />
</div>

<!-- 回到顶部按钮：越过阈值后浮现 -->
<BackToTop />
