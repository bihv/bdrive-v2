<template>
  <SettingsPanelShell
    kicker="Bảo mật"
    title="Truy cập và đăng nhập"
    tag-label="Quan trọng"
    tag-type="warning"
  >
    <SettingsToggleRow
      :value="draft.enforce2fa"
      title="Yêu cầu xác thực 2 lớp"
      description="Nên bật cho admin hoặc user có quyền chia sẻ/chỉnh sửa dữ liệu nhạy cảm."
      @update:value="emit('update:enforce2fa', $event)"
    />
    <SettingsToggleRow
      :value="draft.forceSuspiciousLogout"
      title="Đăng xuất khi phát hiện token bất thường"
      description="Giảm rủi ro với refresh token bị lộ hoặc đăng nhập từ thiết bị lạ."
      @update:value="emit('update:forceSuspiciousLogout', $event)"
    />
    <SettingsSelectRow
      :value="draft.sessionTimeout"
      :options="sessionOptions"
      title="Thời gian session"
      @update:value="emit('update:sessionTimeout', $event as number)"
    />
  </SettingsPanelShell>
</template>

<script setup lang="ts">
import SettingsPanelShell from '~/components/settings/SettingsPanelShell.vue'
import SettingsSelectRow from '~/components/settings/SettingsSelectRow.vue'
import SettingsToggleRow from '~/components/settings/SettingsToggleRow.vue'
import type { SettingsDraft, SettingsOption } from '~/types/settings'

defineProps<{
  draft: SettingsDraft
  sessionOptions: SettingsOption<number>[]
}>()

const emit = defineEmits<{
  'update:enforce2fa': [value: boolean]
  'update:forceSuspiciousLogout': [value: boolean]
  'update:sessionTimeout': [value: number]
}>()
</script>
