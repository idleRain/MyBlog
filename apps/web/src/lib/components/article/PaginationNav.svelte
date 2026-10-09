<script lang="ts">
// 分页导航属性：basePath 为不带查询串的页面路径，query 为翻页时需保留的查询串。
interface Props {
  currentPage: number
  totalPages: number
  basePath: string
  query?: string
}

let { currentPage, totalPages, basePath, query = '' }: Props = $props()

// 翻页链接统一在此拼接，存在保留查询串时页码参数以 & 连接在其后。
function pageHref(page: number): string {
  return query ? `${basePath}?${query}&page=${page}` : `${basePath}?page=${page}`
}
</script>

<!--
  分页版式保留为前台自有实现，未收口到 $ui/pagination。
  原因：组件库的分页原语只渲染 button 元素。bits-ui 的 Page 与 PrevButton 支持经 child
  片段替换元素，但 $ui 包装层未透传该属性，前台无法在消费 $ui 的同时产出可被爬虫与
  无 JavaScript 环境跟随的真实链接。包装层补齐透传之前，此处保留锚点版式。
-->
<nav
  aria-label="分页导航"
  class="mt-12 flex items-center justify-between border-t border-border pt-6 font-mono text-sm text-muted-foreground"
>
  {#if currentPage > 1}
    <a
      href={pageHref(currentPage - 1)}
      rel="prev"
      class="py-2 transition-colors duration-200 hover:text-signal"
    >
      ← 上一页
    </a>
  {:else}
    <span aria-hidden="true" class="py-2">← 上一页</span>
  {/if}
  <span>第 {currentPage} / {totalPages} 页</span>
  {#if currentPage < totalPages}
    <a
      href={pageHref(currentPage + 1)}
      rel="next"
      class="py-2 transition-colors duration-200 hover:text-signal"
    >
      下一页 →
    </a>
  {:else}
    <span aria-hidden="true" class="py-2">下一页 →</span>
  {/if}
</nav>
