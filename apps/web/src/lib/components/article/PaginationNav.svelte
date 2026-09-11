<script lang="ts">
// 分页导航属性：basePath 为不带查询串的页面路径。
interface Props {
  currentPage: number
  totalPages: number
  basePath: string
}

let { currentPage, totalPages, basePath }: Props = $props()
</script>

<nav
  class="mt-12 flex items-center justify-between border-t border-border pt-6 font-mono text-sm text-muted-foreground"
>
  {#if currentPage > 1}
    <a
      href={`${basePath}?page=${currentPage - 1}`}
      class="transition-colors duration-200 hover:text-signal"
    >
      ← 上一页
    </a>
  {:else}
    <span aria-hidden="true">← 上一页</span>
  {/if}
  <span>第 {currentPage} / {totalPages} 页</span>
  {#if currentPage < totalPages}
    <a
      href={`${basePath}?page=${currentPage + 1}`}
      class="transition-colors duration-200 hover:text-signal"
    >
      下一页 →
    </a>
  {:else}
    <span aria-hidden="true">下一页 →</span>
  {/if}
</nav>
