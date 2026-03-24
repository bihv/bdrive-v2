import type { Item } from '~/types/folder'
import { usePreview } from './usePreview'

export interface ItemActionContext {
  displayItems: Ref<Item[]>
  isTrashView: Ref<boolean>
  onNavigate: (id: string, path: string) => void
  onDeleteRequest: (item: { id: string; name: string }) => void
  onRestore?: (item: Item) => void
  onPermanentDelete?: (item: Item) => void
}

export type MenuOption = Record<string, any>

export function useItemActions(context: ItemActionContext) {
  const { displayItems, isTrashView, onNavigate, onDeleteRequest, onRestore, onPermanentDelete } = context
  const { getPreviewType, openPreview, openPreviewAs } = usePreview()

  // State
  const showContextMenu = ref(false)
  const showTrashMenu = ref(false)
  const contextX = ref(0)
  const contextY = ref(0)
  const contextTarget = ref<{ id: string; name: string } | null>(null)
  const propertiesItemId = ref<string | null>(null)
  const showRenameDialog = ref(false)
  const showDeleteDialog = ref(false)
  const showPropertiesModal = ref(false)

  // Computed: context menu options for regular items
  const contextMenuOptions = computed<MenuOption[]>(() => {
    const opts: MenuOption[] = [
      { label: 'Properties', key: 'properties' },
    ]
    if (contextTarget.value) {
      const item = displayItems.value.find(i => i.id === contextTarget.value?.id)
      if (item && !item.is_folder) {
        const type = getPreviewType(item.name)
        opts.push({ type: 'divider', key: 'd-preview' })
        if (type === 'unknown') {
          opts.push({
            label: 'Open with',
            key: 'open-with',
            children: [
              { label: '📄 OnlyOffice', key: 'open-as-office' },
              { label: '🖼️ Image Viewer', key: 'open-as-image' },
              { label: '🎬 Video Player', key: 'open-as-video' },
              { label: '📖 PDF Viewer', key: 'open-as-pdf' },
              { label: '📝 Text Editor', key: 'open-as-text' },
            ],
          })
        } else {
          opts.push({ label: 'Open', key: 'preview' })
        }
      }
    }
    opts.push({ type: 'divider', key: 'd0' })
    opts.push({ label: 'Rename', key: 'rename' })
    opts.push({ type: 'divider', key: 'd1' })
    opts.push({ label: 'Delete', key: 'delete' })
    return opts
  })

  // Computed: context menu options for trash items
  const trashMenuOptions = computed<MenuOption[]>(() => [
    { label: 'Restore', key: 'restore' },
    { type: 'divider', key: 'd1' },
    { label: 'Delete permanently', key: 'permanent-delete' },
  ])

  // Open menu at exact screen coordinates
  function showMenuAt(x: number, y: number, item: Item, isTrash: boolean) {
    contextX.value = x
    contextY.value = y
    contextTarget.value = { id: item.id, name: item.name }
    if (isTrash) {
      showTrashMenu.value = true
    } else {
      showContextMenu.value = true
    }
  }

  // Open menu below trigger element
  function showMenuForItem(item: Item, triggerEl: HTMLElement, isTrash: boolean) {
    const rect = triggerEl.getBoundingClientRect()
    showMenuAt(rect.left, rect.bottom, item, isTrash)
  }

  // Handle context menu selection
  function handleMenuSelect(key: string) {
    showContextMenu.value = false
    if (key === 'properties') {
      propertiesItemId.value = contextTarget.value?.id || null
      showPropertiesModal.value = true
      return
    }
    if (!contextTarget.value) return
    const item = displayItems.value.find(i => i.id === contextTarget.value?.id)
    if (!item) return
    if (key === 'preview') {
      openPreview({ id: item.id, name: item.name }, displayItems.value)
      return
    }
    if (key.startsWith('open-as-')) {
      const typeMap: Record<string, string> = {
        'open-as-office': 'office',
        'open-as-image': 'image',
        'open-as-video': 'video',
        'open-as-pdf': 'pdf',
        'open-as-text': 'text',
      }
      const forceType = typeMap[key] as any
      if (forceType) {
        openPreviewAs({ id: item.id, name: item.name }, forceType, displayItems.value)
      }
      return
    }
    if (key === 'rename') {
      requestRename()
      return
    }
    if (key === 'delete') {
      onDeleteRequest(contextTarget.value)
      return
    }
  }

  // Handle trash menu selection
  function handleTrashSelect(key: string) {
    showTrashMenu.value = false
    if (!contextTarget.value) return
    const item = displayItems.value.find(i => i.id === contextTarget.value?.id)
    if (!item) return
    if (key === 'restore') {
      onRestore?.(item)
      return
    }
    if (key === 'permanent-delete') {
      onPermanentDelete?.(item)
      return
    }
  }

  // Open item (folder → navigate, file → preview)
  function openItem(item: Item) {
    if (item.is_folder) {
      onNavigate(item.id, item.path)
    } else {
      openPreview({ id: item.id, name: item.name }, displayItems.value)
    }
  }

  // Convenience methods
  function requestRename() {
    showRenameDialog.value = true
  }

  function requestDelete() {
    if (contextTarget.value) {
      onDeleteRequest(contextTarget.value)
    }
  }

  // Close menus (for external use, e.g., clickoutside)
  function closeMenus() {
    showContextMenu.value = false
    showTrashMenu.value = false
  }

  return {
    // State
    showContextMenu: readonly(showContextMenu),
    showTrashMenu: readonly(showTrashMenu),
    contextX: readonly(contextX),
    contextY: readonly(contextY),
    contextTarget: readonly(contextTarget),
    // Modal state — writable for v-model binding in page component
    propertiesItemId,
    showRenameDialog,
    showDeleteDialog,
    showPropertiesModal,

    // Computed options
    contextMenuOptions,
    trashMenuOptions,

    // Methods
    showMenuAt,
    showMenuForItem,
    handleMenuSelect,
    handleTrashSelect,
    openItem,
    requestRename,
    requestDelete,
    closeMenus,
  }
}
