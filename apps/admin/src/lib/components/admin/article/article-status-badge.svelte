<script lang="ts">
import type { ArticleStatus } from '@myblog/api/modules/article/types'
import { ARTICLE_STATUS_CONFIG } from '$lib/constants/article'
import { Badge } from '$ui'

interface Props {
  status: ArticleStatus
}

let { status }: Props = $props()
</script>

<!-- 文章状态徽章，样式取自常量配置的语义变体 -->
<Badge variant={ARTICLE_STATUS_CONFIG[status].variant}>
  {ARTICLE_STATUS_CONFIG[status].label}
</Badge>
