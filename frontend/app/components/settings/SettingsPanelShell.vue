<template>
  <article class="settings-panel glass-card" :class="{ 'full-width': fullWidth }">
    <div class="panel-head">
      <div>
        <span class="panel-kicker">{{ kicker }}</span>
        <h2>{{ title }}</h2>
      </div>
      <n-tag v-if="tagLabel" size="small" :type="tagType" round>{{ tagLabel }}</n-tag>
    </div>

    <slot />
  </article>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  kicker: string
  title: string
  tagLabel?: string
  tagType?: 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'
  fullWidth?: boolean
}>(), {
  tagLabel: undefined,
  tagType: 'default',
  fullWidth: false,
})
</script>

<style scoped>
.settings-panel {
  padding: 1.25rem;
  border: 1px solid var(--color-border);
}

.full-width {
  grid-column: 1 / -1;
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.panel-kicker {
  display: inline-flex;
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--color-accent);
  margin-bottom: 0.65rem;
}

.panel-head h2 {
  letter-spacing: -0.03em;
}

@media (max-width: 768px) {
  .panel-head {
    display: grid;
    grid-template-columns: 1fr;
  }
}
</style>
