<script setup lang="ts">
import { computed, toRef } from 'vue'
import type { Item } from '~/types/folder'
import { useFileManagerView } from '~/composables/useFileManagerView'

interface Props {
  items: Item[]
  isTrashView: boolean
  actions: ReturnType<typeof import('~/composables/useItemActions').useItemActions>
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'rename-request'): void
  (e: 'properties-item', id: string): void
}>()

const { currentComponent } = useFileManagerView()

const displayItems = toRef(props, 'items')
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
      handleTrashSelect('restore')
      break
    case 'permanent-delete':
      handleTrashSelect('permanent-delete')
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
    <component
      :is="currentComponent"
      :items="items"
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
