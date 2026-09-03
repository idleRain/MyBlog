<script lang="ts">
import { Button } from '$ui'

interface Props {
  page: number
  total: number
  pageSize: number
  onPageChange: (page: number) => void
}

// 由总数与每页条数推导总页数，后端分页响应不含 pages 字段。
let { page, total, pageSize, onPageChange }: Props = $props()

const totalPages = $derived(Math.max(1, Math.ceil(total / pageSize)))
</script>

<nav class="flex items-center justify-center gap-2" aria-label="分页">
  <Button
    variant="outline"
    size="sm"
    disabled={page <= 1}
    onclick={() => onPageChange(Math.max(1, page - 1))}
  >
    上一页
  </Button>
  <span class="px-2 text-sm text-muted-foreground">
    第 {page} 页，共 {totalPages} 页
  </span>
  <Button
    variant="outline"
    size="sm"
    disabled={page >= totalPages}
    onclick={() => onPageChange(Math.min(totalPages, page + 1))}
  >
    下一页
  </Button>
</nav>
