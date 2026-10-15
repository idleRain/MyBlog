<script lang="ts">
import { BackToTop, FeaturedStory, HeroSection, PostIndexSection } from '$lib/components/home'
import { ArchiveTimeline } from '$lib/components/article'
import { ReadingProgress } from '$lib/components/home'
import type { PageProps } from './$types'
import { m } from '$i18n'

let { data }: PageProps = $props()

// 标题与短描述在多处 head 标签复用，站名取自词表以随语言切换。
const pageTitle = m['home:seo.title']({ site: m['ui:site.name']() })
const pageDescriptionShort = m['home:seo.descriptionShort']()

// 本期精选取热门文章榜首，热门侧栏展示其余热门文章。
const featured = $derived(data.popular[0] ?? null)
const hotPosts = $derived(data.popular.slice(1))
</script>

<svelte:head>
  <title>{pageTitle}</title>
  <meta name="description" content={m['home:seo.description']()} />
  <meta name="keywords" content={m['home:seo.keywords']()} />

  <!-- Open Graph -->
  <meta property="og:title" content={pageTitle} />
  <meta property="og:description" content={pageDescriptionShort} />
  <meta property="og:type" content="website" />
  <meta property="og:url" content="https://myblog.example.com" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content={pageTitle} />
  <meta name="twitter:description" content={pageDescriptionShort} />
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
        <h2 class="font-display text-3xl font-black">{m['home:archivePreview.title']()}</h2>
        <a
          href="/archives"
          class="font-mono text-sm text-muted-foreground transition-colors duration-200 hover:text-signal"
        >
          {m['home:archivePreview.viewAll']()}
        </a>
      </div>
      <ArchiveTimeline groups={data.archives} />
    </section>
  {/if}
</div>

<!-- 回到顶部按钮：越过阈值后浮现 -->
<BackToTop />
