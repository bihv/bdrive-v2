<template>
  <SettingsPanelShell
    kicker="Đề xuất roadmap"
    title="Các nhóm settings nên có cho hệ thống này"
    :full-width="true"
  >
    <template #default>
      <div class="panel-actions">
        <n-button tertiary type="primary" @click="emit('reset')">Reset demo</n-button>
      </div>

      <div class="suggestion-grid">
        <div
          v-for="group in suggestedGroups"
          :key="group.title"
          class="suggestion-card"
        >
          <div class="suggestion-icon">
            <n-icon size="20"><Icon :icon="group.icon" /></n-icon>
          </div>
          <div>
            <h3>{{ group.title }}</h3>
            <p>{{ group.description }}</p>
          </div>
        </div>
      </div>
    </template>
  </SettingsPanelShell>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import SettingsPanelShell from '~/components/settings/SettingsPanelShell.vue'
import type { SuggestedSettingsGroup } from '~/types/settings'

defineProps<{
  suggestedGroups: SuggestedSettingsGroup[]
}>()

const emit = defineEmits<{
  reset: []
}>()
</script>

<style scoped>
.panel-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: -0.25rem;
  margin-bottom: 1rem;
}

.suggestion-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.suggestion-card {
  display: flex;
  gap: 0.9rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.suggestion-card p {
  color: var(--color-text-secondary);
}

.suggestion-icon {
  width: 2.5rem;
  height: 2.5rem;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  border-radius: 14px;
  background: rgba(59, 130, 246, 0.12);
  color: var(--color-primary);
}

@media (max-width: 1024px) {
  .suggestion-grid {
    grid-template-columns: 1fr;
  }
}
</style>
