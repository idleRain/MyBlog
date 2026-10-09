import { describe, expect, it } from 'vitest'
import { renderMarkdown } from './render'

// 正文渲染是存储型注入的唯一入口：文章源文本由作者提供，渲染结果经 {@html} 挂载。
// 当前纵深防御完全依赖两条配置常数——remark-rehype 未开启 allowDangerousHtml，
// 后端 goldmark 也未开启 Unsafe。配置一旦被改动，全站正文即刻变成注入通道，
// 而此前的自动化守卫为零。本文件即为该配置的回归锚点，同时锚定正常渲染不回退。

// 危险或结构性的标签名，原始 HTML 一旦经渲染管线透传就会以真实元素出现。
const STRUCTURAL_TAGS = ['script', 'iframe', 'img', 'div', 'object', 'embed', 'svg']

// 内联事件处理器属性名，透传的原始 HTML 会以此形式出现在输出中。
const EVENT_HANDLER_PATTERN = /\son[a-z]+\s*=\s*["']?[^\s"'>]/i

/**
 * 断言渲染结果中不存在给定标签的真实元素。
 * 标签可能以真实元素形式透传，也可能被转义为实体文本，两种形态都必须被视为违规。
 */
function expectNoElement(html: string, tag: string) {
  expect(html, `不应出现 <${tag}> 元素`).not.toMatch(new RegExp(`<${tag}[\\s>/]`, 'i'))
}

describe('正文渲染：原始 HTML 必须被丢弃', () => {
  it('丢弃 script 标签，不产生可执行标签', async () => {
    const html = await renderMarkdown('<script>alert(1)</script>')

    expectNoElement(html, 'script')
    expect(html).not.toContain('alert(1)')
  })

  it('丢弃内联事件属性', async () => {
    const html = await renderMarkdown('<img src=x onerror=alert(2)>')

    expectNoElement(html, 'img')
    expect(EVENT_HANDLER_PATTERN.test(html), '不应出现内联事件属性').toBe(false)
  })

  it('丢弃普通原始 HTML 块，不作为元素透传', async () => {
    const html = await renderMarkdown('<p>raw paragraph</p>')

    expectNoElement(html, 'p')
  })

  it('丢弃 iframe 元素', async () => {
    const html = await renderMarkdown('<iframe src="https://evil.example"></iframe>')

    expectNoElement(html, 'iframe')
  })

  it('丢弃带样式的 div 与 svg 等结构性标签', async () => {
    const html = await renderMarkdown(
      '<div style="position:fixed">覆盖</div>\n\n<svg onload=alert(4)>'
    )

    expectNoElement(html, 'div')
    expectNoElement(html, 'svg')
    expect(EVENT_HANDLER_PATTERN.test(html), '不应出现内联事件属性').toBe(false)
  })

  it('原始 HTML 与正常 markdown 混排时，仍只保留 markdown 结构', async () => {
    const html = await renderMarkdown('<div onclick="alert(3)">x</div>\n\n**粗体**')

    expectNoElement(html, 'div')
    expect(EVENT_HANDLER_PATTERN.test(html), '不应出现内联事件属性').toBe(false)
    expect(html).toContain('<strong>粗体</strong>')
  })

  // 逐标签参数化断言，确保任一结构性标签被放行时都会单独报红。
  // 单个用例遍历全部标签会在首个失败处中断，无法暴露其余标签的状态。
  it.each(STRUCTURAL_TAGS)('丢弃 %s 标签', async tag => {
    const html = await renderMarkdown(`<${tag} data-probe="1">负载</${tag}>`)

    expectNoElement(html, tag)
    expect(html).not.toContain('data-probe')
  })
})

describe('正文渲染：正常内容不回退', () => {
  it('渲染段落与行内格式', async () => {
    const html = await renderMarkdown('段落文本包含 **粗体** 与 *斜体*。')

    expect(html).toContain('<p>')
    expect(html).toContain('<strong>粗体</strong>')
    expect(html).toContain('<em>斜体</em>')
  })

  it('为标题生成锚点 id', async () => {
    const html = await renderMarkdown('## 二级标题')

    expect(html).toContain('<h2 id="二级标题">')
  })

  it('渲染 GFM 表格', async () => {
    const html = await renderMarkdown('| 列一 | 列二 |\n| --- | --- |\n| a | b |')

    expect(html).toContain('<table>')
    expect(html).toContain('<th>列一</th>')
  })

  it('代码块经高亮管线输出并携带语言标识', async () => {
    const html = await renderMarkdown('```ts\nconst value: number = 1\n```')

    expect(html).toContain('data-language="ts"')
    expect(html).toContain('<figure')
  })
})
