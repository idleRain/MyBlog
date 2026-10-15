<script lang="ts">
import { ArticleIndexList, PaginationNav } from '$lib/components/article'
import { FollowButton } from '$lib/components/user'
import type { PageProps } from './$types'
import { m } from '$i18n'

let { data }: PageProps = $props()

const profile = $derived(data.profile)

// 展示名优先取昵称，缺失时回退用户名。
const displayName = $derived(profile.nickname || profile.username)

// 标题与描述随语言切换，站名取自词表。
const pageTitle = $derived(m['author:title']({ name: displayName, site: m['ui:site.name']() }))
const pageDescription = $derived(profile.bio || m['author:description']({ name: displayName }))

// 总页数向上取整，无文章时保持 1 以呈现完整导航骨架。
const totalPages = $derived(Math.max(1, Math.ceil(data.list.total / data.list.pageSize)))
</script>

<svelte:head>
  <title>{pageTitle}</title>
  <meta name="description" content={pageDescription} />
</svelte:head>

<section class="texture-grid min-h-screen pt-16">
  <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 md:py-24">
    <!-- 作者名片：头像、展示名、简介与公开统计，头像为功能性圆形。 -->
    <header class="border-b border-line pb-8">
      <div class="flex flex-col items-start gap-6 sm:flex-row sm:items-center">
        <div
          class="flex h-20 w-20 shrink-0 items-center justify-center rounded-full bg-secondary font-display text-3xl font-black text-foreground"
          aria-hidden="true"
        >
          {displayName.charAt(0)}
        </div>
        <div class="min-w-0 flex-1">
          <p
            class="mb-3 flex items-center gap-3 text-sm font-bold tracking-[0.35em] text-signal uppercase"
          >
            {m['author:eyebrow']()}
            <span class="h-px w-10 bg-signal/60" aria-hidden="true"></span>
          </p>
          <h1 class="font-display text-3xl font-black">{displayName}</h1>
          {#if profile.bio}
            <p class="mt-3 max-w-prose text-sm leading-relaxed text-muted-foreground">
              {profile.bio}
            </p>
          {/if}
        </div>
        <FollowButton targetUserId={profile.id} />
      </div>

      <!-- 公开统计与站点链接行 -->
      <div
        class="mt-6 flex flex-wrap items-center gap-x-6 gap-y-2 font-mono text-xs text-muted-foreground"
      >
        <span>{m['author:articleCount']({ count: profile.articleCount })}</span>
        <span>{m['author:followerCount']({ count: profile.followerCount })}</span>
        <span>{m['author:followingCount']({ count: profile.followingCount })}</span>
        {#if profile.website}
          <a
            href={profile.website}
            target="_blank"
            rel="noopener noreferrer"
            class="text-signal underline-offset-4 hover:underline"
          >
            {profile.website}
          </a>
        {/if}
      </div>
    </header>

    <!-- 文章目录 -->
    <div class="mt-12">
      {#if data.list.articles.length === 0}
        <div class="border-l-2 border-signal py-4 pl-6">
          <p class="font-display text-xl font-medium">{m['author:emptyTitle']()}</p>
          <p class="mt-2 text-sm text-muted-foreground">{m['author:emptyNote']()}</p>
        </div>
      {:else}
        <ArticleIndexList
          articles={data.list.articles}
          offset={(data.currentPage - 1) * data.list.pageSize}
        />

        <PaginationNav
          currentPage={data.currentPage}
          {totalPages}
          basePath={`/author/${profile.id}`}
        />
      {/if}
    </div>
  </div>
</section>
