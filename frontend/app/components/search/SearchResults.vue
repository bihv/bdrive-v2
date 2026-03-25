<template>
  <div
    class="sp-results"
    role="listbox"
    :aria-label="`${results.length} search results`"
  >
    <!-- Results -->
    <template v-if="results.length > 0">
      <SearchItem
        v-for="(item, index) in results"
        :key="item.id"
        :item="item"
        :selected="index === selectedIndex"
        :search-query="searchQuery"
        @click="$emit('select', index)"
        @mouseenter="$emit('hover', index)"
      />
    </template>

    <!-- Empty state (query but no results) -->
    <div v-else-if="showEmpty" class="sp-empty">
      <span class="sp-empty-icon">🔍</span>
      <span>Không tìm thấy kết quả nào</span>
    </div>

    <!-- Initial state (no query) -->
    <div v-else-if="showInitial" class="sp-empty">
      <span class="sp-empty-icon">⌘</span>
      <span>Nhập từ khóa để tìm kiếm</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SearchResult } from '~/types/folder'

defineProps<{
  results: SearchResult[]
  selectedIndex: number
  searchQuery: string
  showEmpty: boolean
  showInitial: boolean
}>()

defineEmits<{
  (e: 'select', index: number): void
  (e: 'hover', index: number): void
}>()
</script>
