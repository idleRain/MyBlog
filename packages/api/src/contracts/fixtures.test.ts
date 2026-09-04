// 契约锁 ②：TS 侧断言契约金样本与手写类型双向精确相等。
// 金样本由后端 handler 测试（锁①）驱动更新，本测试确保前端类型与金样本任一方漂移即编译失败。
import { describe, expect, expectTypeOf, it } from 'vitest'
import type { LoginData, UserListData } from '../modules/user/types'
import loginSuccess from '../../../../contracts/fixtures/login.success.json'
import loginWrongPassword from '../../../../contracts/fixtures/login.wrong-password.json'
import usersList from '../../../../contracts/fixtures/users.list.json'

describe('契约金样本类型锚定', () => {
  it('登录成功响应的 data 与 LoginData 类型精确一致', () => {
    expectTypeOf(loginSuccess.data).toEqualTypeOf<LoginData>()
    // 运行期兜底：确保金样本确含登录数据字段。
    expect(loginSuccess.data.accessToken).toBeTypeOf('string')
  })

  it('登录失败响应为无 data 字段的错误信封', () => {
    // 后端错误响应省略 data 字段，金样本只声明 code 与 message。
    expectTypeOf(loginWrongPassword).toEqualTypeOf<{ code: number; message: string }>()
  })

  it('用户列表响应的 data 与 UserListData 类型精确一致', () => {
    expectTypeOf(usersList.data).toEqualTypeOf<UserListData>()
  })
})
