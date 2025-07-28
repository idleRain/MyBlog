<script lang="ts">
import { onMount } from 'svelte'
import { authStore } from '$lib/stores/auth'
import { Button } from '$ui'
import { UserAPI } from '$lib/api'
import { safeApiCall } from '$lib/utils/request'
import { performLogout } from '$lib/utils/logout'
import { toast } from 'svelte-sonner'

let authState = $state(null)
let testResult = $state('')

onMount(() => {
  authStore.subscribe(state => {
    authState = state
  })
})

async function testTokenRefresh() {
  testResult = '测试中...'

  // 尝试调用一个需要认证的 API
  const { data, error, success } = await safeApiCall(() => UserAPI.getUserList(1, 10), {
    showErrorToast: false
  })

  if (success) {
    testResult = 'Token 有效，API 调用成功'
    toast.success('API 调用成功')
  } else {
    testResult = `API 调用失败: ${error?.message || '未知错误'}`
    toast.error('API 调用失败')
  }
}

async function testInvalidToken() {
  testResult = '测试无效 token...'

  // 修改 localStorage 中的 token 为无效值
  localStorage.setItem('auth_access_token', 'invalid_token_test')

  // 尝试调用 API
  const { data, error, success } = await safeApiCall(() => UserAPI.getUserList(1, 10), {
    showErrorToast: false
  })

  if (success) {
    testResult = '意外成功 - 这不应该发生'
  } else {
    testResult = `预期失败: ${error?.message || '401 错误'}`
    toast.info('测试完成 - 无效 token 被正确处理')
  }
}

async function clearAuth() {
  testResult = '正在退出登录...'

  const success = await performLogout({
    showToast: false, // 我们手动控制提示
    redirectTo: null // 不自动跳转，继续显示测试页面
  })

  if (success) {
    testResult = '成功退出登录（已调用后端接口使 token 失效）'
    toast.success('已成功退出登录')
  } else {
    testResult = '退出登录时出错，但已清除本地状态'
    toast.error('退出登录时发生错误')
  }
}
</script>

<div class="container mx-auto p-8">
  <h1 class="mb-6 text-2xl font-bold">认证机制测试页面</h1>

  <div class="space-y-6">
    <!-- 当前认证状态 -->
    <div class="rounded-lg border p-4">
      <h2 class="mb-3 text-lg font-semibold">当前认证状态</h2>
      <div class="space-y-2 text-sm">
        <p><strong>已认证:</strong> {authState?.isAuthenticated ? '是' : '否'}</p>
        <p><strong>用户:</strong> {authState?.user?.username || '无'}</p>
        <p><strong>角色:</strong> {authState?.user?.role || '无'}</p>
        <p><strong>Access Token:</strong> {authState?.accessToken ? '存在' : '无'}</p>
        <p><strong>Refresh Token:</strong> {authState?.refreshToken ? '存在' : '无'}</p>
        <p>
          <strong>过期时间:</strong>
          {authState?.expiresAt ? new Date(authState.expiresAt).toLocaleString() : '无'}
        </p>
      </div>
    </div>

    <!-- 测试按钮 -->
    <div class="rounded-lg border p-4">
      <h2 class="mb-3 text-lg font-semibold">测试操作</h2>
      <div class="space-y-2 space-x-2">
        <Button onclick={testTokenRefresh}>测试 API 调用 (自动刷新 Token)</Button>
        <Button onclick={testInvalidToken} variant="outline">测试无效 Token 处理</Button>
        <Button onclick={clearAuth} variant="destructive">清除认证状态</Button>
      </div>
    </div>

    <!-- 测试结果 -->
    {#if testResult}
      <div class="rounded-lg border p-4">
        <h2 class="mb-3 text-lg font-semibold">测试结果</h2>
        <p class="rounded bg-gray-100 p-2 font-mono text-sm dark:bg-gray-800">
          {testResult}
        </p>
      </div>
    {/if}

    <!-- 说明 -->
    <div class="rounded-lg border bg-blue-50 p-4 dark:bg-blue-950">
      <h2 class="mb-3 text-lg font-semibold">测试说明</h2>
      <ul class="space-y-1 text-sm">
        <li>• <strong>测试 API 调用:</strong> 发起需要认证的请求，验证 token 自动刷新机制</li>
        <li>• <strong>测试无效 Token:</strong> 模拟 token 被篡改的情况，验证错误处理</li>
        <li>• <strong>清除认证状态:</strong> 手动退出登录</li>
      </ul>
    </div>
  </div>
</div>
