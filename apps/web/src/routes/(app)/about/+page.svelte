<script lang="ts">
import GithubIcon from '$lib/components/icons/github-icon.svelte'
import { m } from '$i18n'

// 幽灵字为版面前景装饰字形，不承载语义，不随语言切换，故不入词表。
const GHOST_MARK = '介'

// 写作方向栏目按刊目形式陈列，新增方向时按行追加；id 为列表键值，与语言无关。
const writingColumns = [
  {
    id: 'engineering',
    name: m['about:columns.engineeringName'](),
    note: m['about:columns.engineeringNote']()
  },
  {
    id: 'design',
    name: m['about:columns.designName'](),
    note: m['about:columns.designNote']()
  },
  {
    id: 'notes',
    name: m['about:columns.notesName'](),
    note: m['about:columns.notesNote']()
  }
]

// 标题与刊头共用同一条词条，站名取自词表。
const pageTitle = m['ui:meta.pageTitle']({ title: m['about:heading'](), site: m['ui:site.name']() })
</script>

<svelte:head>
  <title>{pageTitle}</title>
  <meta name="description" content={m['about:description']({ site: m['ui:site.name']() })} />
</svelte:head>

<section class="texture-grid relative min-h-screen overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      {m['about:eyebrow']()}
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-4xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      {m['about:heading']()}
    </h1>

    <!-- 站点定位引语：朱红竖线引出。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">
        {m['ui:author.note']()}
      </p>
    </div>

    <!-- 刊物自述：两栏排版，右栏为刊物的由来。 -->
    <div class="mt-16 grid gap-12 md:grid-cols-3">
      <div class="md:col-span-2">
        <h2 class="font-display text-2xl font-black">{m['about:issueTitle']()}</h2>
        <div class="mt-6 space-y-4 text-sm leading-relaxed text-muted-foreground">
          <p>
            {m['about:issueParagraphOne']({ site: m['ui:site.name']() })}
          </p>
          <p>
            {m['about:issueParagraphTwo']()}
          </p>
        </div>
      </div>
      <div>
        <h2 class="font-display text-2xl font-black">{m['about:authorTitle']()}</h2>
        <div class="mt-6 border border-line bg-secondary p-5">
          <div class="flex items-center gap-4">
            <div
              class="flex h-12 w-12 items-center justify-center bg-signal font-display text-lg font-black text-signal-foreground"
            >
              M
            </div>
            <div>
              <p class="font-semibold text-foreground">{m['ui:author.name']()}</p>
              <p class="text-xs text-muted-foreground">{m['ui:author.roleWithFields']()}</p>
            </div>
          </div>
          <p class="mt-4 text-sm leading-relaxed text-muted-foreground">
            {m['ui:author.approach']()}
          </p>
          <a
            href="https://github.com/idleRain"
            target="_blank"
            rel="noopener noreferrer"
            class="mt-4 inline-flex items-center gap-2 text-sm font-bold text-signal underline-offset-4 hover:underline"
          >
            <GithubIcon class="h-4 w-4" aria-hidden="true" />
            {m['ui:brand.github']()}
          </a>
        </div>
      </div>
    </div>

    <!-- 写作方向：以刊目行形式陈列，行首以等宽序号标注。 -->
    <div class="mt-16">
      <h2 class="font-display text-2xl font-black">{m['about:columnsTitle']()}</h2>
      <ol class="mt-6 divide-y divide-line border-y border-line">
        {#each writingColumns as column, index (column.id)}
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
    {GHOST_MARK}
  </div>
</section>
