import { deLocalizeUrl } from '$lib/paraglide/runtime'

// SvelteKit 重路由钩子：将带语言前缀的地址还原为逻辑路由地址。
export const reroute = (request: Request) => deLocalizeUrl(request.url).pathname
