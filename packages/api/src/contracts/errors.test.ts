// 契约锁：@myblog/shared 的传输层响应码与后端 pkg/response 逐值一致。
// 后端常量是运行期唯一权威，前端常量是消费侧镜像；两侧命名习惯不同，
// 故由本文件的对应表登记翻译关系，任一侧新增或改值而另一侧未跟进即在此变红。
// 直接解析两侧源码而非反向导入：前端包无需为读取后端 Go 常量新增依赖。
import { describe, expect, it } from 'vitest'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { readFileSync } from 'node:fs'

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), '../../../..')
const goResponsePath = resolve(repositoryRoot, 'server/pkg/response/response.go')
const sharedConstantsPath = resolve(repositoryRoot, 'packages/shared/src/constants/api.ts')

// Go 常量形如 `CodeSuccess  = 200 // 成功`，捕获组为完整常量名与响应码。
const goCodePattern = /^\s*(Code\w+)\s*=\s*(\d+)/gm
// 前端常量形如 `export const RESPONSE_CODE_SUCCESS = 200`，捕获组为常量名后缀与响应码。
const sharedCodePattern = /export const RESPONSE_CODE_(\w+)\s*=\s*(\d+)/g

/**
 * 后端常量名到前端常量名的对应表，两侧命名习惯不同故显式登记。
 * 后端新增响应码而未在此登记时，比对会失败并提示缺少对应关系。
 */
const responseCodeNames: Record<string, string> = {
  CodeSuccess: 'SUCCESS',
  CodeError: 'ERROR',
  CodeInvalid: 'INVALID',
  CodeAuth: 'UNAUTHORIZED',
  CodeForbid: 'FORBIDDEN',
  CodeNotFound: 'NOT_FOUND'
}

/** 按正则抽取源码中的常量名与取值，返回以常量名为键的响应码映射。 */
function collectCodes(sourcePath: string, pattern: RegExp): Map<string, number> {
  const codes = new Map<string, number>()
  for (const match of readFileSync(sourcePath, 'utf-8').matchAll(pattern)) {
    codes.set(match[1], Number(match[2]))
  }
  return codes
}

describe('错误码契约锚定', () => {
  const goCodes = collectCodes(goResponsePath, goCodePattern)
  const sharedCodes = collectCodes(sharedConstantsPath, sharedCodePattern)

  it('两侧都解析出了传输层响应码', () => {
    expect(goCodes.size).toBeGreaterThan(0)
    expect(sharedCodes.size).toBeGreaterThan(0)
  })

  it('后端每个响应码都有同值的前端常量', () => {
    for (const [goName, goValue] of goCodes) {
      const sharedName = responseCodeNames[goName]
      expect(sharedName, `后端 ${goName} 未登记前端常量对应关系`).toBeTypeOf('string')
      expect(
        sharedCodes.get(sharedName),
        `前端 RESPONSE_CODE_${sharedName} 与后端 ${goName} 不一致`
      ).toBe(goValue)
    }
  })

  it('前端不存在后端未定义的响应码常量', () => {
    const declared = new Set(Object.values(responseCodeNames))
    for (const sharedName of sharedCodes.keys()) {
      expect(declared.has(sharedName), `前端 RESPONSE_CODE_${sharedName} 在后端没有对应常量`).toBe(
        true
      )
    }
  })

  it('两侧响应码取值集合一致', () => {
    const sortAscending = (left: number, right: number) => left - right
    expect([...sharedCodes.values()].sort(sortAscending)).toEqual(
      [...goCodes.values()].sort(sortAscending)
    )
  })
})
