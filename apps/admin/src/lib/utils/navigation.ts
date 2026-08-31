// 基准路径感知的导航工具：设置 kit.paths.base 后，SvelteKit 不会自动为根路径补充前缀。

import { goto as svelteGoto } from '$app/navigation'
import { base } from '$app/paths'

/**
 * 将应用内根路径解析为包含基准路径的完整路径。
 * 可携带查询参数或哈希，例如 /login、/posts?action=create。
 * @param path 以 / 开头的应用内路径。
 */
export function toAdminPath(path: string): string {
  return base + path
}

/**
 * 使用基准路径解析后的地址执行客户端导航，行为与 $app/navigation 的 goto 一致。
 * @param path 以 / 开头的应用内根路径。
 * @param opts 透传给 SvelteKit goto 的导航选项。
 */
export function goto(path: string, opts?: Parameters<typeof svelteGoto>[1]): Promise<void> {
  return svelteGoto(toAdminPath(path), opts)
}
