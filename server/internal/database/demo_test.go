package database

import (
	"strings"
	"testing"

	"MyBlog/internal/model"
)

// TestBuildDemoDatasetIsSelfConsistent 验证内置演示数据集通过自洽性校验。
func TestBuildDemoDatasetIsSelfConsistent(t *testing.T) {
	if err := validateDemoDataset(buildDemoDataset()); err != nil {
		t.Fatalf("内置演示数据集应自洽: %v", err)
	}
}

// TestDemoDatasetCoversArticleStates 验证数据集覆盖可浏览演示所需的状态与分布。
func TestDemoDatasetCoversArticleStates(t *testing.T) {
	dataset := buildDemoDataset()

	statusCount := map[model.ArticleStatus]int{}
	publishedMonths := map[int]struct{}{}
	featured, top, commentDisabled := 0, 0, 0
	for _, item := range dataset.Articles {
		statusCount[item.Status]++
		if item.IsFeatured {
			featured++
		}
		if item.IsTop {
			top++
		}
		if !item.CommentEnabled {
			commentDisabled++
		}
		if item.Status == model.ArticleStatusPublished {
			publishedMonths[item.PublishedMonthsAgo] = struct{}{}
		}
	}

	for _, status := range []model.ArticleStatus{
		model.ArticleStatusPublished,
		model.ArticleStatusDraft,
		model.ArticleStatusPrivate,
	} {
		if statusCount[status] == 0 {
			t.Errorf("演示数据集缺少 %s 状态的文章", status)
		}
	}
	if len(publishedMonths) < 3 {
		t.Errorf("已发布文章覆盖 %d 个月份，归档页至少需要 3 个", len(publishedMonths))
	}
	if featured == 0 || top == 0 {
		t.Errorf("演示数据集应包含精选与置顶文章，实际 featured=%d top=%d", featured, top)
	}
	if commentDisabled == 0 {
		t.Error("演示数据集应包含一篇关闭评论的文章，用于演示该状态")
	}
}

// TestDemoDatasetCoversCommentStates 验证评论覆盖根评论、回复与待审核状态。
func TestDemoDatasetCoversCommentStates(t *testing.T) {
	dataset := buildDemoDataset()

	roots, replies, pending, pinned := 0, 0, 0, 0
	for _, item := range dataset.Comments {
		if item.ReplyTo == "" {
			roots++
		} else {
			replies++
		}
		if item.Status == model.CommentStatusPending {
			pending++
		}
		if item.IsPinned {
			pinned++
		}
	}

	if roots == 0 || replies == 0 {
		t.Errorf("演示评论应同时包含根评论与回复，实际 roots=%d replies=%d", roots, replies)
	}
	if pending == 0 {
		t.Error("演示评论应包含待审核状态，用于演示后台审核队列")
	}
	if pinned == 0 {
		t.Error("演示评论应包含置顶评论")
	}
}

// TestDemoDatasetContentKeepsMarkdownStructure 验证正文拼接未破坏 Markdown 结构。
// 数据集正文由原始字符串与行内代码片段拼接而成，拼接错误不会导致编译失败，只能靠内容断言发现。
func TestDemoDatasetContentKeepsMarkdownStructure(t *testing.T) {
	dataset := buildDemoDataset()

	cases := []struct {
		slug     string
		markers  []string
		scenario string
	}{
		{
			slug: "structured-logging-practice",
			markers: []string{
				"~~~go",
				"`requestId`",
				"`latencyMs`",
				"| 字段 | 来源 | 用途 |",
				"> 调用方传入的请求标识",
			},
			scenario: "代码围栏、行内代码、表格与引用块",
		},
		{
			slug: "waf-field-level-exemption",
			markers: []string{
				"~~~sql",
				"INSERT INTO users",
				"DROP TABLE tmp_cache",
				"`i18n.*.content`",
			},
			scenario: "正文中的 SQL 示例与嵌套字段说明",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.scenario, func(t *testing.T) {
			content := demoArticleContent(t, dataset, testCase.slug)
			for _, marker := range testCase.markers {
				if !strings.Contains(content, marker) {
					t.Errorf("文章 %s 正文缺少 %q", testCase.slug, marker)
				}
			}
		})
	}
}

// demoArticleContent 取出指定 slug 的演示正文，缺失即判定为失败。
func demoArticleContent(t *testing.T, dataset demoDataset, slug string) string {
	t.Helper()

	for _, item := range dataset.Articles {
		if item.Slug == slug {
			return item.Content
		}
	}
	t.Fatalf("演示数据集缺少文章 %s", slug)
	return ""
}

