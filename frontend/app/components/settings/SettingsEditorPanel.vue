<template>
  <SettingsPanelShell
    kicker="OnlyOffice"
    title="Trình chỉnh sửa tài liệu"
    tag-label="Tích hợp"
    tag-type="info"
  >
    <SettingsToggleRow
      :value="draft.editorAutosave"
      title="Tự động lưu khi đang chỉnh sửa"
      description="Giảm mất dữ liệu khi người dùng đóng tab hoặc mạng chập chờn."
      @update:value="emit('update:editorAutosave', $event)"
    />
    <SettingsToggleRow
      :value="draft.lockWhileEditing"
      title="Khóa file khi đang có người chỉnh sửa"
      description="Giảm xung đột version với tài liệu văn phòng."
      @update:value="emit('update:lockWhileEditing', $event)"
    />
    <SettingsSelectRow
      :value="draft.defaultOpenMode"
      :options="openModeOptions"
      title="Chế độ mở mặc định"
      @update:value="emit('update:defaultOpenMode', $event as SettingsDraft['defaultOpenMode'])"
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
  openModeOptions: SettingsOption<SettingsDraft['defaultOpenMode']>[]
}>()

const emit = defineEmits<{
  'update:editorAutosave': [value: boolean]
  'update:lockWhileEditing': [value: boolean]
  'update:defaultOpenMode': [value: SettingsDraft['defaultOpenMode']]
}>()
</script>
