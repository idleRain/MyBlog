<script lang="ts">
import { setLocale, getLocale } from '$lib/paraglide/runtime'
import { m } from '$i18n'
import { Button } from '$ui'

console.log('Current locale:', getLocale())
console.log('Available messages:', Object.keys(m))
</script>

<div class="space-y-6 p-6">
  <!-- 语言切换 -->
  <div class="space-x-2">
    <Button class="rounded bg-blue-500 px-4 py-2 text-white" onclick={() => setLocale('zh')}>
      中文
    </Button>
    <Button class="rounded bg-green-500 px-4 py-2 text-white" onclick={() => setLocale('en')}>
      English
    </Button>
  </div>

  <div>
    {m['ui:button.save']()}
  </div>

  <!-- 方式1：使用t函数 -->
  <section>
    <h2 class="text-xl font-semibold">方式1：t函数（类型安全）</h2>
    <div class="grid grid-cols-2 gap-4">
      <div>
        <h3 class="font-medium">Common</h3>
        <p>{m['common:greeting']()}</p>
        <p>{m['common:welcome']()}</p>
        <p>{m['common:thanks']()}</p>
      </div>

      <div>
        <h3 class="font-medium">Auth</h3>
        <p>{m['auth:login']()}</p>
        <p>{m['auth:register']()}</p>
        <p>{m['auth:username']()}</p>
      </div>
    </div>
  </section>

  <!-- 方式2：使用i18n命名空间对象 -->
  <section>
    <h2 class="text-xl font-semibold">方式2：命名空间对象</h2>
    <div class="grid grid-cols-3 gap-4">
      <div>
        <h3 class="font-medium">UI Buttons</h3>
        <p>{m['ui:button.save']()}</p>
        <p>{m['ui:button.cancel']()}</p>
        <p>{m['ui:button.submit']()}</p>
      </div>

      <div>
        <h3 class="font-medium">Navigation</h3>
        <p>{m['ui:navigation.home']()}</p>
        <p>{m['ui:navigation.about']()}</p>
        <p>{m['ui:navigation.contact']()}</p>
      </div>

      <div>
        <h3 class="font-medium">Status</h3>
        <p>{m['ui:status.loading']()}</p>
        <p>{m['ui:status.success']()}</p>
        <p>{m['ui:status.error']()}</p>
      </div>
    </div>
  </section>

  <!-- 方式3：直接使用paraglide的m对象 -->
  <section>
    <h2 class="text-xl font-semibold">方式3：直接使用m对象</h2>
    <div class="space-y-2">
      <p>Auth Email: {m['auth:email']()}</p>
      <p>Forgot Password: {m['auth:forgotPassword']()}</p>
    </div>
  </section>
</div>
