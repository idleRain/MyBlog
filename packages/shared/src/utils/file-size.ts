/**
 * 获取可读的文件大小字符串，字符串入参原样返回。
 */
export const getFileSize = (value?: number | string) => {
  if (typeof value === 'string') return value
  if (value == null || value === 0) return '0 Bytes'

  const units = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  const index = Math.floor(Math.log(value) / Math.log(1024))
  const size = value / 1024 ** index
  const formattedSize = Number(size.toFixed(2))

  return `${formattedSize} ${units[index]}`
}
