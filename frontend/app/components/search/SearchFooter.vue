<template>
  <div v-if="resultCount > 0 || showShortcuts" class="sp-footer">
    <div class="sp-footer-left">
      <span v-if="resultCount > 0">
        {{ resultCount }} {{ resultCount === 1 ? 'kết quả' : 'kết quả' }}
      </span>
    </div>
    <div v-if="showShortcuts" class="sp-footer-right">
      <span class="sp-shortcut">
        <kbd>↵</kbd> để mở
      </span>
      <span class="sp-shortcut">
        <kbd>esc</kbd> để đóng
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  resultCount: number
  showShortcuts?: boolean
}>()
</script>
