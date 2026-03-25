<template>
  <SettingsPanelShell
    kicker="Lưu trữ"
    title="Quota và vòng đời dữ liệu"
    tag-label="Ưu tiên cao"
    tag-type="success"
  >
    <div class="meter-card">
      <div class="meter-copy">
        <span>Quota hiện tại</span>
        <strong>{{ quotaText }}</strong>
      </div>
      <n-progress
        type="line"
        status="success"
        :percentage="storageUsagePercent"
        :show-indicator="false"
        :height="10"
        :border-radius="999"
      />
      <small>Cần có cảnh báo khi gần đầy, dọn trash tự động và chính sách version file.</small>
    </div>

    <SettingsToggleRow
      :value="draft.storageAlerts"
      title="Cảnh báo khi dùng quá 80%"
      description="Gửi cảnh báo trong ứng dụng và email để tránh hết dung lượng đột ngột."
      @update:value="emit('update:storageAlerts', $event)"
    />
    <SettingsToggleRow
      :value="draft.autoCleanupTrash"
      title="Tự động dọn thùng rác"
      description="Dọn file đã xóa lâu ngày để kiểm soát chi phí lưu trữ."
      @update:value="emit('update:autoCleanupTrash', $event)"
    />
    <SettingsSelectRow
      :value="draft.trashRetentionDays"
      :options="retentionOptions"
      title="Giữ file trong trash"
      @update:value="emit('update:trashRetentionDays', $event as number)"
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
  quotaText: string
  storageUsagePercent: number
  retentionOptions: SettingsOption<number>[]
}>()

const emit = defineEmits<{
  'update:storageAlerts': [value: boolean]
  'update:autoCleanupTrash': [value: boolean]
  'update:trashRetentionDays': [value: number]
}>()
</script>

<style scoped>
.meter-card {
  padding: 1rem;
  margin-bottom: 1rem;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.meter-copy {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.75rem;
}

.meter-copy strong {
  font-size: var(--font-size-xl);
}

.meter-card small {
  color: var(--color-text-secondary);
}

@media (max-width: 768px) {
  .meter-copy {
    display: grid;
  }
}
</style>
