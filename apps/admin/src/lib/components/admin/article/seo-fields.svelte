<script lang="ts">
import { Label } from '$ui/label'
import { Input } from '$ui/input'

interface Props {
  seoTitle?: string
  seoDescription?: string
  seoKeywords?: string
}

let {
  seoTitle = $bindable(''),
  seoDescription = $bindable(''),
  seoKeywords = $bindable('')
}: Props = $props()
</script>

<div class="grid gap-4">
  <div class="space-y-2">
    <Label for="seo-title">SEO 标题</Label>
    <Input
      id="seo-title"
      bind:value={seoTitle}
      placeholder="自定义搜索引擎展示标题，留空使用文章标题"
      maxlength={100}
    />
  </div>

  <div class="space-y-2">
    <Label for="seo-description">SEO 描述</Label>
    <Input
      id="seo-description"
      bind:value={seoDescription}
      placeholder="用于搜索引擎摘要展示"
      maxlength={255}
    />
  </div>

  <div class="space-y-2">
    <Label for="seo-keywords">SEO 关键词</Label>
    <Input
      id="seo-keywords"
      bind:value={seoKeywords}
      placeholder="多个关键词用英文逗号分隔"
      maxlength={200}
    />
  </div>
</div>
