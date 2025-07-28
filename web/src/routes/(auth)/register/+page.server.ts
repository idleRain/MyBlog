import type { Actions, PageServerLoad } from './$types'
import { superValidate } from 'sveltekit-superforms'
import { zod } from 'sveltekit-superforms/adapters'
import { fail, redirect } from '@sveltejs/kit'
import { z } from 'zod'

const registerSchema = z
  .object({
    username: z
      .string()
      .min(3, '用户名至少需要3个字符')
      .max(20, '用户名不能超过20个字符')
      .regex(/^[a-zA-Z0-9_]+$/, '用户名只能包含字母、数字和下划线'),
    email: z.string().email('请输入有效的邮箱地址'),
    password: z.string().min(6, '密码至少需要6个字符').max(50, '密码不能超过50个字符'),
    confirmPassword: z.string()
  })
  .refine(data => data.password === data.confirmPassword, {
    message: '两次输入的密码不一致',
    path: ['confirmPassword']
  })

// @ts-ignore
export const load: PageServerLoad = async () => {
  throw redirect(302, '/login')
  const form = await superValidate(zod(registerSchema))
  return { form }
}

export const actions: Actions = {
  default: async event => {
    const form = await superValidate(event, zod(registerSchema))

    if (!form.valid) {
      return fail(400, {
        form
      })
    }

    // 这里可以添加实际的注册逻辑
    // 现在只是返回成功，客户端会处理实际的 API 调用
    return {
      form
    }
  }
}
