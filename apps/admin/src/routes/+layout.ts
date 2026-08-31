// 后台应用为纯客户端渲染，服务端渲染仅保留给前台 web 使用。

// 关闭服务端渲染，所有页面由浏览器端路由与渲染完成。
export const ssr = false

// 客户端路由分发在运行时进行，不做静态预渲染。
export const prerender = false
