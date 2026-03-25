<template>
  <div class="sp-input-wrapper">
    <span class="sp-input-icon">
      <Icon icon="mdi:magnify" width="18" height="18" />
    </span>

    <input
      ref="inputRef"
      :value="modelValue"
      type="text"
      class="sp-input"
      :placeholder="placeholder"
      autocomplete="off"
      autocorrect="off"
      autocapitalize="off"
      spellcheck="false"
      @input="onInput"
      @keydown="$emit('keydown', $event)"
    />

    <!-- Clear button -->
    <button
      v-if="modelValue.length > 0 && !loading"
      class="sp-clear-btn"
      aria-label="Clear search"
      @click="onClear"
    >
      <Icon icon="mdi:close" width="14" height="14" />
    </button>

    <!-- Loading spinner -->
    <div v-if="loading" class="sp-loading-spinner" aria-label="Searching..." />
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'

const props = defineProps<{
  modelValue: string
  loading: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'clear'): void
  (e: 'keydown', event: KeyboardEvent): void
}>()

const inputRef = ref<HTMLInputElement | null>(null)

function onInput(e: Event) {
  emit('update:modelValue', (e.target as HTMLInputElement).value)
}

function onClear() {
  emit('update:modelValue', '')
  emit('clear')
  inputRef.value?.focus()
}

// Expose focus for parent
defineExpose({
  focus: () => inputRef.value?.focus(),
  selectAll: () => inputRef.value?.select(),
})
</script>
