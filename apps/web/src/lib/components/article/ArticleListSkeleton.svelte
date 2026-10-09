<script lang="ts">
import { Skeleton } from '$ui'
// 骨架行数量：与目录页每页数量保持同一版面节奏，避免加载完成后高度突变。
interface Props {
  rows?: number
}

let { rows = 4 }: Props = $props()

// 逐行错落的标题与摘要宽度，使骨架保留目录行的自然不齐感。
// 固定取值而非随机，避免服务端与客户端取值差异。
const TITLE_WIDTHS = ['w-3/5', 'w-4/5', 'w-2/3', 'w-1/2', 'w-3/4', 'w-5/6']
const SUMMARY_WIDTHS = ['w-full', 'w-11/12', 'w-10/12', 'w-9/12']

const rowIndexes = $derived(Array.from({ length: rows }, (_, index) => index))
</script>

<!-- 文章目录骨架：复用目录行的排版语汇，编号、元信息、标题与摘要各占其位。 -->
<ol class="divide-y divide-line" aria-hidden="true">
  {#each rowIndexes as index (index)}
    <li class="relative py-7 pl-4">
      <span class="absolute top-6 left-0 h-10 w-0.5 bg-signal/30"></span>
      <div class="flex items-start justify-between gap-4">
        <div class="min-w-0 flex-1">
          <div class="mb-2 flex items-center gap-3">
            <Skeleton.Skeleton class="h-5 w-8 rounded-none" />
            <Skeleton.Skeleton class="h-3 w-40 rounded-none" />
          </div>
          <Skeleton.Skeleton
            class={`mb-2 h-6 rounded-none ${TITLE_WIDTHS[index % TITLE_WIDTHS.length]}`}
          />
          <Skeleton.Skeleton
            class={`h-4 rounded-none ${SUMMARY_WIDTHS[index % SUMMARY_WIDTHS.length]}`}
          />
        </div>
        <Skeleton.Skeleton class="hidden h-4 w-6 shrink-0 rounded-none sm:block" />
      </div>
    </li>
  {/each}
</ol>
