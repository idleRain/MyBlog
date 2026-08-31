import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'
import adapter from '@sveltejs/adapter-auto'

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs/kit/integrations
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  kit: {
    // adapter-auto only supports some environments, see https://svelte.dev/docs/kit/adapter-auto for a list.
    // If your environment is not supported, or you settled on a specific environment, switch out the adapter.
    // See https://svelte.dev/docs/kit/adapters for more information about adapters.
    adapter: adapter(),
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
