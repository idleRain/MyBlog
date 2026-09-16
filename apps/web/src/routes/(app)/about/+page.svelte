<script lang="ts">
import GithubIcon from '$lib/components/icons/github-icon.svelte'
import { SITE_NAME_ZH } from '@myblog/shared'

// 写作方向栏目按刊目形式陈列，新增方向时按行追加。
const writingColumns = [
  { name: '工程手记', note: '前端工程与后端架构的实践记录' },
  { name: '设计札记', note: '界面、排版与交互的观察笔记' },
  { name: '随想', note: '阅读与生活里值得留下的片段' }
]
</script>

<svelte:head>
  <title>关于本站 - {SITE_NAME_ZH}</title>
  <meta name="description" content="闲雨小筑的站点缘起、写作者介绍与写作方向。" />
</svelte:head>

<section class="texture-grid relative min-h-screen overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      刊物 · About
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-5xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      关于本站
    </h1>

    <!-- 站点定位引语：朱红竖线引出。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">
        白天写代码，晚上写字。相信缓慢而扎实的积累，偏爱干净的结构与诚实的文字。
      </p>
    </div>

    <!-- 刊物自述：两栏排版，右栏为刊物的由来。 -->
    <div class="mt-16 grid gap-12 md:grid-cols-3">
      <div class="md:col-span-2">
        <h2 class="font-display text-2xl font-black">这是一份什么样的刊物</h2>
        <div class="mt-6 space-y-4 text-sm leading-relaxed text-muted-foreground">
          <p>
            {SITE_NAME_ZH}
            是一份以手记形式更新的个人刊物，记录工程实践里的取舍、设计观察里的细节，以及阅读与生活里值得留下的片段。文章按刊目编排，可从博客目录或归档时间线翻阅。
          </p>
          <p>
            站点本身也是一份作品：以 Go 与 SvelteKit 构建的 Monorepo
            全栈工程，评论、收藏、通知与全文检索均为真实实现。若你对工程结构感兴趣，可在项目档案中查看它的完整档案。
          </p>
        </div>
      </div>
      <div>
        <h2 class="font-display text-2xl font-black">写作者</h2>
        <div class="mt-6 border border-line bg-secondary p-5">
          <div class="flex items-center gap-4">
            <div
              class="flex h-12 w-12 items-center justify-center bg-signal font-display text-lg font-black text-signal-foreground"
            >
              M
            </div>
            <div>
              <p class="font-semibold text-foreground">开发者</p>
              <p class="text-xs text-muted-foreground">全栈工程师 · 编程与设计</p>
            </div>
          </div>
          <p class="mt-4 text-sm leading-relaxed text-muted-foreground">
            关注工程结构与界面品质的交界地带，习惯把复杂问题拆成清晰的小块，再把小块装回优雅的整体。
          </p>
          <a
            href="https://github.com/idleRain"
            target="_blank"
            rel="noopener noreferrer"
            class="mt-4 inline-flex items-center gap-2 text-sm font-bold text-signal underline-offset-4 hover:underline"
          >
            <GithubIcon class="h-4 w-4" aria-hidden="true" />
            GitHub
          </a>
        </div>
      </div>
    </div>

    <!-- 写作方向：以刊目行形式陈列，行首以等宽序号标注。 -->
    <div class="mt-16">
      <h2 class="font-display text-2xl font-black">写作方向</h2>
      <ol class="mt-6 divide-y divide-line border-y border-line">
        {#each writingColumns as column, index (column.name)}
          <li class="flex items-baseline gap-4 py-4">
            <span class="font-mono text-sm text-muted-foreground">
              {String(index + 1).padStart(2, '0')}
            </span>
            <div>
              <p class="font-display text-lg font-bold">{column.name}</p>
              <p class="mt-1 text-sm text-muted-foreground">{column.note}</p>
            </div>
          </li>
        {/each}
      </ol>
    </div>
  </div>

  <!-- 幽灵字：右下角的大号衬线单字，作为版面前景装饰，不参与交互。 -->
  <div
    class="pointer-events-none absolute right-8 bottom-14 hidden font-display text-[11rem] leading-none font-black text-foreground/[0.05] select-none lg:block"
    aria-hidden="true"
  >
    介
  </div>
</section>
