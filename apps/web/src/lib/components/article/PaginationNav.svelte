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

<nav
  class="mt-12 flex items-center justify-between border-t border-border pt-6 font-mono text-sm text-muted-foreground"
>
  {#if currentPage > 1}
    <a href={pageHref(currentPage - 1)} class="transition-colors duration-200 hover:text-signal">
      ← 上一页
    </a>
  {:else}
    <span aria-hidden="true">← 上一页</span>
  {/if}
  <span>第 {currentPage} / {totalPages} 页</span>
  {#if currentPage < totalPages}
    <a href={pageHref(currentPage + 1)} class="transition-colors duration-200 hover:text-signal">
      下一页 →
    </a>
  {:else}
    <span aria-hidden="true">下一页 →</span>
  {/if}
</nav>
