// 契约锁 ②：TS 侧锚定契约金样本与手写类型的结构一致性。
// JSON 导入值的枚举与字面量字段在类型推导中放宽为 string/number，
// 无法与手写联合类型直接比对，故将手写类型经 WidenLiteral 做同样放宽后
// 与金样本做精确相等断言：金样本或手写类型任一方增删字段、改变字段类型即编译失败。
// 本文件经 packages/api 的 typecheck 脚本挂入 contract:check，不再是无自动化覆盖的编译期断言。
import { describe, expect, expectTypeOf, it } from 'vitest'
import type { LoginData, UserListData } from '../modules/user/types'
import loginSuccess from '../../../../contracts/fixtures/login.success.json'
import loginWrongPassword from '../../../../contracts/fixtures/login.wrong-password.json'
import usersList from '../../../../contracts/fixtures/users.list.json'

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
    expectTypeOf(loginSuccess.data).toEqualTypeOf<WidenLiteral<LoginData>>()
    // 运行期兜底：确保金样本确含登录数据字段。
    expect(loginSuccess.data.accessToken).toBeTypeOf('string')
  })

  it('登录失败响应为无 data 字段的错误信封', () => {
    // 后端错误响应省略 data 字段，金样本只声明 code 与 message。
    expectTypeOf(loginWrongPassword).toEqualTypeOf<
      WidenLiteral<{ code: number; message: string }>
    >()
  })

  it('用户列表响应的 data 与 UserListData 结构精确一致', () => {
    expectTypeOf(usersList.data).toEqualTypeOf<WidenLiteral<UserListData>>()
  })
})