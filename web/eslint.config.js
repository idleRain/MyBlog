import autoImportGlobals from './.eslintrc-auto-import.js'
import { includeIgnoreFile } from '@eslint/compat'
import { baseConfig } from '../eslint.config.js'
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
    files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
    languageOptions: {
      parserOptions: {
        projectService: true,
        extraFileExtensions: ['.svelte'],
        parser: ts.parser,
        svelteConfig
      }
    }
  },
  {
    ignores: ['src/lib/components/ui/**', 'src/lib/paraglide/**']
  }
)
