<script lang="ts">
import type { UserStatus } from '@myblog/api/modules/user/types'
import { USER_STATUS_CONFIG } from '$lib/constants/user'
import { Badge } from '$ui'

interface Props {
  status: UserStatus
}

let { status }: Props = $props()
</script>

<Badge variant={USER_STATUS_CONFIG[status].variant}>
  {USER_STATUS_CONFIG[status].label}
</Badge>
