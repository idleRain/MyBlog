/** @type {import('prettier').Config} */
const basePrettierConfig = {
  // 一行最多字符数
  printWidth: 100,
  // 缩进空格数
  tabWidth: 2,
  // 使用空格而非制表符
  useTabs: false,
  // 语句末尾添加分号
  semi: false,
  // 使用单引号
  singleQuote: true,
  // 对象属性引号策略
  quoteProps: 'as-needed',
  // JSX 中使用双引号
  jsxSingleQuote: false,
  // 尾随逗号策略
  trailingComma: 'none',
  // 对象括号内空格
  bracketSpacing: true,
  // 箭头函数参数括号
  arrowParens: 'avoid',
  // 格式化范围
  rangeStart: 0,
  rangeEnd: Infinity,
  // 不需要 pragma 注释
  requirePragma: false,
  insertPragma: false,
  // 换行策略
  proseWrap: 'preserve',
  // HTML 空格敏感度
  htmlWhitespaceSensitivity: 'css',
  // 换行符类型
  endOfLine: 'auto'
}

export default basePrettierConfig