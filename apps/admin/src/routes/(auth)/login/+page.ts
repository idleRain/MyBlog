import { zod4 } from 'sveltekit-superforms/adapters'
import { superValidate } from 'sveltekit-superforms'
import { z } from 'zod'

const loginSchema = z.object({
  username: z.string().min(1, '请输入用户名'),
  password: z.string().min(1, '请输入密码')
})

export const load = async () => {
  // 客户端渲染模式下在浏览器端生成默认表单数据，不依赖服务端渲染。
  const form = await superValidate(zod4(loginSchema))
  return { form }
}
