<script lang="ts">
import { Skeleton } from '$ui/skeleton'

// 两个表单区块的字段数：基础资料四项、修改密码三项，与真实表单一致。
// 末项按多行输入的高度占位，使加载完成后的版式高度不突变。
const FORM_SECTIONS = [
  { key: 'profile', fields: 4 },
  { key: 'password', fields: 3 }
] as const

// 生成字段占位的序号，用于判断末项是否为多行输入。
function fieldIndexes(count: number): number[] {
  return Array.from({ length: count }, (_, index) => index)
}
</script>

<!-- 资料表单骨架：与真实表单同为纸深面双栏卡片。 -->
<div class="grid gap-8 lg:grid-cols-2" aria-hidden="true">
  {#each FORM_SECTIONS as section (section.key)}
    <div class="border border-line bg-secondary p-6">
      <Skeleton class="h-6 w-24 rounded-none" />
      <Skeleton class="mt-2 h-3 w-48 rounded-none" />

      {#each fieldIndexes(section.fields) as index (index)}
        <div class="mt-4 first:mt-6">
          <Skeleton class="mb-2 h-3.5 w-16 rounded-none" />
          <Skeleton
            class={`w-full rounded-none ${index === section.fields - 1 ? 'h-20' : 'h-10'}`}
          />
        </div>
      {/each}

      <Skeleton class="mt-6 h-11 w-full rounded-none" />
    </div>
  {/each}
</div>
