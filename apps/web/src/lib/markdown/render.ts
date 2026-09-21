/**
 * 文章正文渲染管线：将 markdown 源文本渲染为携带 Shiki 代码高亮与标题锚点的 HTML 片段。
 *
 * 渲染在前台 SSR 的 server load 中执行，后端 goldmark 的 contentHtml 缓存保持原样，
 * 两侧遵循同一 GFM 方言；本模块是 Shiki 双主题高亮的唯一接入点。
 */

import { createHighlighterCore, type HighlighterGeneric, type LanguageInput } from 'shiki/core'
import rehypeShikiFromHighlighter, { type RehypeShikiCoreOptions } from '@shikijs/rehype/core'
import { createJavaScriptRegexEngine } from 'shiki/engine/javascript'
import type { BundledLanguage, BundledTheme } from 'shiki'
import rehypeStringify from 'rehype-stringify'
import type { Root as HastRoot } from 'hast'
import { visit } from 'unist-util-visit'
import remarkRehype from 'remark-rehype'
import type { Code, Root } from 'mdast'
import remarkParse from 'remark-parse'
import type { Plugin } from 'unified'
import rehypeSlug from 'rehype-slug'
import remarkGfm from 'remark-gfm'
import { unified } from 'unified'

// 管线统一使用完整泛型签名的高亮器类型；core 模式的 never 泛型与 rehype 插件参数不兼容。
type PipelineHighlighter = HighlighterGeneric<BundledLanguage, BundledTheme>

// 双主题输出采用 CSS 变量方案，亮色为内联默认值，暗色经 --shiki-dark 变量随 html.dark 切换。
const HIGHLIGHT_THEME_LIGHT = 'vitesse-light'
const HIGHLIGHT_THEME_DARK = 'vitesse-dark'

// 首发预注册的常用语言，博客高频技术栈的首次渲染无需触发动态加载。
const PRESET_LANGUAGES: LanguageInput[] = [
  () => import('@shikijs/langs/typescript'),
  () => import('@shikijs/langs/javascript'),
  () => import('@shikijs/langs/tsx'),
  () => import('@shikijs/langs/jsx'),
  () => import('@shikijs/langs/bash'),
  () => import('@shikijs/langs/json'),
  () => import('@shikijs/langs/css'),
  () => import('@shikijs/langs/html'),
  () => import('@shikijs/langs/go'),
  () => import('@shikijs/langs/python'),
  () => import('@shikijs/langs/sql'),
  () => import('@shikijs/langs/yaml'),
  () => import('@shikijs/langs/markdown'),
  () => import('@shikijs/langs/diff')
]

// 代码块语言标注到 Shiki 语言模块的映射，覆盖常用别名；未收录的标注按纯文本渲染。
const LANGUAGE_LOADERS: Record<string, LanguageInput> = {
  typescript: () => import('@shikijs/langs/typescript'),
  ts: () => import('@shikijs/langs/typescript'),
  javascript: () => import('@shikijs/langs/javascript'),
  js: () => import('@shikijs/langs/javascript'),
  tsx: () => import('@shikijs/langs/tsx'),
  jsx: () => import('@shikijs/langs/jsx'),
  bash: () => import('@shikijs/langs/bash'),
  sh: () => import('@shikijs/langs/sh'),
  shell: () => import('@shikijs/langs/shell'),
  shellscript: () => import('@shikijs/langs/shellscript'),
  zsh: () => import('@shikijs/langs/bash'),
  json: () => import('@shikijs/langs/json'),
  jsonc: () => import('@shikijs/langs/jsonc'),
  css: () => import('@shikijs/langs/css'),
  scss: () => import('@shikijs/langs/scss'),
  html: () => import('@shikijs/langs/html'),
  xml: () => import('@shikijs/langs/xml'),
  vue: () => import('@shikijs/langs/vue'),
  svelte: () => import('@shikijs/langs/svelte'),
  go: () => import('@shikijs/langs/go'),
  golang: () => import('@shikijs/langs/go'),
  python: () => import('@shikijs/langs/python'),
  py: () => import('@shikijs/langs/py'),
  sql: () => import('@shikijs/langs/sql'),
  yaml: () => import('@shikijs/langs/yaml'),
  yml: () => import('@shikijs/langs/yml'),
  markdown: () => import('@shikijs/langs/markdown'),
  md: () => import('@shikijs/langs/md'),
  diff: () => import('@shikijs/langs/diff'),
  rust: () => import('@shikijs/langs/rust'),
  rs: () => import('@shikijs/langs/rs'),
  java: () => import('@shikijs/langs/java'),
  kotlin: () => import('@shikijs/langs/kotlin'),
  c: () => import('@shikijs/langs/c'),
  cpp: () => import('@shikijs/langs/cpp'),
  csharp: () => import('@shikijs/langs/csharp'),
  cs: () => import('@shikijs/langs/cs'),
  php: () => import('@shikijs/langs/php'),
  ruby: () => import('@shikijs/langs/ruby'),
  toml: () => import('@shikijs/langs/toml'),
  ini: () => import('@shikijs/langs/ini'),
  lua: () => import('@shikijs/langs/lua'),
  swift: () => import('@shikijs/langs/swift'),
  dart: () => import('@shikijs/langs/dart'),
  graphql: () => import('@shikijs/langs/graphql'),
  dockerfile: () => import('@shikijs/langs/dockerfile'),
  docker: () => import('@shikijs/langs/docker'),
  cmd: () => import('@shikijs/langs/cmd'),
  bat: () => import('@shikijs/langs/bat'),
  batch: () => import('@shikijs/langs/batch')
}

