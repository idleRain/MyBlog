import js from '@eslint/js'
import ts from 'typescript-eslint'
import prettier from 'eslint-config-prettier'
import globals from 'globals'

// 基础 ESLint 配置，可被各个子项目继承
// 职责分离：ESLint 只管代码质量，格式化统一由 prettier --write 承担，
// 不在 lint 中通过 eslint-plugin-prettier 双跑格式校验。
export const baseConfig = [
  js.configs.recommended,
  ...ts.configs.recommended,
  // 关闭与 prettier 冲突的风格规则，格式交由独立 format 步骤。
  prettier,
  {
    languageOptions: {
      globals: { ...globals.node }
    },
    rules: {
      // TypeScript 项目不需要 no-undef 规则
      'no-undef': 'off',
      // 允许使用控制字符
      'no-control-regex': 'off',
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
  },
  {
    // ui 包为 stock shadcn 第三方生成代码，不参与 lint。
    ignores: ['packages/ui/**']
  }
]
