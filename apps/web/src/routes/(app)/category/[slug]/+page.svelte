<script lang="ts">
import { PagePlaceholder } from '$lib/components/layout'
import { page } from '$app/stores'

// 读取路由参数中的分类 slug，仅用于展位标注，分类列表功能接入后整页替换。
const slug = $derived($page.params.slug)
</script>

<PagePlaceholder
  eyebrow={`分类 · ${slug}`}
  ghostMark="类"
  title="分类目录"
  description="该分类下的文章将在这里陈列，支持翻页与排序。"
/>
