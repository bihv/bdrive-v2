<template>
  <!-- Restore Folder Dialog -->
  <n-modal
    :show="showRestoreDialog"
    preset="card"
    title="Select restore folder"
    style="width: 400px"
    @update:show="$emit('update:showRestoreDialog', $event)"
  >
    <div class="restore-folder-list">
      <div
        class="restore-folder-item"
        :class="{ active: selectedRestoreFolder === null }"
        @click="$emit('update:selectedRestoreFolder', null)"
      >
        <n-icon><Icon icon="mdi:home-outline" /></n-icon>
        <span>Root folder</span>
      </div>
      <div
        v-for="folder in folderTree"
        :key="folder.id"
        class="restore-folder-item"
        :class="{ active: selectedRestoreFolder === folder.id }"
        :style="{ paddingLeft: (folder.depth * 20 + 24) + 'px' }"
        @click="$emit('update:selectedRestoreFolder', folder.id)"
      >
        <n-icon><Icon icon="mdi:folder" /></n-icon>
        <span>{{ folder.name }}</span>
      </div>
    </div>
    <template #footer>
      <n-space justify="end">
        <n-button @click="$emit('update:showRestoreDialog', false)">Cancel</n-button>
        <n-button type="primary" @click="$emit('restore-with-folder')">
          Restore
        </n-button>
      </n-space>
    </template>
  </n-modal>

  <!-- Rename Dialog for Name Conflict during Restore -->
  <n-modal
    :show="showRenameDialog"
    preset="card"
    title="Rename"
    style="width: 400px"
    @update:show="$emit('update:showRenameDialog', $event)"
  >
    <n-input
      :value="newItemName"
      @update:value="$emit('update:newItemName', $event)"
      placeholder="New name"
      @keydown.enter="$emit('restore-with-rename')"
    />
    <template #footer>
      <n-space justify="end">
        <n-button @click="$emit('update:showRenameDialog', false)">Cancel</n-button>
        <n-button type="primary" @click="$emit('restore-with-rename')">
          Restore
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { FolderTreeNode } from '~/types/folder'

defineProps<{
  showRestoreDialog: boolean
  showRenameDialog: boolean
  selectedRestoreFolder: string | null
  newItemName: string
  folderTree: FolderTreeNode[]
}>()

defineEmits<{
  (e: 'update:showRestoreDialog', value: boolean): void
  (e: 'update:showRenameDialog', value: boolean): void
  (e: 'update:selectedRestoreFolder', value: string | null): void
  (e: 'update:newItemName', value: string): void
  (e: 'restore-with-folder'): void
  (e: 'restore-with-rename'): void
}>()
</script>

<style scoped>
.restore-folder-list {
  max-height: 300px;
  overflow-y: auto;
}

.restore-folder-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
  border-radius: 6px;
  transition: all var(--transition-base);
}

.restore-folder-item:hover {
  background: var(--color-surface-hover);
}

.restore-folder-item.active {
  background: rgba(64, 158, 255, 0.15);
  color: var(--color-primary);
}
</style>
