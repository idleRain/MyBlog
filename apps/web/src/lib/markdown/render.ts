/**
 * 文章正文渲染管线：将 markdown 源文本渲染为携带 Shiki 代码高亮与标题锚点的 HTML 片段。
 *
 * 渲染在前台 SSR 的 server load 中执行，后端 goldmark 的 contentHtml 缓存保持原样，
 * 两侧遵循同一 GFM 方言；本模块是 Shiki 双主题高亮的唯一接入点。
 * 代码块在 Shiki 高亮之后统一包入 figure.code-block 外壳，头部携带语言标识与复制按钮，
 * 复制行为由 enhance-code-blocks 客户端动作模块接管。
 */

import type {
  Content as HastContent,
  Element as HastElement,
  ElementContent,
  Root as HastRoot
} from 'hast'
import { createHighlighterCore, type HighlighterGeneric, type LanguageInput } from 'shiki/core'
import rehypeShikiFromHighlighter, { type RehypeShikiCoreOptions } from '@shikijs/rehype/core'
import type { BundledLanguage, BundledTheme, ShikiTransformer } from 'shiki'
import { createJavaScriptRegexEngine } from 'shiki/engine/javascript'
import rehypeStringify from 'rehype-stringify'
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

// Shiki 内置的无语法纯文本语言，无需加载 grammar 即可按主题渲染，展示名统一为「文本」。
const PLAIN_TEXT_LANGUAGES = new Set(['text', 'txt', 'plain', 'plaintext'])
// 纯文本类代码块的展示名。
const PLAIN_TEXT_LANGUAGE_LABEL = '文本'

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
 * 语言模块按需动态加载，加载失败的语言降级为纯文本而不阻断正文渲染，
 * 纯文本语言由 Shiki 内置支持，跳过加载也不视为不支持。
 */
