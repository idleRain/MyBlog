<script lang="ts">
import { formatDate } from '$lib/utils/format-date'
import { SITE_NAME_ZH } from '@myblog/shared'
import type { PageProps } from './$types'
import { ArticleAPI } from '$lib/api'

let { data }: PageProps = $props()

const article = $derived(data.article)

// 展示名优先取作者昵称，缺失时回退用户名。
const authorName = $derived(article.author.nickname || article.author.username)

// 浏览上报仅在浏览器执行，切换文章时按文章 ID 去重上报。
let reportedArticleId = 0
$effect(() => {
  if (reportedArticleId === article.id) return
  reportedArticleId = article.id
  void ArticleAPI.view(article.id)
})
</script>

<svelte:head>
  <title>{article.seoTitle || article.title} - {SITE_NAME_ZH}</title>
  {#if article.seoDescription || article.summary}
    <meta name="description" content={article.seoDescription || article.summary} />
  {/if}
  {#if article.seoKeywords}
    <meta name="keywords" content={article.seoKeywords} />
  {/if}
</svelte:head>

<article class="texture-grid min-h-screen pt-16">
  <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <div class="mx-auto max-w-3xl">
      <!-- 版面标题区：栏目标注、标题与元信息行。 -->
      <header>
        <p
          class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
        >
          手记 · 文章
          <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
        </p>
        <h1 class="font-display text-4xl leading-[1.15] font-black tracking-tight sm:text-5xl">
          {article.title}
        </h1>
        <p
          class="mt-6 flex flex-wrap items-center gap-x-4 gap-y-1 font-mono text-xs text-muted-foreground"
        >
          <span>{authorName}</span>
          <span>{formatDate(article.publishedAt ?? article.createdAt)}</span>
          <span>阅读 {article.readingTime} 分钟</span>
          <span>{article.viewCount} 次浏览</span>
        </p>
      </header>

      {#if article.coverImage}
        <img
          src={article.coverImage}
          alt={article.title}
          class="mt-10 w-full border border-line object-cover"
        />
      {/if}

      <!-- 正文：服务端渲染的 HTML 缓存，排版由 .article-body 作用域控制。 -->
      <div class="article-body mt-12">
        <!-- 内容由服务端 goldmark 渲染生成，Unsafe 关闭时原始脚本标签不进入输出，可安全挂载。 -->
        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
        {@html article.contentHtml}
      </div>

      {#if article.tags.length > 0}
        <div class="mt-12 flex flex-wrap gap-2">
          {#each article.tags as tag (tag.id)}
            <span class="border border-line bg-card px-3 py-1 text-xs font-medium">
              {tag.name}
            </span>
          {/each}
        </div>
      {/if}

      <div class="mt-16 border-t border-border pt-6">
        <a
          href="/blog"
          class="group inline-flex items-center gap-2 text-sm font-bold text-signal underline-offset-4 hover:underline"
        >
          ← 返回目录
        </a>
      </div>
    </div>
  </div>
</article>

<style>
/* 正文排版：标题衬线黑体、引用朱红竖线、代码块纸深面，与站点视觉语言一致。 */
.article-body :global(h1),
.article-body :global(h2),
.article-body :global(h3),
.article-body :global(h4) {
  margin: 2.4em 0 0.8em;
  font-weight: 900;
  line-height: 1.3;
  font-family: var(--font-display);
  letter-spacing: -0.01em;
}

.article-body :global(h1) {
  font-size: 1.75rem;
}

.article-body :global(h2) {
  font-size: 1.5rem;
}

.article-body :global(h3) {
  font-size: 1.25rem;
}

.article-body :global(h4) {
  font-size: 1.1rem;
}

.article-body :global(p) {
  margin: 1.1em 0;
  font-size: 1.0625rem;
  line-height: 1.9;
}

.article-body :global(a) {
  color: var(--signal);
  text-decoration: underline;
  text-underline-offset: 4px;
}

.article-body :global(blockquote) {
  margin: 1.6em 0;
  border-left: 2px solid var(--signal);
  padding-left: 1.25rem;
  color: var(--muted-foreground);
}

.article-body :global(ul),
.article-body :global(ol) {
  margin: 1.1em 0;
  padding-left: 1.5rem;
  font-size: 1.0625rem;
  line-height: 1.9;
}

.article-body :global(ul) {
  list-style: disc;
}

.article-body :global(ol) {
  list-style: decimal;
}

.article-body :global(li) {
  margin: 0.4em 0;
}

.article-body :global(code) {
  border: 1px solid var(--border);
  background-color: var(--secondary);
  padding: 0.1em 0.4em;
  font-size: 0.875em;
  font-family: var(--font-mono);
}

.article-body :global(pre) {
  margin: 1.6em 0;
  border: 1px solid var(--border);
  background-color: var(--secondary);
  padding: 1.1rem 1.25rem;
  overflow-x: auto;
}

.article-body :global(pre code) {
  border: none;
  background-color: transparent;
  padding: 0;
}

.article-body :global(img) {
  border: 1px solid var(--border);
  max-width: 100%;
}

.article-body :global(hr) {
  margin: 2.4em 0;
  border: none;
  border-top: 1px solid var(--border);
}

.article-body :global(table) {
  margin: 1.6em 0;
  border-collapse: collapse;
  width: 100%;
  font-size: 0.9375rem;
}

.article-body :global(th),
.article-body :global(td) {
  border: 1px solid var(--border);
  padding: 0.5rem 0.75rem;
  text-align: left;
}
</style>
