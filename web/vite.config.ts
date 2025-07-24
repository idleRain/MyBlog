import { type ConfigEnv, defineConfig, loadEnv } from 'vite'
import { paraglideVitePlugin } from '@inlang/paraglide-js'
import devtoolsJson from 'vite-plugin-devtools-json'
import AutoImport from 'unplugin-auto-import/vite'
import { sveltekit } from '@sveltejs/kit/vite'
import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import ViteJson5 from 'vite-plugin-json5'

export default ({ mode }: ConfigEnv) => {
  const env = loadEnv(mode, process.cwd())

  return defineConfig({
    plugins: [
      tailwindcss(),
      sveltekit(),
      devtoolsJson(),
      ViteJson5(),
      AutoImport({
        // 自动导入常用的 SvelteKit 和 Svelte 函数
        imports: [
          {
            // SvelteKit 核心
            '$app/environment': ['browser', 'dev', 'building', 'version'],
            '$app/navigation': [
              'goto',
              'invalidate',
              'invalidateAll',
              'preloadData',
              'preloadCode',
              'beforeNavigate',
              'afterNavigate'
            ],
            '$app/stores': ['page', 'navigating', 'updated'],
            // Svelte 核心
            svelte: [
              'onMount',
              'onDestroy',
              'beforeUpdate',
              'afterUpdate',
              'tick',
              'createEventDispatcher'
            ],
            'svelte/store': ['writable', 'readable', 'derived', 'get'],
            'svelte-sonner': ['toast']
          }
        ],
        // 生成类型定义文件
        dts: './typings/auto-imports.d.ts',
        // 包含的文件类型
        include: [/\.[tj]sx?$/, /\.svelte$/],
        eslintrc: {
          enabled: true,
          filepath: './.eslintrc-auto-import.js',
          globalsPropValue: true
        }
      }),
      paraglideVitePlugin({
        project: './project.inlang',
        outdir: './src/lib/paraglide'
      })
    ],
    server: {
      port: Number(env.VITE_SERVER_PORT),
      host: '0.0.0.0',
      proxy: {
        [env.VITE_BASE_URL as string]: {
          target: env.VITE_PROXY_URL,
          ws: true,
          changeOrigin: true
          // rewrite: (path: string) => path.replace(new RegExp(`^${env.VITE_BASE_URL}`), '')
        }
      }
    },
    css: {
      postcss: './postcss.config.js'
    },
    resolve: {
      alias: {
        $lib: fileURLToPath(new URL('./src/lib', import.meta.url))
      }
    }
  })
}
