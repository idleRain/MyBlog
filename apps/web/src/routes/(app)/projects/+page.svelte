<script lang="ts">
import { m } from '$i18n'

// 幽灵字为版面前景装饰字形，不承载语义，不随语言切换，故不入词表。
const GHOST_MARK = '档'

// 项目档案仅收录真实可查证的作品，避免以虚构条目填充版面。
// 技术名为专有名词，不随语言切换。
const techStack = [
  'Go / Gin / GORM',
  'SvelteKit / Svelte 5',
  'TypeScript',
  'MySQL',
  'pnpm Monorepo'
]

// 标题与刊头共用同一条词条，站名取自词表。
const pageTitle = m['ui:meta.pageTitle']({
  title: m['projects:heading'](),
  site: m['ui:site.name']()
})
</script>

<svelte:head>
  <title>{pageTitle}</title>
  <meta name="description" content={m['projects:description']({ site: m['ui:site.name']() })} />
</svelte:head>

<section class="texture-grid relative min-h-screen overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      {m['projects:eyebrow']()}
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-4xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      {m['projects:heading']()}
    </h1>

    <!-- 档案定位引语：朱红竖线引出。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">
        {m['projects:lead']()}
      </p>
    </div>

    <!-- 项目档案卡：MyBlog 本站工程。 -->
    <article class="mt-16 border border-line bg-secondary">
      <div
        class="flex flex-wrap items-baseline justify-between gap-3 border-b border-line px-6 py-4"
      >
        <h2 class="font-display text-2xl font-black">
          {m['projects:card.title']({ site: m['ui:site.name']() })}
        </h2>
        <span class="font-mono text-xs text-muted-foreground">{m['projects:card.meta']()}</span>
      </div>

      <div class="grid gap-8 px-6 py-6 md:grid-cols-3">
        <div class="space-y-4 text-sm leading-relaxed text-muted-foreground md:col-span-2">
          <p>
            {m['projects:card.paragraphOne']()}
          </p>
          <p>
            {m['projects:card.paragraphTwo']()}
          </p>
          <a
            href="https://github.com/idleRain"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-2 text-sm font-bold text-signal underline-offset-4 hover:underline"
          >
            {m['projects:card.sourceLink']()}
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
            {m['projects:techStackTitle']()}
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
      <p class="font-display text-xl font-medium">{m['projects:moreTitle']()}</p>
      <p class="mt-2 text-sm text-muted-foreground">
        {m['projects:moreNote']()}
      </p>
    </div>
  </div>

  <!-- 幽灵字：右下角的大号衬线单字，作为版面前景装饰，不参与交互。 -->
  <div
    class="pointer-events-none absolute right-8 bottom-14 hidden font-display text-[11rem] leading-none font-black text-foreground/[0.05] select-none lg:block"
    aria-hidden="true"
  >
    {GHOST_MARK}
  </div>
</section>
