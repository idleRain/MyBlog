// toast-ui-editor 的类型声明补全：其 package.json 的 exports 未声明 types 条件，
// 在 bundler 模块解析下无法定位到 types/index.d.ts，此处通过模块增强重新导出。

declare module '@toast-ui/editor' {
  export * from '@toast-ui/editor/types'
  export { default } from '@toast-ui/editor/types'
}

// 编辑器语言包为副作用导入，无类型声明，声明为空模块以避免报错。
declare module '@toast-ui/editor/dist/i18n/zh-cn'
