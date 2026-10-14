<script lang="ts">
import { m } from '$i18n'

// 幽灵字为版面前景装饰字形，不承载语义，不随语言切换，故不入词表。
const GHOST_MARK = '隐'

// 收集项按数据类别陈列，新增采集场景时在此追加类别并在描述中说明启用时点。
// id 为列表键值，与语言无关；文案一律取自词表。
const collectionItems = [
  {
    id: 'browsing',
    title: m['privacy:collection.browsing.title'](),
    items: [m['privacy:collection.browsing.item1'](), m['privacy:collection.browsing.item2']()]
  },
  {
    id: 'search',
    title: m['privacy:collection.search.title'](),
    items: [m['privacy:collection.search.item1'](), m['privacy:collection.search.item2']()]
  },
  {
    id: 'comment',
    title: m['privacy:collection.comment.title'](),
    items: [m['privacy:collection.comment.item1'](), m['privacy:collection.comment.item2']()]
  },
  {
    id: 'account',
    title: m['privacy:collection.account.title'](),
    items: [m['privacy:collection.account.item1'](), m['privacy:collection.account.item2']()]
  },
  {
    id: 'session',
    title: m['privacy:collection.session.title'](),
    items: [
      m['privacy:collection.session.item1'](),
      m['privacy:collection.session.item2'](),
      m['privacy:collection.session.item3']()
    ]
  }
]

// 标题与刊头共用同一条词条，站名取自词表。
const pageTitle = m['ui:meta.pageTitle']({
  title: m['privacy:heading'](),
  site: m['ui:site.name']()
})
</script>

<svelte:head>
  <title>{pageTitle}</title>
  <meta name="description" content={m['privacy:description']({ site: m['ui:site.name']() })} />
</svelte:head>

<section class="texture-grid relative min-h-screen overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      {m['privacy:eyebrow']()}
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-4xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      {m['privacy:heading']()}
    </h1>

    <!-- 政策总纲引语：朱红竖线引出。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">
        {m['privacy:lead']()}
      </p>
    </div>

    <!-- 收集项清单：按数据类别分节陈列。 -->
    <div class="mt-16 grid gap-12 md:grid-cols-3">
      <div class="md:col-span-2">
        <h2 class="font-display text-2xl font-black">{m['privacy:collectionTitle']()}</h2>
        <div class="mt-8 space-y-10">
          {#each collectionItems as section (section.id)}
            <section>
              <h3 class="font-display text-lg font-bold">{section.title}</h3>
              <ul class="mt-3 space-y-2 text-sm leading-relaxed text-muted-foreground">
                {#each section.items as item (item)}
                  <li class="border-l-2 border-signal/40 pl-4">{item}</li>
                {/each}
              </ul>
            </section>
          {/each}
        </div>
      </div>

      <!-- 侧栏承诺：与正文分栏对应，逐条列出边界承诺。 -->
      <aside class="h-fit border border-line bg-secondary p-6">
        <h3 class="font-display text-lg font-black">{m['privacy:promiseTitle']()}</h3>
        <ul class="mt-4 space-y-3 text-sm leading-relaxed text-muted-foreground">
          <li class="border-l-2 border-signal/40 pl-4">{m['privacy:promise1']()}</li>
          <li class="border-l-2 border-signal/40 pl-4">{m['privacy:promise2']()}</li>
          <li class="border-l-2 border-signal/40 pl-4">
            {m['privacy:promise3']()}
          </li>
          <li class="border-l-2 border-signal/40 pl-4">
            {m['privacy:promise4']()}
          </li>
          <li class="border-l-2 border-signal/40 pl-4">{m['privacy:promise5']()}</li>
        </ul>
      </aside>
    </div>

    <!-- 更新说明：政策本身也是版面的一部分，标注生效口径。 -->
    <div class="mt-16 border-t border-line pt-6">
      <p class="text-sm leading-relaxed text-muted-foreground">
        {m['privacy:updateBefore']()}
        <a href="/contact" class="text-signal underline-offset-4 hover:underline">
          {m['contact:heading']()}
        </a>
        {m['privacy:updateAfter']()}
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
