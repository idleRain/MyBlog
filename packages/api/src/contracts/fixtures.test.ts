// 契约锁 ②：TS 侧锚定契约金样本为手写契约类型的合法实例。
// JSON 导入值的字面量与枚举字段在类型推导中放宽为 string/number，
// 手写类型经 WidenLiteral 做同样放宽后，以 toExtend 单向断言金样本可赋给契约类型：
// 金样本缺字段或字段类型不符、手写类型收窄时，tsc 类型检查即失败。
// 本文件经 packages/api 的 typecheck 脚本挂入 contract:check，不再是无自动化覆盖的编译期断言。
// 每个金样本都须在此有锚点，Go 侧 TestFixtureCoverage 会校验本文件的引用清单与目录一致。
import loginWrongPassword from '../../../../contracts/fixtures/login.wrong-password.json'
import type { LoginData, SessionData, UserListData } from '../modules/user/types'
import sessionCreated from '../../../../contracts/fixtures/session.created.json'
import logoutSuccess from '../../../../contracts/fixtures/logout.success.json'
import articleDetail from '../../../../contracts/fixtures/article.detail.json'
import loginSuccess from '../../../../contracts/fixtures/login.success.json'
import usersList from '../../../../contracts/fixtures/users.list.json'
import { describe, expect, expectTypeOf, it } from 'vitest'
import type { Article } from '../modules/article/types'

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

// 后端成功响应对 nil 数据省略 data 键，错误响应同样不输出 data，故此类金样本只锚定信封。
type EnvelopeWithoutData = { code: number; message: string }

describe('契约金样本类型锚定', () => {
  it('登录成功响应的 data 与 LoginData 结构精确一致', () => {
    expectTypeOf(loginSuccess.data).toExtend<WidenLiteral<LoginData>>()
    // 运行期校验：确保金样本确含登录数据字段。
    expect(loginSuccess.data.accessToken).toBeTypeOf('string')
  })

  it('登录失败响应为无 data 字段的错误信封', () => {
    // 后端错误响应省略 data 字段，金样本只声明 code 与 message。
    expectTypeOf(loginWrongPassword).toExtend<WidenLiteral<EnvelopeWithoutData>>()
  })

  it('用户列表响应的 data 与 UserListData 结构精确一致', () => {
    expectTypeOf(usersList.data).toExtend<WidenLiteral<UserListData>>()
  })

  it('文章详情响应的 data 与 Article 结构精确一致', () => {
    expectTypeOf(articleDetail.data).toExtend<WidenLiteral<Article>>()
  })

  it('会话建立响应的 data 与 SessionData 结构精确一致', () => {
    expectTypeOf(sessionCreated.data).toExtend<WidenLiteral<SessionData>>()
    // 令牌对经 Set-Cookie 下发，响应体不得回传令牌本体。
    expect(sessionCreated.data).not.toHaveProperty('accessToken')
  })

  it('登出成功响应为无 data 字段的信封', () => {
    // 空数据成功响应与错误响应同形，锁定 data 键的省略语义。
    expectTypeOf(logoutSuccess).toExtend<WidenLiteral<EnvelopeWithoutData>>()
    expect(logoutSuccess).not.toHaveProperty('data')
  })
})
