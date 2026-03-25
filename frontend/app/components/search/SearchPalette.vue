<template>
  <Teleport to="body">
    <div
      v-if="isOpen"
      class="sp-overlay"
      role="dialog"
      aria-modal="true"
      aria-label="Tìm kiếm file và thư mục"
      @click.self="close"
      @keydown="handleKeydown"
    >
      <div class="sp-modal" ref="modalRef">
        <!-- Search Input -->
        <SearchInput
          v-model="query"
          :loading="isLoading"
          placeholder="Tìm kiếm file hoặc thư mục..."
          @keydown="handleKeydown"
          @update:model-value="debouncedSearch"
          @clear="clearResults"
        />

        <!-- Results -->
        <SearchResults
          :results="results"
          :selected-index="selectedIndex"
          :search-query="query"
          :show-empty="showEmpty"
          :show-initial="showInitial"
          @select="onSelect"
          @hover="selectedIndex = $event"
        />

        <!-- Footer -->
        <SearchFooter
          :result-count="results.length"
          :show-shortcuts="query.length > 0"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import SearchInput from './SearchInput.vue'
import SearchResults from './SearchResults.vue'
import SearchFooter from './SearchFooter.vue'
import { useSearchPalette } from '~/composables/useSearchPalette'

const {
  isOpen,
  query,
  results,
  selectedIndex,
  isLoading,
  showEmpty,
  showInitial,
  close,
  clearResults,
  debouncedSearch,
  navigateToResult,
  handleKeydown,
} = useSearchPalette()

const modalRef = ref<HTMLElement | null>(null)

function onSelect(index: number) {
  const item = results.value[index]
  if (item) {
    navigateToResult(item)
  }
}

// Focus + select-all input when palette opens
watch(isOpen, (open) => {
  if (open) {
    nextTick(() => {
      const input = modalRef.value?.querySelector('.sp-input') as HTMLInputElement
      input?.focus()
      input?.select()
    })
  }
})
</script>
