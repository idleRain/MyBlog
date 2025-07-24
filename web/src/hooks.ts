import { deLocalizeUrl } from '$lib/paraglide/runtime'

// eslint-disable-next-line prettier/prettier
export const reroute = request => deLocalizeUrl(request.url).pathname