async function loadRequestedLanguages(
  highlighter: PipelineHighlighter,
  tree: Root
): Promise<Set<string>> {
  const loaded = new Set(highlighter.getLoadedLanguages())
  const unsupported = new Set<string>()

  for (const language of collectCodeLanguages(tree)) {
    if (loaded.has(language) || PLAIN_TEXT_LANGUAGES.has(language)) continue
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

// —— 代码块外壳：为 Shiki 输出的 pre 生成语言标识与复制按钮的容器结构 ——

// 外壳各级节点的样式类名，与 app.css 中 .code-block 系列样式一一对应。
const CODE_BLOCK_FIGURE_CLASS = 'code-block'
const CODE_BLOCK_HEAD_CLASS = 'code-block-head'
const CODE_BLOCK_LANG_CLASS = 'code-block-lang'
const CODE_BLOCK_COPY_BUTTON_CLASS = 'code-block-copy'
const CODE_BLOCK_COPY_TEXT_CLASS = 'code-block-copy-text'
const CODE_BLOCK_ICON_COPY_CLASS = 'code-block-icon-copy'
const CODE_BLOCK_ICON_CHECK_CLASS = 'code-block-icon-check'

// 复制按钮的初始文案，成功与失败文案由客户端交互模块在复制后切换。
const COPY_BUTTON_LABEL = '复制'
// 复制按钮的无障碍名称，供读屏器播报按钮用途。
const COPY_BUTTON_ARIA_LABEL = '复制代码'

// 代码块语言标注到展示名的映射，键的收录范围与 LANGUAGE_LOADERS 对齐，未收录的标注首字母大写展示。
const LANGUAGE_DISPLAY_NAMES: Record<string, string> = {
  typescript: 'TypeScript',
  ts: 'TypeScript',
  javascript: 'JavaScript',
  js: 'JavaScript',
  tsx: 'TSX',
  jsx: 'JSX',
  bash: 'Bash',
  sh: 'Shell',
  shell: 'Shell',
  shellscript: 'Shell',
  zsh: 'Shell',
  json: 'JSON',
  jsonc: 'JSONC',
  css: 'CSS',
  scss: 'SCSS',
  html: 'HTML',
  xml: 'XML',
  vue: 'Vue',
  svelte: 'Svelte',
  go: 'Go',
  golang: 'Go',
  python: 'Python',
  py: 'Python',
  sql: 'SQL',
  yaml: 'YAML',
  yml: 'YAML',
  markdown: 'Markdown',
  md: 'Markdown',
  diff: 'Diff',
  rust: 'Rust',
  rs: 'Rust',
  java: 'Java',
  kotlin: 'Kotlin',
  c: 'C',
  cpp: 'C++',
  csharp: 'C#',
  cs: 'C#',
  php: 'PHP',
  ruby: 'Ruby',
  toml: 'TOML',
  ini: 'INI',
  lua: 'Lua',
  swift: 'Swift',
  dart: 'Dart',
  graphql: 'GraphQL',
  dockerfile: 'Dockerfile',
  docker: 'Dockerfile',
  cmd: 'Batch',
  bat: 'Batch',
  batch: 'Batch',
  ansi: 'ANSI'
}

// 图标部件的几何数据表，路径取自 Lucide 的 copy 与 check 图标，配色经 currentColor 随按钮状态联动。
interface IconPart {
  tagName: 'rect' | 'path'
  attributes: Record<string, string>
}

const COPY_ICON_PARTS: IconPart[] = [
  { tagName: 'rect', attributes: { x: '8', y: '8', width: '14', height: '14', rx: '2', ry: '2' } },
  { tagName: 'path', attributes: { d: 'M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2' } }
]

const CHECK_ICON_PARTS: IconPart[] = [{ tagName: 'path', attributes: { d: 'M20 6 9 17l-5-5' } }]

/** 读取 pre 上的 Shiki 语言标注，无语言或纯文本无标注时返回 null。 */
function readCodeLanguage(pre: HastElement): string | null {
  const language = pre.properties.dataLanguage
  if (typeof language !== 'string' || language === '') return null
  return language.toLowerCase()
}

/** 解析语言标注的展示名，无语言标注返回 null 使头部只保留复制按钮。 */
function resolveLanguageLabel(language: string | null): string | null {
  if (language === null) return null
  if (PLAIN_TEXT_LANGUAGES.has(language)) return PLAIN_TEXT_LANGUAGE_LABEL
  const display = LANGUAGE_DISPLAY_NAMES[language]
  if (display) return display
  return `${language.charAt(0).toUpperCase()}${language.slice(1)}`
}

/** 构建线性图标的 hast 元素，几何数据来自部件表，输出对读屏器隐藏。 */
function createIconElement(className: string, parts: IconPart[]): HastElement {
  return {
    type: 'element',
    tagName: 'svg',
    properties: {
      className: [className],
      viewBox: '0 0 24 24',
      fill: 'none',
      stroke: 'currentColor',
      strokeWidth: '2',
      strokeLinecap: 'round',
      strokeLinejoin: 'round',
      ariaHidden: 'true'
    },
    children: parts.map(part => ({
      type: 'element',
      tagName: part.tagName,
      properties: part.attributes,
      children: []
    }))
  }
}

/** 构建复制按钮，双图标由样式按 data-copied 状态切换，文案节点带 aria-live 供读屏器播报结果。 */
function buildCopyButton(): ElementContent {
  return {
    type: 'element',
    tagName: 'button',
    properties: {
      type: 'button',
      className: [CODE_BLOCK_COPY_BUTTON_CLASS],
      dataCodeCopy: '',
      ariaLabel: COPY_BUTTON_ARIA_LABEL
    },
    children: [
      createIconElement(CODE_BLOCK_ICON_COPY_CLASS, COPY_ICON_PARTS),
      createIconElement(CODE_BLOCK_ICON_CHECK_CLASS, CHECK_ICON_PARTS),
      {
        type: 'element',
        tagName: 'span',
        properties: { className: [CODE_BLOCK_COPY_TEXT_CLASS], ariaLive: 'polite' },
        children: [{ type: 'text', value: COPY_BUTTON_LABEL }]
      }
    ]
  }
}

/** 构建代码块头部：左侧为语言展示名，无语言时只保留右侧复制按钮。 */
function buildCodeBlockHead(label: string | null): HastElement {
  const children: ElementContent[] = []
  if (label) {
    children.push({
      type: 'element',
      tagName: 'span',
      properties: { className: [CODE_BLOCK_LANG_CLASS] },
      children: [{ type: 'text', value: label }]
    })
  }
  children.push(buildCopyButton())
  return {
    type: 'element',
    tagName: 'figcaption',
    properties: { className: [CODE_BLOCK_HEAD_CLASS] },
    children
  }
}

/** 以 Shiki 输出的 pre 为内容构建代码块外壳，外壳同时透传语言标注供样式定制。 */
function buildCodeBlockShell(pre: HastElement): HastElement {
  const language = readCodeLanguage(pre)
  return {
    type: 'element',
    tagName: 'figure',
    properties: {
      className: [CODE_BLOCK_FIGURE_CLASS],
      ...(language === null ? {} : { dataLanguage: language })
    },
    children: [buildCodeBlockHead(resolveLanguageLabel(language)), pre]
  }
}

/** 深度遍历 hast 树，将全部 pre 元素替换为代码块外壳，正文中的代码块均经由此管线产出。 */
function wrapCodeBlocksWithShell(node: HastRoot | HastElement): void {
  for (let index = 0; index < node.children.length; index += 1) {
    const child = node.children[index]
    if (!child || child.type !== 'element') continue
    if (child.tagName === 'pre') {
      node.children[index] = buildCodeBlockShell(child)
      continue
    }
    wrapCodeBlocksWithShell(child)
  }
}

/** 判断子节点是否为 root 片段，该形态由 shiki 的 rehype 插件插入而 hast 类型未表达。 */
function isRootFragment(child: HastContent | HastRoot): child is HastRoot {
  return child.type === 'root'
}

/** 将树中 root 片段的孩子展开拼接回父层，恢复规范的 hast 结构供后续遍历与序列化处理。 */
function flattenShikiRootFragments(node: HastRoot | HastElement): void {
  const children = node.children as Array<HastContent | HastRoot>
  node.children = children.flatMap(child => {
    if (!isRootFragment(child)) {
      if (child.type === 'element') flattenShikiRootFragments(child)
      return [child]
    }
    flattenShikiRootFragments(child)
    return child.children
  })
}

// 代码块外壳插件：在 Shiki 高亮之后执行，先修复 root 片段结构，再为每个 pre 添加语言标识与复制按钮容器。
const rehypeCodeBlockShell: Plugin<[], HastRoot, HastRoot> = () => tree => {
  flattenShikiRootFragments(tree)
  wrapCodeBlocksWithShell(tree)
}

// 显式锚定 Shiki 插件的参数元组类型，绕开 unified 可变参数重载的推断失败。
const rehypeShikiHighlighter: Plugin<
  [PipelineHighlighter, RehypeShikiCoreOptions],
  HastRoot,
  HastRoot
> = rehypeShikiFromHighlighter

// shiki v4 的 pre 节点不再默认输出语言标注，经 transformer 写回 data-language 供外壳插件读取。
const SHIKI_PRE_LANGUAGE_TRANSFORMER: ShikiTransformer = {
  name: 'web:pre-language-attribute',
  pre(node) {
    node.properties.dataLanguage = this.options.lang
    return node
  }
}

const SHIKI_PIPELINE_OPTIONS: RehypeShikiCoreOptions = {
  themes: { light: HIGHLIGHT_THEME_LIGHT, dark: HIGHLIGHT_THEME_DARK },
  defaultColor: 'light',
  transformers: [SHIKI_PRE_LANGUAGE_TRANSFORMER]
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
    .use(rehypeCodeBlockShell)
    .use(rehypeSlug)
    .use(rehypeStringify)
    .process(source)
  return String(file)
}
