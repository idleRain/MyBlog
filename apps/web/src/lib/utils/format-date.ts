// ISO 日期截取的字符长度，对应 YYYY-MM-DD 的目录标注格式。
const ISO_DATE_LENGTH = 10

/**
 * 将 ISO 时间字符串格式化为 YYYY-MM-DD 的展示格式。
 * 直接截取原始字符串，避免时区换算导致日期偏移。
 */
export function formatDate(iso: string): string {
  return iso.slice(0, ISO_DATE_LENGTH)
}
