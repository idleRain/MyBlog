// 设置模块常量：分组标题映射与默认分组名。

// 设置分组中文标题映射，未知分组回退显示原始分组名。
export const SETTING_GROUP_LABELS: Record<string, string> = {
  general: '基础设置',
  seo: 'SEO 设置',
  content: '内容设置',
  media: '媒体设置',
  mail: '邮件设置',
  security: '安全设置',
  cache: '缓存设置',
  third_party: '第三方集成'
}

// 敏感设置的掩码值，与后端脱敏输出保持一致。
export const SENSITIVE_MASK = '********'

// 后端业务逻辑实际消费的设置键，来源以 server/internal/service 全量检索为准；
// 新键接入业务后须同步补录此处，未在此列的键仅为存储值，不产生任何运行时效果。
export const EFFECTIVE_SETTING_KEYS: ReadonlySet<string> = new Set([
  'allow_guest_comment',
  'comment_auto_approve'
])

// 判断设置键是否已被业务消费，未消费项在界面标注"未生效"。
export function isSettingEffective(keyName: string): boolean {
  return EFFECTIVE_SETTING_KEYS.has(keyName)
}
