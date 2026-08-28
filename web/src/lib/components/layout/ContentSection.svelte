<script lang="ts">
import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card'
import { Calendar, Clock, Eye, Heart } from '@lucide/svelte'
import { Badge } from '$lib/components/ui/badge'

// 精选博客的展示数据，属于静态占位内容，接入文章接口后由真实数据替换。
const featuredPosts = [
  {
    id: 1,
    title: '现代Web开发的最佳实践',
    excerpt: '探索最新的前端技术栈，从SvelteKit到TailwindCSS的完整工作流程...',
    category: '技术',
    coverImage: 'https://images.unsplash.com/photo-1461749280684-dccba630e2f6?w=600&h=400&fit=crop',
    date: '2024-01-20',
    readTime: '8 分钟',
    views: 1240,
    likes: 45,
    tags: ['SvelteKit', 'TailwindCSS', '前端']
  },
  {
    id: 2,
    title: 'Go语言微服务架构设计',
    excerpt: '深入解析Go生态中的微服务最佳实践，从设计模式到部署策略...',
    category: '技术',
    coverImage: 'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?w=600&h=400&fit=crop',
    date: '2024-01-18',
    readTime: '12 分钟',
    views: 890,
    likes: 32,
    tags: ['Golang', '微服务', '架构']
  },
  {
    id: 3,
    title: '设计系统的构建之道',
    excerpt: '如何从零开始构建一套完整的设计系统，提升团队协作效率...',
    category: '设计',
    coverImage: 'https://images.unsplash.com/photo-1558655146-9f40138edfeb?w=600&h=400&fit=crop',
    date: '2024-01-15',
    readTime: '6 分钟',
    views: 650,
    likes: 28,
    tags: ['设计系统', 'UI/UX', '协作']
  }
]

// 日常分享的展示数据，条目以两位等宽序号编号，呼应规格书的条目标注语言。
const dailyShares = [
  { id: 1, index: '01', title: '晨间咖啡思考', description: '今天想到的一些关于代码优雅性的思考' },
  { id: 2, index: '02', title: '读书笔记', description: '《Clean Architecture》的核心要点总结' },
  { id: 3, index: '03', title: '编码音乐', description: '分享一些适合写代码时听的音乐playlist' },
  { id: 4, index: '04', title: '技术成长', description: '学习新技术的方法论和心得体会' },
  { id: 5, index: '05', title: '设计灵感', description: '在Dribbble上发现的有趣设计案例' },
  { id: 6, index: '06', title: '工具推荐', description: '提升开发效率的神器工具分享' },
  { id: 7, index: '07', title: '创意想法', description: '一些有趣的项目想法和实现思路' },
  { id: 8, index: '08', title: '生活健康', description: '程序员如何保持健康的工作生活平衡' },
  { id: 9, index: '09', title: '目标规划', description: '2024年的技术学习和职业发展规划' }
]

// 分类徽章统一使用 signal 强调色的弱背景，全站仅保留一种强调色。
const CATEGORY_BADGE_CLASS = 'rounded-none bg-signal/10 text-signal'
</script>

