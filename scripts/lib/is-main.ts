import { resolve } from 'node:path'
import { pathToFileURL } from 'node:url'

/**
 * 判断当前模块是否为直接运行的入口模块。
 * Node.js 的 ESM 不提供 import.meta.main，此函数比较进程入口与模块地址。
 * @param importMetaUrl 调用方的 import.meta.url
 */
export function isMainModule(importMetaUrl: string): boolean {
  const entryArg = process.argv[1]
  if (!entryArg) return false
  try {
    const entryUrl = pathToFileURL(resolve(entryArg)).href
    return entryUrl === importMetaUrl
  } catch {
    return false
  }
}
