<script lang="ts">
import { SITE_NAME_ZH } from '@myblog/shared'

// 项目档案仅收录真实可查证的作品，避免以虚构条目填充版面。
const techStack = [
  'Go / Gin / GORM',
  'SvelteKit / Svelte 5',
  'TypeScript',
  'MySQL',
  'pnpm Monorepo'
]
</script>

<svelte:head>
  <title>项目档案 - {SITE_NAME_ZH}</title>
  <meta name="description" content="个人项目与开源作品的档案陈列，当前收录闲雨小筑全栈工程。" />
</svelte:head>

<section class="texture-grid relative min-h-screen overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      档案 · Projects
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-4xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      项目档案
    </h1>

    <!-- 档案定位引语：朱红竖线引出。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">
        这里陈列真实可查证的作品，每份档案随项目演进同步更新。
      </p>
    </div>

    <!-- 项目档案卡：MyBlog 本站工程。 -->
    <article class="mt-16 border border-line bg-secondary">
      <div
        class="flex flex-wrap items-baseline justify-between gap-3 border-b border-line px-6 py-4"
      >
        <h2 class="font-display text-2xl font-black">MyBlog · 闲雨小筑</h2>
        <span class="font-mono text-xs text-muted-foreground">2026 年起 · 持续迭代</span>
      </div>

      <div class="grid gap-8 px-6 py-6 md:grid-cols-3">
        <div class="space-y-4 text-sm leading-relaxed text-muted-foreground md:col-span-2">
          <p>
            你正在浏览的站点，本身就是档案里的第一件作品：一套 Go 与 SvelteKit
            构成的全栈个人博客工程，覆盖文章、评论、收藏、关注、通知、统计与友链等十一个业务模块，
            支持角色权限控制与中文全文检索。
          </p>
          <p>
            工程采用 pnpm Monorepo 组织：前端拆分为 api、http、auth、shared、ui 五个共享包，后端按
            handler、service、repository
            三层分层；评论审核状态机、通知生产链路与备份恢复脚本均为真实实现，而非演示样例。
          </p>
          <a
            href="https://github.com/idleRain"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-2 text-sm font-bold text-signal underline-offset-4 hover:underline"
          >
            查看源码仓库
            <span
              class="transition-transform duration-200 ease-(--ease-out-strong)"
              aria-hidden="true"
            >
              →
            </span>
          </a>
        </div>

        <div>
          <h3 class="text-xs font-bold tracking-[0.2em] text-muted-foreground uppercase">
            技术构成
          </h3>
          <ul class="mt-4 space-y-2 font-mono text-xs text-muted-foreground">
            {#each techStack as item (item)}
              <li class="border-l border-signal/40 pl-3">{item}</li>
            {/each}
          </ul>
        </div>
      </div>
    </article>

    <!-- 更多档案：以诚实空态说明收录节奏，不以占位条目填充。 -->
    <div class="mt-12 border-l-2 border-signal py-4 pl-6">
      <p class="font-display text-xl font-medium">更多档案整理中，随刊出陆续上架。</p>
      <p class="mt-2 text-sm text-muted-foreground">
        新的作品会先在博客目录以手记形式亮相，再归入此处。
      </p>
    </div>
  </div>

  <!-- 幽灵字：右下角的大号衬线单字，作为版面前景装饰，不参与交互。 -->
  <div
    class="pointer-events-none absolute right-8 bottom-14 hidden font-display text-[11rem] leading-none font-black text-foreground/[0.05] select-none lg:block"
    aria-hidden="true"
  >
    档
  </div>
</section>
