<script lang="ts">
import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card'
import { Badge } from '$lib/components/ui/badge'
import { Calendar, Clock, Eye, Heart } from '@lucide/svelte'

// 模拟精选博客数据
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

// 模拟日常分享数据
const dailyShares = [
  { id: 1, emoji: '☕', title: '晨间咖啡思考', description: '今天想到的一些关于代码优雅性的思考' },
  { id: 2, emoji: '📚', title: '读书笔记', description: '《Clean Architecture》的核心要点总结' },
  { id: 3, emoji: '🎵', title: '编码音乐', description: '分享一些适合写代码时听的音乐playlist' },
  { id: 4, emoji: '🌱', title: '技术成长', description: '学习新技术的方法论和心得体会' },
  { id: 5, emoji: '🎨', title: '设计灵感', description: '在Dribbble上发现的有趣设计案例' },
  { id: 6, emoji: '🔧', title: '工具推荐', description: '提升开发效率的神器工具分享' },
  { id: 7, emoji: '💡', title: '创意想法', description: '一些有趣的项目想法和实现思路' },
  { id: 8, emoji: '🏃', title: '生活健康', description: '程序员如何保持健康的工作生活平衡' },
  { id: 9, emoji: '🎯', title: '目标规划', description: '2024年的技术学习和职业发展规划' }
]

function getCategoryColor(category: string) {
  switch (category) {
    case '技术':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300'
    case '设计':
      return 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300'
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-800/30 dark:text-gray-300'
  }
}
</script>

<section class="bg-gray-50/50 py-20 dark:bg-gray-900/50">
  <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
    <!-- 精选博客板块 -->
    <div class="mb-20">
      <!-- 标题和装饰线 -->
      <div class="mb-12 flex items-center justify-center">
        <div
          class="h-px flex-grow bg-gradient-to-r from-transparent via-gray-300 to-transparent dark:via-gray-600"
        ></div>
        <h2 class="px-6 text-3xl font-bold text-gray-900 sm:text-4xl dark:text-white">深度思考</h2>
        <div
          class="h-px flex-grow bg-gradient-to-r from-transparent via-gray-300 to-transparent dark:via-gray-600"
        ></div>
      </div>

      <!-- 博客卡片网格 -->
      <div class="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3">
        {#each featuredPosts as post, index (index)}
          <Card
            class="group cursor-pointer border border-gray-200/50 bg-white/80 pt-0 backdrop-blur-sm transition-all duration-300 hover:-translate-y-2 hover:shadow-xl dark:border-gray-700/50 dark:bg-gray-800/80"
          >
            <!-- 封面图片 -->
            <div class="relative aspect-video overflow-hidden rounded-t-lg">
              <img
                src={post.coverImage}
                alt={post.title}
                class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
              />
              <!-- 分类标签 -->
              <div class="absolute top-4 left-4">
                <Badge class={getCategoryColor(post.category)}>
                  {post.category}
                </Badge>
              </div>
              <!-- 悬停遮罩 -->
              <div
                class="absolute inset-0 bg-black/0 transition-colors duration-300 group-hover:bg-black/20"
              ></div>
            </div>

            <CardHeader class="pb-3">
              <CardTitle
                class="line-clamp-2 text-lg font-semibold text-gray-900 transition-colors group-hover:text-blue-600 dark:text-white dark:group-hover:text-blue-400"
              >
                {post.title}
              </CardTitle>
            </CardHeader>

            <CardContent class="pt-0">
              <!-- 摘要 -->
              <p class="mb-4 line-clamp-2 text-sm text-gray-600 dark:text-gray-300">
                {post.excerpt}
              </p>

              <!-- 元信息 -->
              <div
                class="mb-3 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400"
              >
                <div class="flex items-center space-x-3">
                  <div class="flex items-center space-x-1">
                    <Calendar class="h-3 w-3" />
                    <span>{post.date}</span>
                  </div>
                  <div class="flex items-center space-x-1">
                    <Clock class="h-3 w-3" />
                    <span>{post.readTime}</span>
                  </div>
                </div>
                <div class="flex items-center space-x-3">
                  <div class="flex items-center space-x-1">
                    <Eye class="h-3 w-3" />
                    <span>{post.views}</span>
                  </div>
                  <div class="flex items-center space-x-1">
                    <Heart class="h-3 w-3" />
                    <span>{post.likes}</span>
                  </div>
                </div>
              </div>

              <!-- 标签 -->
              <div class="flex flex-wrap gap-1">
                {#each post.tags as tag, index (index)}
                  <Badge variant="secondary" class="text-xs">
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
      <!-- 标题和装饰线 -->
      <div class="mb-12 flex items-center justify-center">
        <div
          class="h-px flex-grow bg-gradient-to-r from-transparent via-gray-300 to-transparent dark:via-gray-600"
        ></div>
        <h2 class="px-6 text-3xl font-bold text-gray-900 sm:text-4xl dark:text-white">生活切片</h2>
        <div
          class="h-px flex-grow bg-gradient-to-r from-transparent via-gray-300 to-transparent dark:via-gray-600"
        ></div>
      </div>

      <!-- 瀑布流布局 - 移动端两列布局优化 -->
      <div class="columns-2 gap-3 space-y-3 sm:gap-6 sm:space-y-6 lg:columns-3">
        {#each dailyShares as share, index (index)}
          <div class="break-inside-avoid">
            <Card
              class="group cursor-pointer border border-gray-200/50 bg-white/80 backdrop-blur-sm transition-all duration-300 hover:-translate-y-1 hover:shadow-lg dark:border-gray-700/50 dark:bg-gray-800/80"
            >
              <CardContent
                class="relative flex h-24 flex-col justify-start overflow-hidden p-3 sm:h-32 sm:p-6"
              >
                <!-- 表情图标和标题容器 -->
                <div class="transition-transform duration-300 group-hover:-translate-y-2">
                  <!-- 表情图标 -->
                  <div class="mb-1.5 text-xl transition-transform duration-300 sm:mb-2 sm:text-2xl">
                    <span class="transition-[font_size] group-hover:text-3xl">{share.emoji}</span>
                  </div>

                  <!-- 标题 -->
                  <h3
                    class="text-xs leading-tight font-semibold text-gray-900 transition-colors group-hover:text-blue-600 sm:text-sm dark:text-white dark:group-hover:text-blue-400"
                  >
                    {share.title}
                  </h3>
                </div>

                <!-- 描述 (绝对定位，从底部缓慢出现) -->
                <div
                  class="absolute inset-x-3 bottom-3 translate-y-full transform opacity-0 transition-all duration-500 group-hover:translate-y-0 group-hover:opacity-100 sm:inset-x-6 sm:bottom-6"
                >
                  <p class="line-clamp-2 text-xs text-gray-600 sm:line-clamp-3 dark:text-gray-300">
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

<style>
/* 多行截断样式 */
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 优化瀑布流布局 */
.columns-1 > div,
.columns-2 > div,
.columns-3 > div {
  display: inline-block;
  width: 100%;
}
</style>
