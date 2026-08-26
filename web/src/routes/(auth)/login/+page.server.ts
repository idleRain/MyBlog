import type { Actions, PageServerLoad } from './$types'
import { zod4 } from 'sveltekit-superforms/adapters'
import { superValidate } from 'sveltekit-superforms'
import { fail } from '@sveltejs/kit'
import { z } from 'zod'

const loginSchema = z.object({
  username: z.string().min(1, '请输入用户名'),
  password: z.string().min(1, '请输入密码')
})

export const load: PageServerLoad = async () => {
  const form = await superValidate(zod4(loginSchema))
  return { form }
}

export const actions: Actions = {
  default: async event => {
    const form = await superValidate(event, zod4(loginSchema))

    if (!form.valid) {
      return fail(400, {
        form
      })
    }

    // 这里可以添加实际的登录逻辑
    // 现在只是返回成功，客户端会处理实际的 API 调用
    return {
      form
    }
  }
}
