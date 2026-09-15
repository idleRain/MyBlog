import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'
import adapter from '@sveltejs/adapter-node'

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs/kit/integrations
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  kit: {
    // 前台自托管 SSR：adapter-node 产出 Node 服务，部署时以 node build 启动（OPS-03 定案）。
    adapter: adapter(),
    // 路径别名；$ui 指向 packages/ui 源码，$i18n 仅前台应用需要。
    alias: {
      $ui: '../../packages/ui/src',
      '$ui/*': '../../packages/ui/src/*',
      '~/*': './*',
      '#/*': './src/types/*',
      $i18n: './src/lib/paraglide/messages',
      '@/*': './src/*'
    }
  }
}

export default config