// 模块级单例复用 grammar 与主题资源，避免每次渲染重建高亮器。
let highlighterPromise: Promise<PipelineHighlighter> | null = null

/**
 * 懒创建 Shiki 高亮器单例，采用 JavaScript 正则引擎避免引入 WASM 运行时。
 * forgiving 选项让个别语法不完全兼容的正则静默降级而非抛错。
 * core 高亮器实例的泛型签名提升为完整语言联合，实际可用语言以已注册集合为准。
 */
function getHighlighter(): Promise<PipelineHighlighter> {
  highlighterPromise ??= createHighlighterCore({
    themes: [import('@shikijs/themes/vitesse-light'), import('@shikijs/themes/vitesse-dark')],
    langs: PRESET_LANGUAGES,
    engine: createJavaScriptRegexEngine({ forgiving: true })
  }) as Promise<PipelineHighlighter>
  return highlighterPromise
}

/** 收集正文中全部代码块语言标注，小写化并去重以保持加载顺序稳定。 */
function collectCodeLanguages(tree: Root): string[] {
  const languages = new Set<string>()
  visit(tree, 'code', (node: Code) => {
    const language = node.lang?.trim().toLowerCase()
    if (language) languages.add(language)
  })
  return [...languages]
}

/** 清空无法高亮语言标注的 lang 字段，使对应代码块按纯文本渲染。 */
function stripUnsupportedLanguages(tree: Root, unsupported: Set<string>): void {
  visit(tree, 'code', (node: Code) => {
    const language = node.lang?.trim().toLowerCase()
    if (language && unsupported.has(language)) node.lang = null
  })
}

/**
 * 确保 Shiki 高亮器已就绪正文请求的全部语言，返回无法加载的语言集合。
 * 语言模块按需动态加载，加载失败的语言降级为纯文本而不阻断正文渲染。
 */
async function loadRequestedLanguages(
  highlighter: PipelineHighlighter,
  tree: Root
): Promise<Set<string>> {
  const loaded = new Set(highlighter.getLoadedLanguages())
  const unsupported = new Set<string>()

  for (const language of collectCodeLanguages(tree)) {
    if (loaded.has(language)) continue
    const loader = LANGUAGE_LOADERS[language]
    if (!loader) {
      unsupported.add(language)
      continue
    }
    try {
      await highlighter.loadLanguage(loader)
      for (const name of highlighter.getLoadedLanguages()) loaded.add(name)
    } catch {
      unsupported.add(language)
    }
  }
  return unsupported
}

/** remark 转换器：渲染前补齐代码块语言，保证 Shiki 转换阶段不会遇到未注册语言。 */
const remarkEnsureHighlightLanguages: Plugin<[PipelineHighlighter], Root, Root> = highlighter => {
  return async tree => {
    const unsupported = await loadRequestedLanguages(highlighter, tree)
    stripUnsupportedLanguages(tree, unsupported)
  }
}

// 显式锚定 Shiki 插件的参数元组类型，绕开 unified 可变参数重载的推断失败。
const rehypeShikiHighlighter: Plugin<
  [PipelineHighlighter, RehypeShikiCoreOptions],
  HastRoot,
  HastRoot
> = rehypeShikiFromHighlighter

const SHIKI_PIPELINE_OPTIONS: RehypeShikiCoreOptions = {
  themes: { light: HIGHLIGHT_THEME_LIGHT, dark: HIGHLIGHT_THEME_DARK },
  defaultColor: 'light'
}

/**
 * 渲染 markdown 源文本为 HTML 片段。
 * 原始 HTML 与 goldmark 的 Unsafe 关闭行为一致地被丢弃，输出可安全挂载。
 */
export async function renderMarkdown(source: string): Promise<string> {
  const highlighter = await getHighlighter()
  const file = await unified()
    .use(remarkParse)
    .use(remarkGfm)
    .use(remarkEnsureHighlightLanguages, highlighter)
    .use(remarkRehype)
    .use(rehypeShikiHighlighter, highlighter, SHIKI_PIPELINE_OPTIONS)
    .use(rehypeSlug)
    .use(rehypeStringify)
    .process(source)
  return String(file)
}
