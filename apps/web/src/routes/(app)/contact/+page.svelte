<script lang="ts">
import GithubIcon from '$lib/components/icons/github-icon.svelte'
import { Mail } from '@lucide/svelte'
import { m } from '$i18n'

// 站外渠道固定地址集中声明，避免散落的魔法字符串。
const GITHUB_URL = 'https://github.com/idleRain'
const CONTACT_EMAIL_HREF = 'mailto:gold.experience@foxmail.com'

// 幽灵字为版面前景装饰字形，不承载语义，不随语言切换，故不入词表。
const GHOST_MARK = '联'

// 联系渠道集中声明，卡片区按渠道逐项渲染；图标按 icon 标识分支，不依赖已翻译的渠道名。
const contactChannels = [
  {
    icon: 'github',
    name: m['ui:brand.github'](),
    note: m['contact:channels.githubNote'](),
    href: GITHUB_URL,
    external: true
  },
  {
    icon: 'mail',
    name: m['ui:label.email'](),
    note: m['contact:channels.emailNote'](),
    href: CONTACT_EMAIL_HREF,
    external: false
  }
]

// 标题与刊头共用同一条词条，站名取自词表。
const pageTitle = m['ui:meta.pageTitle']({
  title: m['contact:heading'](),
  site: m['ui:site.name']()
})
</script>

<svelte:head>
  <title>{pageTitle}</title>
  <meta name="description" content={m['contact:description']({ site: m['ui:site.name']() })} />
</svelte:head>

<section class="texture-grid relative min-h-screen overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      {m['contact:eyebrow']()}
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-4xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      {m['contact:heading']()}
    </h1>

    <!-- 往来引语：朱红竖线引出。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">
        {m['contact:lead']()}
      </p>
    </div>

    <!-- 联系渠道：卡片逐项陈列，悬停以 signal 色反馈。 -->
    <div class="mt-16 grid gap-6 md:grid-cols-2">
      {#each contactChannels as channel (channel.name)}
        <a
          href={channel.href}
          target={channel.external ? '_blank' : undefined}
          rel={channel.external ? 'noopener noreferrer' : undefined}
          class="group border border-line bg-secondary p-6 transition-colors duration-200 hover:border-signal"
        >
          <div class="flex items-center gap-3">
            {#if channel.icon === 'github'}
              <GithubIcon class="h-5 w-5 text-muted-foreground" aria-hidden="true" />
            {:else}
              <Mail class="h-5 w-5 text-muted-foreground" aria-hidden="true" />
            {/if}
            <h2 class="font-display text-xl font-black">{channel.name}</h2>
          </div>
          <p class="mt-3 text-sm leading-relaxed text-muted-foreground">{channel.note}</p>
          <p class="mt-4 font-mono text-xs text-signal underline-offset-4 group-hover:underline">
            {channel.href}
          </p>
        </a>
      {/each}
    </div>

    <!-- 站内通道：说明无需写信也能完成的往来事项。 -->
    <div class="mt-12">
      <h2 class="font-display text-2xl font-black">{m['contact:inSiteTitle']()}</h2>
      <ul class="mt-6 space-y-3 text-sm leading-relaxed text-muted-foreground">
        <li class="border-l-2 border-signal/40 pl-4">
          {m['contact:inSiteComments']()}
        </li>
        <li class="border-l-2 border-signal/40 pl-4">
          {m['contact:inSiteLinks']()}
        </li>
      </ul>
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
