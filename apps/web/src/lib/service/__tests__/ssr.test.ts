import { createServerRequest, extractSessionCookies } from '$lib/service/ssr'
import { describe, expect, it } from 'vitest'

// 服务端请求器与浏览器的差异集中在 Cookie 转发上，本文件锚定两件事：
// 只转发本站会话 Cookie，以及未携带会话时仍返回可用的请求器。

describe('会话 Cookie 筛选', () => {
  it('保留两个会话 Cookie 并丢弃无关条目', () => {
    const header = 'mb_access_token=abc; theme=dark; mb_refresh_token=def; _ga=GA1.1'

    expect(extractSessionCookies(header)).toBe('mb_access_token=abc; mb_refresh_token=def')
  })

  it('入站头缺失时返回空字符串', () => {
    expect(extractSessionCookies(null)).toBe('')
  })

  it('不含会话 Cookie 时返回空字符串', () => {
    expect(extractSessionCookies('theme=dark; locale=zh')).toBe('')
  })

  it('只保留以 mb_ 开头的条目，前缀相近的其他键不误入', () => {
    const header = 'mbx_access_token=forged; mb_access_token=real'

    expect(extractSessionCookies(header)).toBe('mb_access_token=real')
  })

  it('容忍入站头中多余的空格', () => {
    const header = '  mb_access_token=abc ;   mb_refresh_token=def  '

    expect(extractSessionCookies(header)).toBe('mb_access_token=abc; mb_refresh_token=def')
  })
})

describe('服务端请求器构造', () => {
  // 请求器的注入结果由 packages/http 的请求拦截器用例锚定，
  // 此处只断言构造入口在两种入站情形下都产出可用实例。
  it('携带会话 Cookie 时返回可用请求器', () => {
    const client = createServerRequest('mb_access_token=abc; theme=dark')

    expect(client).toBeDefined()
    expect(typeof client.post).toBe('function')
  })

  it('未携带会话时同样返回请求器，由后端裁决登录态', () => {
    expect(createServerRequest(null)).toBeDefined()
    expect(createServerRequest('theme=dark')).toBeDefined()
  })
})
