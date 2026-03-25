<script setup lang="ts">
import { computed } from 'vue'
import type { Item } from '~/types/folder'
import FileManagerGrid from './FileManagerGrid.vue'
import FileManagerList from './FileManagerList.vue'
import { useFileManagerView } from '~/composables/useFileManagerView'

interface Props {
  items: Item[]
  displayLoading: boolean
  isTrashView: boolean
  actions: ReturnType<typeof import('~/composables/useItemActions').useItemActions>
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'rename-request'): void
  (e: 'properties-item', id: string): void
}>()

const { viewMode } = useFileManagerView()

const isTrashView = computed(() => props.isTrashView)

const {
  showContextMenu,
  showTrashMenu,
  contextX,
  contextY,
  contextMenuOptions,
  trashMenuOptions,
  openItem,
  showMenuForItem,
  showMenuAt,
  handleMenuSelect,
  handleTrashSelect,
  requestRename,
  propertiesItemId,
  closeMenus,
  applyTrashAction,
} = props.actions

watch(propertiesItemId, (id) => {
  if (id) {
    emit('properties-item', id)
  }
})

function onAction(event: {
  type: string
  item: Item
  element?: HTMLElement
  eventX?: number
  eventY?: number
}) {
  const { type, item, element, eventX, eventY } = event

  switch (type) {
    case 'open':
      openItem(item)
      break
    case 'menu':
      if (element) {
        showMenuForItem(item, element, isTrashView.value)
      }
      break
    case 'context':
      showMenuAt(eventX!, eventY!, item, false)
      break
    case 'trash-context':
      showMenuAt(eventX!, eventY!, item, true)
      break
    case 'restore':
      applyTrashAction(item, 'restore')
      break
    case 'permanent-delete':
      applyTrashAction(item, 'permanent-delete')
      break
    case 'rename':
      requestRename()
      emit('rename-request')
      break
  }
}
</script>

<template>
  <div class="fm-view">
    <FileManagerGrid
      v-if="viewMode === 'grid'"
      :display-items="items"
      :display-loading="displayLoading"
      :is-trash-view="isTrashView"
      @action="onAction"
    />
    <FileManagerList
      v-else-if="viewMode === 'list'"
      :display-items="items"
      :display-loading="displayLoading"
      :is-trash-view="isTrashView"
      @action="onAction"
    />

    <n-dropdown
      :show="showContextMenu"
      :x="contextX"
      :y="contextY"
      trigger="manual"
      placement="bottom-start"
      :options="contextMenuOptions"
      @select="handleMenuSelect"
      @clickoutside="closeMenus"
    />
    <n-dropdown
      :show="showTrashMenu"
      :x="contextX"
      :y="contextY"
      trigger="manual"
      placement="bottom-start"
      :options="trashMenuOptions"
      @select="handleTrashSelect"
      @clickoutside="closeMenus"
    />
  </div>
</template>

<style scoped>
.fm-view {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
}
</style>
