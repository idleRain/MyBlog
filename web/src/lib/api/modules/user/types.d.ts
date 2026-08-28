// 兼容过渡层：web 侧类型从 @myblog/api 统一再导出。
// Phase 4 拆分应用后，各应用直接使用 @myblog/api 类型，本文件随 web 移除。

export * from '@myblog/api/modules/user/types'
