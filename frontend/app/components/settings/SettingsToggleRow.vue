<template>
  <div class="setting-row">
    <div>
      <strong>{{ title }}</strong>
      <p v-if="description">{{ description }}</p>
    </div>
    <n-switch :value="value" @update:value="emit('update:value', $event)" />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  value: boolean
}>()

const emit = defineEmits<{
  'update:value': [value: boolean]
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

.setting-row strong {
  display: block;
  margin-bottom: 0.25rem;
}

.setting-row p {
  color: var(--color-text-secondary);
}

@media (max-width: 768px) {
  .setting-row {
    display: grid;
    grid-template-columns: 1fr;
  }
}
</style>
