<template>
  <SettingsPanelShell
    kicker="Chia sẻ"
    title="Link công khai và phân quyền"
    tag-label="Nên có"
    tag-type="default"
  >
    <SettingsToggleRow
      :value="draft.shareLinksExpire"
      title="Mặc định link chia sẻ có hạn dùng"
      description="Tránh tạo link sống vĩnh viễn cho tài liệu nội bộ."
      @update:value="emit('update:shareLinksExpire', $event)"
    />
    <SettingsSelectRow
      :value="draft.defaultSharePermission"
      :options="sharePermissionOptions"
      title="Quyền chia sẻ mặc định"
      @update:value="emit('update:defaultSharePermission', $event as SettingsDraft['defaultSharePermission'])"
    />
    <SettingsToggleRow
      :value="draft.auditPublicDownloads"
      title="Ghi log tải xuống từ link public"
      description="Phục vụ audit và điều tra khi có truy cập bất thường."
      @update:value="emit('update:auditPublicDownloads', $event)"
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
  sharePermissionOptions: SettingsOption<SettingsDraft['defaultSharePermission']>[]
}>()

const emit = defineEmits<{
  'update:shareLinksExpire': [value: boolean]
  'update:defaultSharePermission': [value: SettingsDraft['defaultSharePermission']]
  'update:auditPublicDownloads': [value: boolean]
}>()
</script>
