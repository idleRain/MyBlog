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
    // 路径别名
    alias: {
      $ui: './src/lib/components/ui',
      '$ui/*': './src/lib/components/ui/*',
      '~/*': './*',
      '#/*': './src/types/*',
      $i18n: './src/lib/paraglide/messages',
      '@/*': './src/*'
    }
  }
}

export default config
