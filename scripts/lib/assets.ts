// 构建产物度量公共库：按扩展名分类统计产物体积，供字体门禁与体积度量脚本共用。
// 度量口径集中在此处，避免两个脚本各写一份扫描逻辑而逐渐分叉。

import { readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'

// 产物分类标识，用于汇总表中的行标签。
// sourcemap 单列而不并入 other：源映射体积随源码规模自然增长，不是可优化项，
// 若混入其他分类会让体积提醒指向无行动价值的文件。
export type AssetKind = 'font' | 'javascript' | 'stylesheet' | 'sourcemap' | 'other'

// 各分类对应的文件扩展名，小写比较。
// sourcemap 必须排在 javascript 与 stylesheet 之前判定，否则 .js.map 会被当作脚本。
const EXTENSIONS_BY_KIND: Record<AssetKind, string[]> = {
  font: ['.woff', '.woff2', '.ttf', '.otf', '.eot'],
  sourcemap: ['.map'],
  javascript: ['.js', '.mjs', '.cjs'],
  stylesheet: ['.css'],
  other: []
}

// 汇总表输出顺序，按本项目的关注度排列。
export const ASSET_KINDS: AssetKind[] = ['font', 'javascript', 'stylesheet', 'sourcemap', 'other']

// 单个文件的路径与字节数。
export interface AssetFile {
  path: string
  bytes: number
  kind: AssetKind
}

// 按分类汇总的体积结果。
export interface AssetSummary {
  files: AssetFile[]
  bytesByKind: Record<AssetKind, number>
  countByKind: Record<AssetKind, number>
}

/**
 * 判定文件所属分类。
 * 依据扩展名匹配，未命中任何类别时归入 other，因此资产分类不会漏项。
 */
export function classifyAsset(fileName: string): AssetKind {
  const lowerName = fileName.toLowerCase()
  for (const kind of ASSET_KINDS) {
    if (EXTENSIONS_BY_KIND[kind].some(extension => lowerName.endsWith(extension))) {
      return kind
    }
  }
  return 'other'
}

/**
 * 递归收集目录下的全部文件。
 * 目录不存在时返回空列表，使调用方在产物缺失时给出可读结论而非异常堆栈。
 */
export function collectAssets(directory: string): AssetFile[] {
  const assets: AssetFile[] = []

  const walk = (currentDirectory: string): void => {
    let entries
    try {
      entries = readdirSync(currentDirectory, { withFileTypes: true })
    } catch (error) {
      if (isMissingPathError(error)) {
        return
      }
      throw error
    }

    for (const entry of entries) {
      const entryPath = join(currentDirectory, entry.name)
      if (entry.isDirectory()) {
        walk(entryPath)
        continue
      }
      assets.push({
        path: entryPath,
        bytes: statSync(entryPath).size,
        kind: classifyAsset(entry.name)
      })
    }
  }

  walk(directory)
  return assets
}

/**
 * 判定错误是否为路径不存在。
 * 产物目录被清理后度量脚本应给出提示而非报错，因此需要区分该情形与其他读取失败。
 */
function isMissingPathError(error: unknown): boolean {
  if (!(error instanceof Error)) {
    return false
  }
  const code = (error as NodeJS.ErrnoException).code
  return code === 'ENOENT' || code === 'ENOTDIR'
}

/**
 * 汇总资产列表为按分类的字节数与文件数。
 */
export function summarizeAssets(assets: AssetFile[]): AssetSummary {
  const bytesByKind = emptyKindRecord()
  const countByKind = emptyKindRecord()

  for (const asset of assets) {
    bytesByKind[asset.kind] += asset.bytes
    countByKind[asset.kind] += 1
  }

  return { files: assets, bytesByKind, countByKind }
}

// 生成按分类初始化为零的记录。
function emptyKindRecord(): Record<AssetKind, number> {
  return ASSET_KINDS.reduce(
    (record, kind) => {
      record[kind] = 0
      return record
    },
    {} as Record<AssetKind, number>
  )
}

/**
 * 以人类可读单位输出字节数，保留两位小数。
 * 体积对比是本脚本的主要用途，因此统一到 MB 便于跨次构建比较。
 */
export function formatBytes(bytes: number): string {
  const kilobytes = 1024
  const megabytes = kilobytes * 1024

  if (bytes >= megabytes) {
    return `${(bytes / megabytes).toFixed(2)} MB`
  }
  if (bytes >= kilobytes) {
    return `${(bytes / kilobytes).toFixed(2)} KB`
  }
  return `${bytes} B`
}
