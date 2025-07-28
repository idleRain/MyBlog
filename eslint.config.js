import js from '@eslint/js'
import ts from 'typescript-eslint'
import prettier from 'eslint-config-prettier'
import prettierPlugin from 'eslint-plugin-prettier/recommended'
import globals from 'globals'

// 基础 ESLint 配置，可被各个子项目继承
export const baseConfig = [
  js.configs.recommended,
  ...ts.configs.recommended,
  prettier,
  prettierPlugin,
  {
    languageOptions: {
      globals: { ...globals.node }
    },
    rules: {
      // TypeScript 项目不需要 no-undef 规则
      'no-undef': 'off',
      // 允许使用控制字符
      'no-control-regex': 'off',
      // 启用 Prettier 格式检查
      'prettier/prettier': 'warn',
      // 允许使用 any 类型
      '@typescript-eslint/no-explicit-any': 'off',
      // 未使用变量警告而非错误
      '@typescript-eslint/no-unused-vars': 'off',
      // 允许使用 this 别名
      '@typescript-eslint/no-this-alias': [
        'error',
        {
          allowedNames: ['that']
        }
      ]
    }
  }
]

// 默认导出根目录配置
export default [
  ...baseConfig,
  {
    files: ['scripts/**/*.js', '*.config.js'],
    languageOptions: {
      globals: { ...globals.node }
    }
  }
]
