<script lang="ts">
import type { Article } from '@myblog/api/modules/article/types'
import { formatDate } from '$lib/utils/format-date'

// 目录行列表属性：articles 为当页文章，offset 用于跨页连续编号。
interface Props {
  articles: Article[]
  offset?: number
}

let { articles, offset = 0 }: Props = $props()

// 目录编号固定两位展示，文章总量超过 99 篇后按实际位数展开。
const indexLabel = (position: number) => String(offset + position + 1).padStart(2, '0')
</script>

<ol class="divide-y divide-line">
  {#each articles as article, position (article.id)}
    <li class="post-row group relative py-7 pl-4">
      <span class="post-rule absolute top-6 left-0 h-10 w-0.5 bg-signal" aria-hidden="true"></span>
      <div class="flex items-start justify-between gap-4">
        <div>
          <p class="mb-2 flex items-center gap-3 font-mono text-xs text-muted-foreground">
            <span class="post-index font-display text-lg font-black">{indexLabel(position)}</span>
            {formatDate(article.publishedAt ?? article.createdAt)} · 阅读 {article.readingTime} 分钟
          </p>
          <h2 class="font-display text-xl leading-snug font-bold">
            <a
              href={`/blog/${article.slug}`}
              class="transition-colors duration-200 hover:text-signal"
            >
              {article.title}
            </a>
          </h2>
          <p class="mt-2 max-w-prose text-sm leading-relaxed text-muted-foreground">
            {article.summary}
          </p>
        </div>
        <span
          class="hidden shrink-0 text-signal transition-transform duration-200 ease-(--ease-out-strong) group-hover:translate-x-1 sm:block"
          aria-hidden="true"
        >
          →
        </span>
      </div>
    </li>
  {/each}
</ol>

<style>
/*
   * 目录行的索引反馈：悬停时编号由淡转朱红，
   * 左侧竖条自顶部展开，形成翻阅目录的手感。
   */
.post-row .post-index {
  transition: color 200ms var(--ease-out-strong);
  color: var(--border);
}

.post-row:hover .post-index {
  color: var(--signal);
}

.post-row .post-rule {
  transform: scaleY(0);
  transform-origin: top;
  transition: transform 240ms var(--ease-out-strong);
}

.post-row:hover .post-rule {
  transform: scaleY(1);
}

/* 减少动态偏好下取消竖条展开，保留编号变色以维持可读反馈。 */
@media (prefers-reduced-motion: reduce) {
  .post-row .post-rule {
    transition: none;
  }
}
</style>
