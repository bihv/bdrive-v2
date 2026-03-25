<template>
  <div
    class="sp-item"
    :class="{ selected }"
    role="option"
    :aria-selected="selected"
    @click="$emit('click')"
    @mouseenter="$emit('mouseenter')"
  >
    <!-- Icon -->
    <div class="sp-item-icon">
      <FileIcon
        v-if="item.type === 'folder'"
        :filename="item.name"
        :is-folder="true"
        :size="24"
      />
      <FileIcon
        v-else
        :filename="item.name"
        :is-folder="false"
        :size="24"
      />
    </div>

    <!-- Name + Path -->
    <div class="sp-item-content">
      <!-- eslint-disable-next-line vue/no-v-html -->
      <div class="sp-item-name" v-html="highlightedName" />
      <div class="sp-item-path">{{ displayPath }}</div>
    </div>

    <!-- Folder arrow hint -->
    <div v-if="item.type === 'folder'" class="sp-item-icon" style="opacity: 0.4;">
      <Icon icon="mdi:chevron-right" width="16" height="16" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import FileIcon from '~/components/folders/FileIcon.vue'
import type { SearchResult } from '~/types/folder'

const props = defineProps<{
  item: SearchResult
  selected: boolean
  searchQuery: string
}>()

defineEmits<{
  (e: 'click'): void
  (e: 'mouseenter'): void
}>()

// Highlight matching characters in the name
const highlightedName = computed(() => {
  const q = props.searchQuery.trim()
  if (!q) return escapeHtml(props.item.name)

  const idx = props.item.name.toLowerCase().indexOf(q.toLowerCase())
  if (idx === -1) return escapeHtml(props.item.name)

  const before = escapeHtml(props.item.name.slice(0, idx))
  const match = escapeHtml(props.item.name.slice(idx, idx + q.length))
  const after = escapeHtml(props.item.name.slice(idx + q.length))
  return `${before}<span class="match">${match}</span>${after}`
})

// Shortened path display — show last 2 segments max
const displayPath = computed(() => {
  if (!props.item.path) return ''
  const parts = props.item.path.split(' / ').filter(Boolean)
  if (parts.length <= 2) return props.item.path
  return '... / ' + parts.slice(-2).join(' / ')
})

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}
</script>
