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
      'svelte/css-unused-selector': 'off'
    }
  },
  {
    // Ts 特定规则
    rules: {
      '@typescript-eslint/ban-ts-comment': 'off'
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
