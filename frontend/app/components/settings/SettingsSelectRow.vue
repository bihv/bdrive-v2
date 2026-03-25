<template>
  <div class="setting-row compact">
    <div>
      <strong>{{ title }}</strong>
      <p v-if="description">{{ description }}</p>
    </div>
    <n-select
      :value="value"
      :options="options"
      class="setting-select"
      @update:value="emit('update:value', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import type { SettingsOption } from '~/types/settings'

defineProps<{
  title: string
  description?: string
  value: string | number
  options: SettingsOption[]
}>()

const emit = defineEmits<{
  'update:value': [value: string | number]
}>()
</script>

<style scoped>
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 0;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.setting-row.compact {
  align-items: flex-start;
}

.setting-row strong {
  display: block;
  margin-bottom: 0.25rem;
}

.setting-row p {
  color: var(--color-text-secondary);
}

.setting-select {
  width: 180px;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .setting-row {
    display: grid;
    grid-template-columns: 1fr;
  }

  .setting-select {
    width: 100%;
  }
}
</style>
