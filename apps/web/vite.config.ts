import { type ConfigEnv, type PluginOption, defineConfig, loadEnv } from 'vite'
import { paraglideVitePlugin } from '@inlang/paraglide-js'
import devtoolsJson from 'vite-plugin-devtools-json'
import AutoImport from 'unplugin-auto-import/vite'
import { sveltekit } from '@sveltejs/kit/vite'
import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import ViteJson5 from 'vite-plugin-json5'

// 环境变量缺失时的本地开发缺省值，与 apps/web/.env 的默认配置保持一致。
const DEFAULT_DEV_PORT = 8899
const DEFAULT_BASE_URL = '/api'
const DEFAULT_PROXY_TARGET = 'http://localhost:3000'

export default ({ mode }: ConfigEnv) => {
  const env = loadEnv(mode, process.cwd())
  // json5 插件自带嵌套的 vite 类型声明，与根版本存在 exactOptionalPropertyTypes 差异。
  // 两个插件类型来自不同的 vite 实例，先经 unknown 中转再收束为当前 vite 的插件类型。
  const json5Plugin = ViteJson5() as unknown as PluginOption

  return defineConfig({
    plugins: [
      tailwindcss(),
      sveltekit(),
      devtoolsJson(),
      json5Plugin,
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
      port: Number(env.VITE_SERVER_PORT) || DEFAULT_DEV_PORT,
      host: '0.0.0.0',
      proxy: {
        [env.VITE_BASE_URL ?? DEFAULT_BASE_URL]: {
          target: env.VITE_PROXY_URL ?? DEFAULT_PROXY_TARGET,
          ws: true,
          changeOrigin: true
          // rewrite: (path: string) => path.replace(new RegExp(`^${env.VITE_BASE_URL}`), '')
        }
      }
    },
    ssr: {
      // ui 包以源码直连方式被引用，需参与 SSR 编译以支持 .svelte 组件。
      noExternal: ['@myblog/ui']
    },
    resolve: {
      alias: {
        $lib: fileURLToPath(new URL('./src/lib', import.meta.url))
      }
    }
  })
}
