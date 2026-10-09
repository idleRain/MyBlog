import autoImportGlobals from './.eslintrc-auto-import.js'
import { baseConfig } from '../../eslint.config.js'
import { includeIgnoreFile } from '@eslint/compat'
import svelteConfig from './svelte.config.js'
import svelte from 'eslint-plugin-svelte'
import { fileURLToPath } from 'node:url'
import ts from 'typescript-eslint'
import globals from 'globals'

const gitignorePath = fileURLToPath(new URL('./.gitignore', import.meta.url))

export default ts.config(
  includeIgnoreFile(gitignorePath),
  ...baseConfig,
  ...svelte.configs.recommended,
  ...svelte.configs.prettier,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
        // 自动导入的全局变量
        ...autoImportGlobals.globals
      }
    }
  },
  {
    // Svelte 特定规则
    rules: {
      'svelte/html-self-closing': 'warn',
      'svelte/spaced-html-comment': 'warn',
      'svelte/valid-prop-names-in-kit-pages': 'off',
      'svelte/css-unused-selector': 'off',
      // 项目既有导航调用未使用 resolve 包装，新版本插件默认启用该规则导致既有代码报错。
      'svelte/no-navigation-without-resolve': 'off'
    }
  },
  {
    // Ts 特定规则
    rules: {
      '@typescript-eslint/ban-ts-comment': 'off',
      // 空对象接口沿用项目既有写法，新版本插件默认启用导致既有代码报错。
      '@typescript-eslint/no-empty-object-type': 'off'
    }
  },
  {
    // 导入守门：拦截影子类型层与组件库桶文件的重新引入。
    rules: {
      'no-restricted-imports': [
        'error',
        {
          // 桶文件导入会把整个组件库拉进类型检查与打包范围，组件一律走子路径。
          paths: [
            {
              name: '$ui',
              message:
                '$ui 桶文件会一次性引入全部组件，显著扩大类型检查范围，请按组件路径导入，例如 $ui/button。'
            },
            {
              name: '@myblog/ui',
              message:
                '@myblog/ui 桶文件会一次性引入全部组件，显著扩大类型检查范围，请按组件路径导入，例如 @myblog/ui/button。'
            }
          ],
          patterns: [
            {
              group: ['$lib/types', '$lib/types/*', '$lib/types/**'],
              message:
                '应用层影子类型目录已删除，接口类型一律来自 @myblog/api/modules/*/types，禁止重新引入。'
            }
          ]
        }
      ]
    }
  },
  {
    files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
    languageOptions: {
      parserOptions: {
        extraFileExtensions: ['.svelte'],
        parser: ts.parser,
        svelteConfig
      }
    }
  }
)
