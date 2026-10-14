import basePrettierConfig from '../../prettier.config.js'

// 包内含 .svelte 组件，需在基础配置上补 Svelte 解析器与该语法的既定排版口径。
// 口径与两个应用的 prettier 配置保持一致，避免同一组件在不同目录下排版不同。
/** @type {import('prettier').Config} */
const componentsPrettierConfig = {
  ...basePrettierConfig,
  svelteAllowShorthand: true,
  svelteIndentScriptAndStyle: false,
  plugins: ['prettier-plugin-svelte', 'prettier-plugin-css-order', 'prettier-plugin-sort-imports'],
  overrides: [
    {
      files: '*.svelte',
      options: {
        parser: 'svelte'
      }
    }
  ]
}

export default componentsPrettierConfig
