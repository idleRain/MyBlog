// 契约锁 ②：TS 侧锚定契约金样本为手写契约类型的合法实例。
// JSON 导入值的字面量与枚举字段在类型推导中放宽为 string/number，
// 手写类型经 WidenLiteral 做同样放宽后，以 toExtend 单向断言金样本可赋给契约类型：
// 金样本缺字段或字段类型不符、手写类型收窄时，tsc 类型检查即失败。
// 本文件经 packages/api 的 typecheck 脚本挂入 contract:check，不再是无自动化覆盖的编译期断言。
import { describe, expect, expectTypeOf, it } from 'vitest'
import type { LoginData, UserListData } from '../modules/user/types'
import type { Article } from '../modules/article/types'
import loginSuccess from '../../../../contracts/fixtures/login.success.json'
import loginWrongPassword from '../../../../contracts/fixtures/login.wrong-password.json'
import usersList from '../../../../contracts/fixtures/users.list.json'
import articleDetail from '../../../../contracts/fixtures/article.detail.json'

// 将字面量与联合类型放宽为对应基础类型，用于与 JSON 导入值的推导类型做精确比对。
type WidenLiteral<T> = T extends string
  ? string
  : T extends number
    ? number
    : T extends boolean
      ? boolean
      : T extends readonly (infer U)[]
        ? WidenLiteral<U>[]
        : T extends object
          ? { [K in keyof T]: WidenLiteral<T[K]> }
          : T

describe('契约金样本类型锚定', () => {
  it('登录成功响应的 data 与 LoginData 结构精确一致', () => {
    expectTypeOf(loginSuccess.data).toExtend<WidenLiteral<LoginData>>()
    // 运行期兜底：确保金样本确含登录数据字段。
    expect(loginSuccess.data.accessToken).toBeTypeOf('string')
  })

  it('登录失败响应为无 data 字段的错误信封', () => {
    // 后端错误响应省略 data 字段，金样本只声明 code 与 message。
    expectTypeOf(loginWrongPassword).toExtend<WidenLiteral<{ code: number; message: string }>>()
  })

  it('用户列表响应的 data 与 UserListData 结构精确一致', () => {
    expectTypeOf(usersList.data).toExtend<WidenLiteral<UserListData>>()
  })

  it('文章详情响应的 data 与 Article 结构精确一致', () => {
    expectTypeOf(articleDetail.data).toExtend<WidenLiteral<Article>>()
  })
})