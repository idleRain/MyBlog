<script lang="ts">
import { SITE_NAME_ZH } from '@myblog/shared'

// 收集项按数据类别陈列，新增采集场景时在此追加类别并在描述中说明启用时点。
const collectionItems = [
  {
    title: '文章浏览数据',
    items: [
      '访问文章时记录来源 IP 地址，用于按访客与自然日去重统计浏览量；站点当前未启用浏览器指纹等额外访客标识。',
      '浏览明细仅用于站内统计分析，聚合结果即文章页展示的浏览次数。'
    ]
  },
  {
    title: '站内搜索记录',
    items: [
      '使用目录页检索时记录搜索关键词、结果数量与检索耗时，用于检索质量分析与热词优化。',
      '搜索记录不与你的其他行为信息关联成画像。'
    ]
  },
  {
    title: '评论信息',
    items: [
      '游客留言需填写昵称与邮箱，昵称随评论公开展示，邮箱不公开展示。',
      '评论表预留的 IP 与浏览器标识字段当前未启用；如未来启用，会先在本页更新告知。'
    ]
  },
  {
    title: '注册账户资料',
    items: [
      '注册或登录账户时收集用户名、邮箱与密码，密码以 bcrypt 单向哈希存储，任何人（包括站长）无法看到原文。',
      '自助资料页中的昵称、头像、简介与网站地址仅用于站内展示。'
    ]
  }
]
</script>

<svelte:head>
  <title>隐私政策 - {SITE_NAME_ZH}</title>
  <meta
    name="description"
    content="闲雨小筑的隐私政策：说明本站收集哪些数据、用途、留存与脱敏策略。"
  />
</svelte:head>

<section class="texture-grid relative min-h-screen overflow-hidden pt-16">
  <div class="relative z-10 mx-auto w-full max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 栏目标注：朱红粗标签与发丝短线，与站内版面同构。 -->
    <p
      class="mb-6 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
    >
      合规 · Privacy
      <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
    </p>

    <!-- 刊头标题：衬线黑体大字，延续 Display 层级规格。 -->
    <h1 class="font-display text-5xl leading-[1.08] font-black tracking-tight sm:text-6xl">
      隐私政策
    </h1>

    <!-- 政策总纲引语：朱红竖线引出。 -->
    <div class="mt-10 max-w-xl border-l-2 border-signal pl-6">
      <p class="font-display text-xl leading-relaxed font-medium">
        本站以最小必要原则处理访客数据，只收集维持功能运转所必需的信息，并在本页公开说明。
      </p>
    </div>

    <!-- 收集项清单：按数据类别分节陈列。 -->
    <div class="mt-16 grid gap-12 md:grid-cols-3">
      <div class="md:col-span-2">
        <h2 class="font-display text-2xl font-black">本站收集哪些数据</h2>
        <div class="mt-8 space-y-10">
          {#each collectionItems as section (section.title)}
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
        <h3 class="font-display text-lg font-black">本站的承诺</h3>
        <ul class="mt-4 space-y-3 text-sm leading-relaxed text-muted-foreground">
          <li class="border-l-2 border-signal/40 pl-4">数据不向第三方出售，也不与第三方共享。</li>
          <li class="border-l-2 border-signal/40 pl-4">站点无广告脚本与跨站追踪器。</li>
          <li class="border-l-2 border-signal/40 pl-4">
            公开接口输出经脱敏处理，评论邮箱、来源 IP 等敏感字段不随公开接口返回。
          </li>
          <li class="border-l-2 border-signal/40 pl-4">
            浏览与搜索明细长期保存且仅用于统计分析；如引入自动清理周期，会在本页更新说明。
          </li>
          <li class="border-l-2 border-signal/40 pl-4">新增数据收集场景前，先更新本页告知。</li>
        </ul>
      </aside>
    </div>

    <!-- 更新说明：政策本身也是版面的一部分，标注生效口径。 -->
    <div class="mt-16 border-t border-line pt-6">
      <p class="text-sm leading-relaxed text-muted-foreground">
        本政策描述当前站点的实际数据处理方式；内容如有变更，会在本页直接更新并保留最新版本。对政策或数据处理的疑问，可经
        <a href="/contact" class="text-signal underline-offset-4 hover:underline">联系方式</a>
        页面中的渠道提出。
      </p>
    </div>
  </div>

  <!-- 幽灵字：右下角的大号衬线单字，作为版面前景装饰，不参与交互。 -->
  <div
    class="pointer-events-none absolute right-8 bottom-14 hidden font-display text-[11rem] leading-none font-black text-foreground/[0.05] select-none lg:block"
    aria-hidden="true"
  >
    隐
  </div>
</section>