// TestValidateDemoDatasetRejectsInconsistencies 红组：数据集不自洽时必须拒绝写入。
func TestValidateDemoDatasetRejectsInconsistencies(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(dataset *demoDataset)
	}{
		{
			name: "分类 slug 重复",
			mutate: func(dataset *demoDataset) {
				dataset.Categories = append(dataset.Categories, dataset.Categories[0])
			},
		},
		{
			name: "标签名称重复",
			mutate: func(dataset *demoDataset) {
				dataset.Tags = append(dataset.Tags, dataset.Tags[0])
			},
		},
		{
			name: "文章 slug 重复",
			mutate: func(dataset *demoDataset) {
				dataset.Articles = append(dataset.Articles, dataset.Articles[0])
			},
		},
		{
			name: "文章引用不存在的分类",
			mutate: func(dataset *demoDataset) {
				dataset.Articles[0].CategorySlug = "not-exist"
			},
		},
		{
			name: "文章引用不存在的标签",
			mutate: func(dataset *demoDataset) {
				dataset.Articles[0].TagNames = []string{"not-exist"}
			},
		},
		{
			name: "缺少已发布文章",
			mutate: func(dataset *demoDataset) {
				for index := range dataset.Articles {
					dataset.Articles[index].Status = model.ArticleStatusDraft
				}
			},
		},
		{
			name: "评论引用不存在的文章",
			mutate: func(dataset *demoDataset) {
				dataset.Comments[0].ArticleSlug = "not-exist"
			},
		},
		{
			name: "回复指向不存在的根评论",
			mutate: func(dataset *demoDataset) {
				dataset.Comments[2].ReplyTo = "不存在的作者"
			},
		},
		{
			name: "同文章根评论作者名重复导致幂等键歧义",
			mutate: func(dataset *demoDataset) {
				dataset.Comments[1].AuthorName = dataset.Comments[0].AuthorName
			},
		},
		{
			name: "回复与根评论作者名重复导致幂等键歧义",
			mutate: func(dataset *demoDataset) {
				dataset.Comments[2].AuthorName = dataset.Comments[0].AuthorName
			},
		},
		{
			name: "同文章两回复作者名重复",
			mutate: func(dataset *demoDataset) {
				dataset.Comments = append(dataset.Comments, demoComment{
					ArticleSlug: dataset.Comments[2].ArticleSlug,
					AuthorName:  dataset.Comments[2].AuthorName,
					Content:     "重复作者的回复",
					Status:      model.CommentStatusApproved,
					ReplyTo:     dataset.Comments[1].AuthorName,
				})
			},
		},
		{
			name: "评论缺少作者名",
			mutate: func(dataset *demoDataset) {
				dataset.Comments[0].AuthorName = ""
			},
		},
		{
			name: "标签 slug 为空",
			mutate: func(dataset *demoDataset) {
				dataset.Tags[1].Slug = ""
			},
		},
		{
			name: "标签 slug 形态非法",
			mutate: func(dataset *demoDataset) {
				dataset.Tags[1].Slug = "SvelteKit"
			},
		},
		{
			name: "标签 slug 重复",
			mutate: func(dataset *demoDataset) {
				dataset.Tags[1].Slug = dataset.Tags[0].Slug
			},
		},
		{
			name: "分类 slug 为空",
			mutate: func(dataset *demoDataset) {
				dataset.Categories[1].Slug = ""
			},
		},
		{
			name: "文章 slug 形态非法",
			mutate: func(dataset *demoDataset) {
				dataset.Articles[1].Slug = "含中文的标识"
			},
		},
		{
			name: "分类为空",
			mutate: func(dataset *demoDataset) {
				dataset.Categories = nil
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			dataset := buildDemoDataset()
			testCase.mutate(&dataset)

			if err := validateDemoDataset(dataset); err == nil {
				t.Error("不自洽的数据集应返回错误")
			}
		})
	}
}

// TestSeedDemoDataGatesBeforeDatabase 验证写入前置校验在触达数据库之前生效。
// 该用例只覆盖数据库之前的两道闸门，写入路径的验证见交接报告中的实机记录。
func TestSeedDemoDataGatesBeforeDatabase(t *testing.T) {
	t.Run("release 模式默认拒绝", func(t *testing.T) {
		err := SeedDemoData(nil, DemoSeedOptions{AuthorID: 1, ServerMode: productionServerMode})
		if err == nil {
			t.Fatal("release 模式下应拒绝写入演示数据")
		}
		if !strings.Contains(err.Error(), "--force") {
			t.Errorf("拒绝原因应指出放行方式，实际: %v", err)
		}
	})

	t.Run("缺少作者时拒绝", func(t *testing.T) {
		if err := SeedDemoData(nil, DemoSeedOptions{ServerMode: "debug"}); err == nil {
			t.Error("缺少作者时应拒绝写入演示数据")
		}
	})
}
