<script lang="ts">
import { BackToTop, FeaturedStory, HeroSection, PostIndexSection } from '$lib/components/home'
import { ArchiveTimeline } from '$lib/components/article'
import { ReadingProgress } from '$lib/components/home'
import { SITE_NAME_ZH } from '@myblog/shared'
import type { PageProps } from './$types'

let { data }: PageProps = $props()

// 本期精选取热门文章榜首，热门侧栏展示其余热门文章。
const featured = $derived(data.popular[0] ?? null)
const hotPosts = $derived(data.popular.slice(1))
</script>

<svelte:head>
  <title>{`${SITE_NAME_ZH} - 把想法，编译成界面`}</title>
  <meta
    name="description"
    content="一个专注于技术分享和创意设计的个人博客，探索现代Web开发的无限可能。"
  />
  <meta name="keywords" content="博客,技术,开发,设计,SvelteKit,Go,全栈开发" />

  <!-- Open Graph -->
  <meta property="og:title" content={`${SITE_NAME_ZH} - 把想法，编译成界面`} />
  <meta property="og:description" content="一个专注于技术分享和创意设计的个人博客" />
  <meta property="og:type" content="website" />
  <meta property="og:url" content="https://myblog.example.com" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content={`${SITE_NAME_ZH} - 把想法，编译成界面`} />
  <meta name="twitter:description" content="一个专注于技术分享和创意设计的个人博客" />
</svelte:head>

<!-- 阅读进度条：视口顶端的朱红细线 -->
<ReadingProgress />

<!-- 首页正文：编辑杂志版式，叠加印刷网格纸纹 -->
<div class="texture-grid">
  <HeroSection />
  {#if featured}
    <FeaturedStory article={featured} />
  {/if}
  <PostIndexSection posts={data.recent} {hotPosts} tags={data.tags} />
  {#if data.archives.length > 0}
    <!-- 创作年轮：与归档页共用时间线组件，此处仅作节选预览。 -->
    <section class="mx-auto max-w-6xl px-4 pb-24 sm:px-6">
      <div class="mb-14 flex items-end justify-between border-b border-line pb-4">
        <h2 class="font-display text-3xl font-black">创作年轮</h2>
        <a
          href="/archives"
          class="font-mono text-sm text-muted-foreground transition-colors duration-200 hover:text-signal"
        >
          完整归档 →
        </a>
      </div>
      <ArchiveTimeline groups={data.archives} />
    </section>
  {/if}
</div>

<!-- 回到顶部按钮：越过阈值后浮现 -->
<BackToTop />
