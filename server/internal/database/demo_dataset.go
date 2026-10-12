// 演示数据集内容定义。
//
// 正文使用波浪号围栏（~~~）而非反引号围栏：Go 原始字符串无法包含反引号，
// 波浪号围栏是 CommonMark 的等价写法，服务端 goldmark 与前台渲染管线均按同一语义处理。
package database

import "MyBlog/internal/model"

// buildDemoDataset 构造演示数据集。
// 内容固定不随时间变化，配合按幂等键复用的写入逻辑，可在任意环境得到可复现的演示数据。
func buildDemoDataset() demoDataset {
	return demoDataset{
		Categories: []demoCategory{
			{
				Slug:        "engineering",
				Name:        "工程实践",
				Description: "架构取舍、重构记录与工具链踩坑。",
				SortOrder:   1,
				IsFeatured:  true,
			},
			{
				Slug:        "reading",
				Name:        "读书笔记",
				Description: "技术书与论文的阅读摘录与批注。",
				SortOrder:   2,
			},
			{
				Slug:        "life",
				Name:        "生活随笔",
				Description: "与代码无关的记录。",
				SortOrder:   3,
			},
		},
		Tags: []demoTag{
			{Name: "Go", Slug: "go", Color: "#00ADD8", IsHot: true},
			{Name: "SvelteKit", Slug: "sveltekit", Color: "#FF3E00"},
			{Name: "MySQL", Slug: "mysql", Color: "#00758F"},
			{Name: "可观测性", Slug: "observability", Color: "#5B6C7F", IsHot: true},
			{Name: "重构", Slug: "refactor", Color: "#8E44AD"},
		},
		Articles: []demoArticle{
			{
				Slug:           "structured-logging-practice",
				Title:          "把访问日志改成结构化 JSON 之后",
				Summary:        "从文本行到 JSON 行，日志的检索方式与关联能力发生了什么变化。",
				CategorySlug:   "engineering",
				TagNames:       []string{"Go", "可观测性"},
				Status:         model.ArticleStatusPublished,
				IsFeatured:     true,
				IsTop:          true,
				ViewCount:      128,
				LikeCount:      12,
				BookmarkCount:  5,
				CommentEnabled: true,
				Content: `## 背景

站点早期用的是 gin 默认的文本访问日志，一行里塞了客户端地址、方法、路径、状态码与耗时。
它在单次排障时够用，但有两个绕不开的问题：

1. 想按状态码聚合时只能靠正则，字段顺序一变正则就失效；
2. 一次请求产生的多条记录之间没有任何关联字段，无法串起来看。

## 落地方式

改用标准库的 ` + "`log/slog`" + ` 输出 JSON，字段键集中声明为常量，访问日志与错误日志共用同一套键名。

~~~go
attributes := []slog.Attr{
	slog.String(logger.FieldMethod, c.Request.Method),
	slog.Int(logger.FieldStatus, c.Status(c.Writer.Status())),
	slog.Int64(logger.FieldLatencyMs, time.Since(start).Milliseconds()),
}
logger.FromGin(c).LogAttrs(ctx, level, "请求处理完成", attributes...)
~~~

关键改动是**请求标识**：它由中间件写入上下文并回写响应头，此后同一次请求的访问日志、
错误日志与 panic 记录都会带上同一个 ` + "`requestId`" + `。

## 字段对照

| 字段 | 来源 | 用途 |
|---|---|---|
| ` + "`requestId`" + ` | 中间件生成或复用调用方传入值 | 串联同一次请求的全部记录 |
| ` + "`userId`" + ` | 认证中间件写入的上下文 | 定位到具体操作者 |
| ` + "`latencyMs`" + ` | 请求开始到响应写出 | 延迟分位的原始数据 |

> 调用方传入的请求标识只在长度不超过 64 且仅含安全字符时复用，其余一律替换为服务端生成的 UUID，
> 否则一个超长请求头就能把日志检索搅乱。

## 小结

结构化日志的价值不在格式本身，而在**字段可聚合、请求可串联**这两件事上。
格式换完但字段随意命名，收益会立刻归零。`,
			},
			{
				Slug:           "waf-field-level-exemption",
				Title:          "内容级黑名单为什么要做字段级豁免",
				Summary:        "讲解 SQL 与 HTML 的技术文章被 WAF 拦下之后，我们如何划定扫描边界。",
				CategorySlug:   "engineering",
				TagNames:       []string{"Go", "重构"},
				Status:         model.ArticleStatusPublished,
				ViewCount:      96,
				LikeCount:      9,
				CommentEnabled: true,
				Content: `## 问题

内容级黑名单对整个 JSON 请求体做模式扫描，于是下面这段完全正常的文章正文会被判为攻击：

~~~sql
INSERT INTO users (name) VALUES ('demo');
DELETE FROM sessions WHERE expired_at < NOW();
DROP TABLE tmp_cache;
~~~

同理，正文里提到 ` + "`<script>`" + ` 标签、或写下 ` + "`../../assets/logo.png`" + ` 这样的相对路径，都会命中模式。

## 划定边界

处理办法是按**字段路径**扫描，而不是按整个请求体：

- 正文（` + "`content`" + ` 与 ` + "`i18n.*.content`" + `）不参与模式扫描；
- 标题、摘要、SEO 字段、URL 参数、请求头与表单参数保持全量扫描。

这条边界成立的前提是**正文的输出必经消毒渲染**：服务端渲染关闭原始 HTML，前台渲染管线丢弃原始 HTML，
两处都有独立的回归用例。豁免的代价因此被限制在「正文里可以存下可执行标记，但它永远不会变成可执行标记」。

## 两个副产品

按解码后的取值扫描之后，JSON 转义序列不再遮蔽模式中的词首边界，行首的 SQL 特征重新进入检测范围；
结构无法解析时退化为整段扫描，避免借畸形结构绕过字段判定。

> 扩大豁免清单前必须先确认该字段的每一处输出都经过消毒渲染或转义输出，
> 否则就是拿安全换便利。`,
			},
			{
				Slug:               "sveltekit-ssr-session",
				Title:              "SvelteKit SSR 下的会话传递",
				Summary:            "服务端渲染的请求由 Node 进程发出，浏览器不会自动带上会话 Cookie。",
				CategorySlug:       "reading",
				TagNames:           []string{"SvelteKit", "Go"},
				Status:             model.ArticleStatusPublished,
				PublishedMonthsAgo: 3,
				ViewCount:          210,
				LikeCount:          21,
				BookmarkCount:      8,
				CommentEnabled:     true,
				Content: `## 现象

把取数逻辑搬到 ` + "`+page.server.ts`" + ` 之后，页面首屏是空的：接口返回未登录。
原因不在鉴权逻辑，而在请求的发出方——**服务端渲染的请求由 Node 进程发起，浏览器不会自动携带 Cookie**。

## 处理

在服务端筛选会话 Cookie 并转发，同时按请求新建请求器：

~~~ts
// 只转发会话 Cookie，其余请求头保持服务端默认行为
const cookies = event.request.headers.get('cookie')
const sessionCookies = filterSessionCookies(cookies)

const client = createHttpClient({ headers: { cookie: sessionCookies } })
~~~

注意不要复用应用级单例请求器：它注册了续期与失效跳转回调，
而服务端既无法把新令牌写回浏览器，也不应该触发客户端跳转。

## 验证清单

1. 登录后访问受保护页面，查看页面源码应直接含数据，而不是先渲染空壳再补拉；
2. 未携带会话时，服务端应返回重定向而不是渲染空态；
3. 越界页码应落到空态，不应回落到第一页。

> 服务端渲染的收益是首屏直接含数据，如果取数仍留在客户端，那只是把同一件事做两遍。`,
			},
			{
				Slug:               "mysql-index-from-filesort",
				Title:              "从 filesort 到命中索引：一次排序查询的排查",
				Summary:            "按点赞数排序的列表查询为什么走了文件排序，以及补索引之后的变化。",
				CategorySlug:       "life",
				TagNames:           []string{"MySQL"},
				Status:             model.ArticleStatusPublished,
				PublishedMonthsAgo: 5,
				ViewCount:          340,
				LikeCount:          30,
				CommentEnabled:     false,
				Content: `## 现象

按点赞数排序的文章列表在数据量上来之后明显变慢。` + "`EXPLAIN`" + ` 显示 ` + "`Extra`" + ` 里带 ` + "`Using filesort`" + `。

## 原因

排序字段没有索引，查询只能把命中行取回后再排序。对照表结构可以确认这一点：

| 查询模式 | 依赖的索引 | 补索引前的行为 |
|---|---|---|
| 按发布时间倒序 | ` + "`idx_status_published`" + ` | 命中索引 |
| 按点赞数倒序 | 无 | 文件排序 |
| 按昵称模糊检索 | 无 | 全表扫描 |

## 处理

补上缺失的索引，并且**同时**维护迁移脚本与模型标签两处：

~~~sql
CREATE INDEX idx_articles_like_count ON articles (like_count);
CREATE INDEX idx_users_nickname ON users (nickname);
~~~

两处必须一致：迁移脚本管生产库的增量变更，模型标签管开发库的自动迁移。
只改一处，开发环境与生产环境的表结构就会分叉，而这种分叉往往在很久以后才被发现。

> 索引不是越多越好。每加一条索引，写入路径就多一次维护成本，
> 加之前先确认这条查询真的高频。`,
			},
			{
				Slug:         "reading-notes-draft",
				Title:        "读书笔记：待整理",
				Summary:      "草稿状态的示例文章，用于演示未发布内容在后台与作者视图中的表现。",
				CategorySlug: "engineering",
				TagNames:     []string{"重构"},
				Status:       model.ArticleStatusDraft,
				Content: `## 待整理

这篇保持草稿状态，用来演示几件事：

- 未发布内容不会出现在公开列表与聚合面；
- 作者本人与管理员在列表接口中可以看到它；
- 归档页只统计已发布文章，因此这里不会出现它的月份。

## 摘录

> 把取舍写下来，是让取舍成立的必要条件。`,
			},
			{
				Slug:               "private-release-checklist",
				Title:              "私有：上线前检查清单",
				Summary:            "私有状态示例文章，仅登录且具备权限的账号可见。",
				CategorySlug:       "reading",
				TagNames:           []string{"可观测性", "Go"},
				Status:             model.ArticleStatusPrivate,
				PublishedMonthsAgo: 2,
				ViewCount:          15,
				LikeCount:          1,
				CommentEnabled:     true,
				Content: `## 检查清单

- 配置项：确认发布模式下会话 Cookie 的传输限制已开启；
- 依赖：确认漏洞扫描与依赖审计均为零阻塞项；
- 数据：确认迁移脚本与模型标签一致；
- 观测：确认指标抓取目标可达，告警规则已加载；
- 回滚：确认回滚路径与备份可用。

## 说明

私有文章不进入公开列表，但作者与管理员在列表接口中可见。
它适合承载上线流程、运维手册这类不适合公开的内容。`,
			},
		},
		Comments: []demoComment{
			{
				ArticleSlug: "structured-logging-practice",
				AuthorName:  "阿岚",
				Content:     "结构化日志上线之后，按 requestId 检索确实方便很多，排障时不用再拼时间窗口了。",
				Status:      model.CommentStatusApproved,
			},
			{
				ArticleSlug: "structured-logging-practice",
				AuthorName:  "老周",
				Content:     "想问一下输出到文件时的轮转是怎么处理的？直接写同一个文件会不会越来越大？",
				Status:      model.CommentStatusApproved,
				IsPinned:    true,
			},
			{
				ArticleSlug: "structured-logging-practice",
				AuthorName:  "站长",
				Content:     "应用侧不内置轮转，交给采集端或系统侧的 logrotate 处理，这样职责更清楚。",
				Status:      model.CommentStatusApproved,
				ReplyTo:     "老周",
			},
			{
				ArticleSlug: "structured-logging-practice",
				AuthorName:  "游客小 K",
				Content:     "这条评论保持待审核状态，用来演示后台的评论审核队列。",
				Status:      model.CommentStatusPending,
			},
			{
				ArticleSlug: "sveltekit-ssr-session",
				AuthorName:  "小满",
				Content:     "Cookie 转发这一段很受用，之前一直卡在服务端拿不到身份这一步。",
				Status:      model.CommentStatusApproved,
			},
		},
	}
}
