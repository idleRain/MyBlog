import basePrettierConfig from '../prettier.config.js'

/** @type {import('prettier').Config} */
const webPrettierConfig = {
  ...basePrettierConfig,
  // Svelte 特定配置
  svelteStrictMode: false,
  svelteAllowShorthand: true,
  svelteIndentScriptAndStyle: false,
  // 插件
  plugins: [
    'prettier-plugin-svelte',
    'prettier-plugin-tailwindcss',
    'prettier-plugin-css-order',
    'prettier-plugin-sort-imports'
  ],
  // 针对 Svelte 文件的特殊配置
  overrides: [
    {
      files: '*.svelte',
      options: {
        parser: 'svelte'
      }
    }
  ],
  // Tailwind 配置
  tailwindConfig: './tailwind.config.js',
  tailwindStylesheet: './src/app.css',
  tailwindFunctions: ['clsx']
}

export default webPrettierConfig
