import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'
import adapter from '@sveltejs/adapter-static'

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs/kit/integrations
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  kit: {
    // 后台为 ssr=false 的 SPA：adapter-static 产出纯静态文件，
    // fallback 指向 index.html，所有未知路由回退到单页入口（OPS-03 定案）。
    adapter: adapter({ fallback: 'index.html' }),
    // 后台应用部署在站点 /admin 子路径下，与前台同源区分；relative 关闭以使用绝对的根路径。
    paths: {
      base: '/admin',
      relative: false
    },
    // 路径别名；$ui 指向 packages/ui 源码，后台应用无需 i18n 别名。
    alias: {
      $ui: '../../packages/ui/src',
      '$ui/*': '../../packages/ui/src/*',
      '~/*': './*',
      '#/*': './src/types/*',
      '@/*': './src/*'
    }
  }
}

export default config
