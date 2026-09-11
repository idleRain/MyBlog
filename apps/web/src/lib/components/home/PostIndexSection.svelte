<script lang="ts">
import { postEntries, hotPosts, tagCloud } from '$lib/data/home-content'
import { scrollReveal } from '$lib/motion/reveal'

// 归档页入口：真实路由，页面由文章归档接口驱动。
const ARCHIVE_LINK = '/archives'
const GITHUB_URL = 'https://github.com/idleRain'

const authorSocialLinks = [
  { name: 'GitHub', href: GITHUB_URL },
  { name: 'RSS', href: '#' },
  { name: '邮箱', href: '#' }
]

/** 订阅表单为版面演示用途，拦截提交以避免无效请求。 */
function handleSubscribe(event: SubmitEvent) {
  event.preventDefault()
}
</script>

<section id="articles" class="mx-auto max-w-6xl px-4 pb-20 sm:px-6">
  <div class="grid gap-12 lg:grid-cols-12">
    <!-- 文章目录列表：编号 + 日期 + 标题的索引行版式 -->
    <div class="lg:col-span-8">
      <div class="mb-8 flex items-end justify-between border-b border-line pb-4" use:scrollReveal>
        <h2 class="font-display text-3xl font-black">近期文章</h2>
        <span class="font-mono text-sm text-muted-foreground">
          共 {postEntries.length} 篇 · 持续更新
        </span>
      </div>

      <ol class="divide-y divide-line">
        {#each postEntries as post (post.id)}
          <li class="post-row group relative py-7 pl-4" use:scrollReveal={{ y: 26 }}>
            <span class="post-rule absolute top-6 left-0 h-10 w-0.5 bg-signal" aria-hidden="true"
            ></span>
            <div class="flex items-start justify-between gap-4">
              <div>
                <p class="mb-2 flex items-center gap-3 font-mono text-xs text-muted-foreground">
                  <span class="post-index font-display text-lg font-black">{post.index}</span>
                  {post.date} · 阅读 {post.readMinutes} 分钟
                </p>
                <h3 class="font-display text-xl leading-snug font-bold">
                  <a href="#featured" class="transition-colors duration-200 hover:text-signal">
                    {post.title}
                  </a>
                </h3>
                <p class="mt-2 max-w-prose text-sm leading-relaxed text-muted-foreground">
                  {post.excerpt}
                </p>
              </div>
              <span
                class="hidden shrink-0 text-signal transition-transform duration-200 ease-(--ease-out-strong) group-hover:translate-x-1 sm:block"
                aria-hidden="true"
              >
                →
              </span>
            </div>
          </li>
        {/each}
      </ol>

      <a
        href={ARCHIVE_LINK}
        class="group mt-8 inline-flex items-center gap-2 text-sm font-bold text-signal underline-offset-4 hover:underline"
        use:scrollReveal
      >
        查看全部文章
        <span
          class="transition-transform duration-200 ease-(--ease-out-strong) group-hover:translate-x-1"
          aria-hidden="true"
        >
          →
        </span>
      </a>
    </div>

    <!-- 侧边栏：作者、热门、标签与订阅，粘性跟随滚动 -->
    <aside class="lg:col-span-4">
      <div class="sticky top-24 flex flex-col gap-8">
        <div id="about" class="border border-line bg-secondary p-6" use:scrollReveal={{ x: 24 }}>
          <div class="flex items-center gap-4">
            <div
              class="flex h-16 w-16 items-center justify-center rounded-full bg-primary font-display text-2xl font-black text-primary-foreground"
              aria-hidden="true"
            >
              M
            </div>
            <div>
              <h3 class="font-display text-lg font-black">开发者</h3>
              <p class="text-sm text-muted-foreground">全栈工程师 · 编程与设计</p>
            </div>
          </div>
          <p class="mt-4 text-sm leading-relaxed text-muted-foreground">
            白天写代码，晚上写字。相信缓慢而扎实的积累，偏爱干净的结构与诚实的文字。
          </p>
          <div class="mt-4 flex gap-3 text-sm font-bold">
            {#each authorSocialLinks as link (link.name)}
              <a href={link.href} class="text-signal hover:underline">{link.name}</a>
            {/each}
          </div>
        </div>

        <div class="border border-line p-6" use:scrollReveal={{ x: 24, delay: 0.06 }}>
          <h3 class="mb-4 font-display text-lg font-black">热门文章</h3>
          <ol class="flex flex-col gap-4">
            {#each hotPosts as post (post.id)}
              <li>
                <a href="#featured" class="group block">
                  <span class="font-mono text-xs text-muted-foreground">
                    {post.date} · 阅读 {post.readMinutes} 分钟
                  </span>
                  <p
                    class="mt-1 text-sm leading-snug font-bold transition-colors duration-200 group-hover:text-signal"
                  >
                    {post.title}
                  </p>
                </a>
              </li>
            {/each}
          </ol>
        </div>

        <div class="border border-line p-6" use:scrollReveal={{ x: 24, delay: 0.12 }}>
          <h3 class="mb-4 font-display text-lg font-black">标签</h3>
          <div class="flex flex-wrap gap-2">
            {#each tagCloud as tag (tag.name)}
              <a
                href="#articles"
                class="border border-line bg-card px-3 py-1 text-xs font-medium transition-colors duration-150 hover:border-signal hover:text-signal"
              >
                {tag.name}
              </a>
            {/each}
          </div>
        </div>

        <div
          class="bg-primary p-6 text-primary-foreground"
          use:scrollReveal={{ x: 24, delay: 0.18 }}
        >
          <h3 class="font-display text-xl font-black">订阅手记</h3>
          <p class="mt-2 text-sm leading-relaxed text-primary-foreground/70">
            每月一封邮件，汇总当月文章与书单，不打扰。
          </p>
          <form class="mt-4 flex" onsubmit={handleSubscribe}>
            <label for="subscribe-email" class="sr-only">邮箱地址</label>
            <input
              id="subscribe-email"
              type="email"
              required
              placeholder="you@example.com"
              class="min-w-0 flex-1 border border-line bg-background px-3 py-2.5 text-sm text-foreground transition-colors duration-150 outline-none placeholder:text-muted-foreground focus:border-signal"
            />
            <button
              type="submit"
              class="shrink-0 bg-signal px-4 text-sm font-bold text-signal-foreground transition-[background-color,transform] duration-200 ease-(--ease-out-strong) hover:bg-signal/90 active:scale-[0.97]"
            >
              订阅
            </button>
          </form>
        </div>
      </div>
    </aside>
  </div>
</section>

<style>
/*
   * 目录行的索引反馈：悬停时编号由淡转朱红，
   * 左侧竖条自顶部展开，形成翻阅目录的手感。
   */
.post-row .post-index {
  transition: color 200ms var(--ease-out-strong);
  color: var(--border);
}

.post-row:hover .post-index {
  color: var(--signal);
}

.post-row .post-rule {
  transform: scaleY(0);
  transform-origin: top;
  transition: transform 240ms var(--ease-out-strong);
}

.post-row:hover .post-rule {
  transform: scaleY(1);
}

/* 减少动态偏好下取消竖条展开，保留编号变色以维持可读反馈。 */
@media (prefers-reduced-motion: reduce) {
  .post-row .post-rule {
    transition: none;
  }
}
</style>
