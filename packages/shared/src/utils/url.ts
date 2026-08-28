/**
 * 获取 URL 参数，key 提供时返回对应值，否则返回全部参数对象。
 */
export function getUrlParams<T extends string | Record<string, string>>(key?: string): T {
  const url = new URL(location.href)
  const search = url.search || url.hash.split('?')[1] || ''
  const params = new URLSearchParams(search)
  if (key) {
    const value = params.get(key)
    return (value ? decodeURIComponent(value.replace(/\+/g, ' ')) : value) as T
  }
  const result: Record<string, string> = {}
  for (const [k, v] of params.entries()) {
    result[k] = decodeURIComponent(v.replace(/\+/g, ' '))
  }
  return result as T
}
