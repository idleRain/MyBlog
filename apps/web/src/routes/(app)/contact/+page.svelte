<script lang="ts">
import GithubIcon from '$lib/components/icons/github-icon.svelte'
import { SITE_NAME_ZH } from '@myblog/shared'
import { Mail } from '@lucide/svelte'

// 联系渠道集中声明，卡片区按渠道逐项渲染。
const contactChannels = [
  {
    name: 'GitHub',
    note: 'Issue 与讨论最适合技术话题，处理进度可见。',
    href: 'https://github.com/idleRain',
    external: true
  },
  {
    name: '邮箱',
    note: '合作与长文交流请来信，工作日通常两至三天内回复。',
    href: 'mailto:gold.experience@foxmail.com',
    external: false
  }
]
</script>

<svelte:head>
  <title>联系方式 - {SITE_NAME_ZH}</title>
  <meta name="description" content="与闲雨小筑站长取得联系的真实渠道：GitHub、邮箱与友链互换。" />
</svelte:head>

<section class="texture-grid relative min-h-screen overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      往来 · Contact
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-4xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      联系方式
    </h1>

    <!-- 往来引语：朱红竖线引出。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">
        欢迎写信交流，无论是勘误指正、技术讨论还是单纯打个招呼。
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
            {#if channel.name === 'GitHub'}
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
      <h2 class="font-display text-2xl font-black">站内通道</h2>
      <ul class="mt-6 space-y-3 text-sm leading-relaxed text-muted-foreground">
        <li class="border-l-2 border-signal/40 pl-4">
          评论留言：任意文章页下方均可留言，游客可评，站长会逐条查看。
        </li>
        <li class="border-l-2 border-signal/40 pl-4">
          友链互换：在页头「友情链接」弹窗底部提交申请，通过审核后即会展示。
        </li>
      </ul>
    </div>
  </div>

  <!-- 幽灵字：右下角的大号衬线单字，作为版面前景装饰，不参与交互。 -->
  <div
    class="pointer-events-none absolute right-8 bottom-14 hidden font-display text-[11rem] leading-none font-black text-foreground/[0.05] select-none lg:block"
    aria-hidden="true"
  >
    联
  </div>
</section>
