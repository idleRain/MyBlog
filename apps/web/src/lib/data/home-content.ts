/**
 * 首页展示用的占位内容，文案以编辑手记的口吻撰写。
 * 文章业务接入后，由 @myblog/api 的文章模块提供真实数据并替换本模块。
 */

/** 本期精选的版面数据，序号用于封面的大号刊号排版。 */
export type FeaturedStory = {
  id: number
  issueNo: string
  category: string
  title: string
  excerpt: string
  date: string
  readMinutes: number
}

/** 文章目录行的条目数据，index 为两位目录编号。 */
export type PostEntry = {
  id: number
  index: string
  title: string
  excerpt: string
  date: string
  readMinutes: number
}

/** 时间线中的单条记录，按年分组呈现。 */
export type TimelineEntry = {
  id: number
  month: string
  title: string
  excerpt: string
}

/** 时间线的年度分组，total 为该年记录数。 */
export type TimelineYear = {
  year: string
  total: number
  entries: TimelineEntry[]
}

/** 侧栏热门文章的条目数据。 */
export type HotPost = {
  id: number
  title: string
  date: string
  readMinutes: number
}

/** 标签云的单个标签。 */
export type TagItem = {
  name: string
}

export const featuredStory: FeaturedStory = {
  id: 1,
  issueNo: '08',
  category: '设计 · 前端工程',
  title: '设计令牌先行：一次全站视觉语言的系统化重构',
  excerpt:
    '当样式散落成一千行无序的碎片，重构就不再是风格问题，而是工程问题。这篇文章记录如何以设计令牌为锚，把混乱逐步收敛为可组合的秩序。',
  date: '2026-01-18',
  readMinutes: 8
}

export const postEntries: PostEntry[] = [
  {
    id: 1,
    index: '01',
    title: '为什么我把博客从模板站迁移到了 SvelteKit 自研',
    excerpt:
      '速度只是表象，可控才是本质。这篇记录迁移动机、架构取舍，以及一年来维护自研工具的得与失。',
    date: '2026-02-03',
    readMinutes: 6
  },
  {
    id: 2,
    index: '02',
    title: 'TypeScript 泛型的三种建模姿势：从能用走向优雅',
    excerpt:
      '泛型不是用来炫技的，而是用来约束关系的。配合三个真实业务案例，讲清楚何时该用泛型，何时该收手。',
    date: '2026-01-27',
    readMinutes: 9
  },
  {
    id: 3,
    index: '03',
    title: '深夜书桌：重读《禅与摩托车维修艺术》的三点感想',
    excerpt: '质量不是一个可以被测量的属性，而是一种专注的状态。这本书在二十年后再读，依然锋利。',
    date: '2026-01-12',
    readMinutes: 4
  },
  {
    id: 4,
    index: '04',
    title: 'Go 服务端的优雅关停：从信号处理到连接排空',
    excerpt:
      '进程退出是系统设计里最容易被轻视的环节。这篇梳理一次完整关停链路的实现细节与验证方法。',
    date: '2025-12-30',
    readMinutes: 7
  },
  {
    id: 5,
    index: '05',
    title: '写给自己的一年级：独立博客的价值复利',
    excerpt: '写作是思考的淬炼，坚持是复利的本金。盘点一年更文的真实收获，以及那些没有写完的草稿。',
    date: '2025-12-18',
    readMinutes: 5
  }
]

export const timelineYears: TimelineYear[] = [
  {
    year: '2026',
    total: 3,
    entries: [
      {
        id: 1,
        month: '02',
        title: '为什么我把博客从模板站迁移到了 SvelteKit 自研',
        excerpt: '速度只是表象，可控才是本质。'
      },
      {
        id: 2,
        month: '01',
        title: '设计令牌先行：一次全站视觉语言的系统化重构',
        excerpt: '以设计令牌为锚，把混乱收敛为可组合的秩序。'
      },
      {
        id: 3,
        month: '01',
        title: '深夜书桌：重读《禅与摩托车维修艺术》',
        excerpt: '质量是一种专注的状态，而非可测量的属性。'
      }
    ]
  },
  {
    year: '2025',
    total: 5,
    entries: [
      {
        id: 4,
        month: '12',
        title: 'Go 服务端的优雅关停：从信号处理到连接排空',
        excerpt: '进程退出是系统设计里最容易被轻视的环节。'
      },
      {
        id: 5,
        month: '11',
        title: '从零实现一个可中断的动画调度器',
        excerpt: '让动画听用户的话：随时打断，随时改向。'
      },
      {
        id: 6,
        month: '09',
        title: '笔记软件选型三年，我最终回到纯文本',
        excerpt: '工具越轻，思考越重。'
      },
      {
        id: 7,
        month: '06',
        title: '城市漫游：胡同里的五金店与算法之外的生活',
        excerpt: '当推荐流越来越同质，真实的附近反而最鲜活。'
      },
      {
        id: 8,
        month: '03',
        title: 'SQL 迁移双轨制：开发与生产的边界设计',
        excerpt: 'AutoMigrate 负责开发效率，golang-migrate 负责生产纪律。'
      }
    ]
  },
  {
    year: '2024',
    total: 4,
    entries: [
      {
        id: 9,
        month: '11',
        title: 'Monorepo 里的版本治理：pnpm catalog 实践',
        excerpt: '一处声明，处处同步，让依赖版本不再漂移。'
      },
      {
        id: 10,
        month: '08',
        title: '晨间咖啡思考：关于代码优雅性的三个断言',
        excerpt: '优雅是删不掉的东西都删掉之后剩下的样子。'
      },
      {
        id: 11,
        month: '05',
        title: '从 Hexo 到静态生成器的迁移动机',
        excerpt: '构建时间从分钟级降到秒级的完整路径。'
      },
      {
        id: 12,
        month: '02',
        title: '创刊：为什么在 2024 年开始写博客',
        excerpt: '把想清楚的事写下来，把写下来的事做出来。'
      }
    ]
  }
]

export const hotPosts: HotPost[] = [
  { id: 1, title: '从零实现一个可中断的动画调度器', date: '2025-11-02', readMinutes: 12 },
  { id: 2, title: '笔记软件选型三年，我最终回到纯文本', date: '2025-09-20', readMinutes: 5 },
  { id: 3, title: 'Monorepo 里的版本治理：pnpm catalog 实践', date: '2024-11-18', readMinutes: 9 }
]

export const tagCloud: TagItem[] = [
  { name: '前端工程' },
  { name: 'TypeScript' },
  { name: 'SvelteKit' },
  { name: 'Go' },
  { name: '设计系统' },
  { name: '读书笔记' },
  { name: '生活随笔' },
  { name: '工具链' }
]
