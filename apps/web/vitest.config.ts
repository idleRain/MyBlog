import { svelte } from '@sveltejs/vite-plugin-svelte'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vitest/config'

// 前台此前零测试基建，本配置只为单元测试服务，不参与生产构建。
// 被测模块含响应式状态与 markdown 渲染管线，均需经 svelte 插件编译；
// 显式声明 runes 模式以确保 .svelte.ts 内的 $state 被正确处理。
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
      // 与应用源码保持同一套别名，避免测试走另一条解析路径。
      $lib: fileURLToPath(new URL('./src/lib', import.meta.url)),
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      $ui: fileURLToPath(new URL('../../packages/ui/src', import.meta.url))
    }
  },
  test: {
    // 当前被测面为纯逻辑与文本渲染，Node 环境足够，不引入浏览器模拟依赖。
    environment: 'node',
    include: ['src/**/*.test.ts'],
    // markdown 渲染需按需加载语言包，放宽单个用例的超时以便冷启动完成。
    testTimeout: 30000
  }
})
