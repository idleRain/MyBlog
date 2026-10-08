// 本地字体来源：仅由 vite 在非生产模式下解析到本模块（#fonts 别名）。
// fontsource 的 CSS 自带 .woff 与 .woff2 双格式声明，仅服务于开发环境，
// 生产产物经 #fonts 别名切换为 CDN 链接，本地字体不进入 build 产物。

import '@fontsource/noto-serif-sc/chinese-simplified-900.css'
import '@fontsource/noto-serif-sc/chinese-simplified-700.css'
import '@fontsource/noto-serif-sc/chinese-simplified-500.css'
import '@fontsource/noto-sans-sc/chinese-simplified-700.css'
import '@fontsource/noto-sans-sc/chinese-simplified-500.css'
import '@fontsource/noto-sans-sc/chinese-simplified-400.css'
import '@fontsource/fira-mono/latin-700.css'
import '@fontsource/fira-mono/latin-500.css'
import '@fontsource/fira-mono/latin-400.css'

// 本地模式经模块副作用注册 @font-face，无需额外的链接注入。
export const fontCssLinks: string[] = []