<section class="bg-muted/40 py-20">
  <div class="mx-auto max-w-7xl px-6 sm:px-10 lg:px-20">
    <!-- 板块标题的公共结构：两侧发丝线夹居中标题，形成规格书的章节样式。 -->
    {#snippet sectionHeading(title: string)}
      <div class="mb-12 flex items-center justify-center">
        <div class="h-px flex-grow bg-border" aria-hidden="true"></div>
        <h2 class="px-6 text-3xl font-bold text-foreground sm:text-4xl">{title}</h2>
        <div class="h-px flex-grow bg-border" aria-hidden="true"></div>
      </div>
    {/snippet}

    <!-- 精选博客板块 -->
    <div class="mb-20">
      {@render sectionHeading('深度思考')}

      <!-- 博客卡片网格 -->
      <div class="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3">
        {#each featuredPosts as post (post.id)}
          <Card
            class="group cursor-pointer rounded-none border-border bg-card pt-0 ring-0 transition-colors duration-200 hover:border-signal"
          >
            <!-- 封面图片 -->
            <div class="relative aspect-video overflow-hidden">
              <img
                src={post.coverImage}
                alt={post.title}
                class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105 motion-reduce:transition-none"
              />
              <!-- 分类标签 -->
              <div class="absolute top-4 left-4">
                <Badge class={CATEGORY_BADGE_CLASS}>
                  {post.category}
                </Badge>
              </div>
              <!-- 悬停遮罩：以 signal 弱色覆盖提示整卡可点击。 -->
              <div
                class="absolute inset-0 bg-signal/0 transition-colors duration-300 group-hover:bg-signal/10"
                aria-hidden="true"
              ></div>
            </div>

            <CardHeader class="pb-3">
              <CardTitle
                class="line-clamp-2 text-lg font-semibold text-foreground transition-colors group-hover:text-signal"
              >
                {post.title}
              </CardTitle>
            </CardHeader>

            <CardContent class="pt-0">
              <!-- 摘要 -->
              <p class="mb-4 line-clamp-2 text-sm text-muted-foreground">
                {post.excerpt}
              </p>

              <!-- 元信息 -->
              <div class="mb-3 flex items-center justify-between text-xs text-muted-foreground">
                <div class="flex items-center space-x-3">
                  <div class="flex items-center space-x-1">
                    <Calendar class="h-4 w-4" />
                    <span>{post.date}</span>
                  </div>
                  <div class="flex items-center space-x-1">
                    <Clock class="h-4 w-4" />
                    <span>{post.readTime}</span>
                  </div>
                </div>
                <div class="flex items-center space-x-3">
                  <div class="flex items-center space-x-1">
                    <Eye class="h-4 w-4" />
                    <span>{post.views}</span>
                  </div>
                  <div class="flex items-center space-x-1">
                    <Heart class="h-4 w-4" />
                    <span>{post.likes}</span>
                  </div>
                </div>
              </div>

              <!-- 标签 -->
              <div class="flex flex-wrap gap-1">
                {#each post.tags as tag (tag)}
                  <Badge variant="secondary" class="rounded-none text-xs">
                    {tag}
                  </Badge>
                {/each}
              </div>
            </CardContent>
          </Card>
        {/each}
      </div>
    </div>

    <!-- 日常分享板块 -->
    <div>
      {@render sectionHeading('生活切片')}

      <!-- 日常分享采用多列瀑布流布局，移动端保持两列以提升信息密度。 -->
      <div class="columns-2 gap-3 space-y-3 sm:gap-6 sm:space-y-6 lg:columns-3">
        {#each dailyShares as share (share.id)}
          <div class="break-inside-avoid">
            <Card
              class="group cursor-pointer rounded-none border-border bg-card ring-0 transition-colors duration-200 hover:border-signal"
            >
              <CardContent
                class="relative flex h-24 flex-col justify-start overflow-hidden p-3 sm:h-32 sm:p-6"
              >
                <!-- 序号与标题整体上移，为底部描述腾出空间。 -->
                <div
                  class="transition-transform duration-300 group-hover:-translate-y-2 motion-reduce:transition-none"
                >
                  <!-- 等宽序号延续规格书的条目标注语言。 -->
                  <div
                    class="mb-1.5 font-mono text-sm font-medium tracking-[0.18em] text-signal sm:mb-2"
                  >
                    {share.index}
                  </div>

                  <!-- 标题 -->
                  <h3
                    class="text-xs leading-tight font-semibold text-foreground transition-colors group-hover:text-signal sm:text-sm"
                  >
                    {share.title}
                  </h3>
                </div>

                <!-- 描述采用绝对定位，悬停时自底部浮现。 -->
                <div
                  class="absolute inset-x-3 bottom-3 translate-y-full opacity-0 transition-[translate,opacity] duration-500 group-hover:translate-y-0 group-hover:opacity-100 motion-reduce:transition-none sm:inset-x-6 sm:bottom-6"
                >
                  <p class="line-clamp-2 text-xs text-muted-foreground sm:line-clamp-3">
                    {share.description}
                  </p>
                </div>
              </CardContent>
            </Card>
          </div>
        {/each}
      </div>
    </div>
  </div>
</section>
