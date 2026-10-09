import { svelte } from '@sveltejs/vite-plugin-svelte'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vitest/config'

// 后台应用此前零测试基建，本配置只为单元测试服务，不参与生产构建。
// 页面状态模块以 .svelte.ts 承载 runes，因此必须经 svelte 插件编译，
// 且需启用 runes 模式，否则模块内的 $state 不会被编译器识别。
export default defineConfig({
  plugins: [
    svelte({
      compilerOptions: {
        runes: true
      }
    })
  ],
  resolve: {
    alias: {
      // 与被测模块和应用源码保持同一套别名，避免测试走另一条解析路径。
      $lib: fileURLToPath(new URL('./src/lib', import.meta.url)),
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      $ui: fileURLToPath(new URL('../../packages/ui/src', import.meta.url))
    }
  },
  test: {
    // 页面状态模块不触碰 DOM，纯 Node 环境即可运行，避免引入额外的浏览器模拟依赖。
    environment: 'node',
    include: ['src/**/*.test.ts']
  }
})
