// #fonts 字体来源模块的类型声明。
// 实际解析目标由 vite 在构建期按模式切换：开发指向 fonts-local.ts（本地 fontsource 副作用），
// 生产指向 fonts-cdn.ts（jsDelivr 样式表链接）。类型检查以 ambient 声明为准，
// 不经 tsconfig paths 挂载，避免浅合并覆盖 .svelte-kit 生成的既有别名配置。
declare module '#fonts' {
  // 供布局注入的样式表链接；本地模式下为空数组，字体经模块副作用注册。
  export const fontCssLinks: string[]
}
