// 脚本终端输出公共库：ANSI 转义码常量与着色函数，供 scripts 下各脚本统一引用。

// ANSI 转义码常量，供模板字符串拼接着色片段使用。
export const ansi = {
  blue: '\x1b[34m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  red: '\x1b[31m',
  cyan: '\x1b[36m',
  reset: '\x1b[0m',
  bold: '\x1b[1m'
} as const

// 着色函数类型：接收文本并输出带色彩前缀与复位后缀的片段。
type Painter = (text: string) => string

// 由色彩码构造着色函数。
function painter(code: string): Painter {
  return (text: string) => `${code}${text}${ansi.reset}`
}

// 着色函数集合，覆盖脚本中常用的单段彩色输出场景。
export const colors: {
  blue: Painter
  green: Painter
  yellow: Painter
  red: Painter
  cyan: Painter
  bold: Painter
} = {
  blue: painter(ansi.blue),
  green: painter(ansi.green),
  yellow: painter(ansi.yellow),
  red: painter(ansi.red),
  cyan: painter(ansi.cyan),
  bold: painter(ansi.bold)
}
