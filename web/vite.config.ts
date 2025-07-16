import { type ConfigEnv, defineConfig, loadEnv } from 'vite'
import devtoolsJson from 'vite-plugin-devtools-json'
import { sveltekit } from '@sveltejs/kit/vite'
import tailwindcss from '@tailwindcss/vite'

export default ({ mode }: ConfigEnv) => {
  const env = loadEnv(mode, process.cwd())

  return defineConfig({
    plugins: [tailwindcss(), sveltekit(), devtoolsJson()],
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
    }
  })
}
