<template>
  <div class="folder-tree">
    <n-spin :show="loading" size="small">
      <div v-if="tree.length === 0 && !loading" class="tree-empty">
        <span class="tree-empty-text">Chưa có thư mục</span>
      </div>

      <n-tree
        v-else
        block-line
        expand-on-click
        :data="tree as any[]"
        key-field="id"
        label-field="name"
        children-field="children"
        :selected-keys="selectedKeys"
        @update:selected-keys="handleSelectedKeysUpdate"
        :node-props="nodeProps"
        :render-prefix="renderPrefix"
      />
    </n-spin>

    <n-dropdown
      placement="bottom-start"
      trigger="manual"
      :x="menuX"
      :y="menuY"
      :options="menuOptions"
      :show="showMenu"
      @clickoutside="showMenu = false"
      @select="handleMenuSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { h, ref, computed } from 'vue'
import type { TreeOption } from 'naive-ui'
import type { FolderTreeNode } from '~/types/folder'
import FileIcon from './FileIcon.vue'

const props = defineProps<{
  tree: FolderTreeNode[]
  selectedId: string | null | undefined
  loading: boolean
}>()

const emit = defineEmits<{
  select: [node: FolderTreeNode]
  create: [parentId: string]
  rename: [node: FolderTreeNode]
  delete: [node: FolderTreeNode]
}>()

const selectedKeys = computed(() => (props.selectedId ? [props.selectedId] : []))

function handleSelectedKeysUpdate(keys: Array<string | number>, options: Array<TreeOption | null>) {
  if (options.length > 0 && options[0]) {
    emit('select', options[0] as unknown as FolderTreeNode)
  }
}

// Context Menu State
const showMenu = ref(false)
const menuX = ref(0)
const menuY = ref(0)
const contextMenuNode = ref<FolderTreeNode | null>(null)

const menuOptions = [
  { label: 'Tạo thư mục con', key: 'create' },
  { label: 'Đổi tên', key: 'rename' },
  { type: 'divider', key: 'd1' },
  { label: 'Xóa', key: 'delete' },
]

function handleContextMenu(e: MouseEvent, option: TreeOption) {
  e.preventDefault()
  showMenu.value = false
  setTimeout(() => {
    menuX.value = e.clientX
    menuY.value = e.clientY
    contextMenuNode.value = option as unknown as FolderTreeNode
    showMenu.value = true
  }, 50)
}

function handleMenuSelect(key: string) {
  showMenu.value = false
  if (!contextMenuNode.value) return
  if (key === 'create') emit('create', contextMenuNode.value.id)
  if (key === 'rename') emit('rename', contextMenuNode.value)
  if (key === 'delete') emit('delete', contextMenuNode.value)
}

const nodeProps = ({ option }: { option: TreeOption }) => {
  return {
    onContextmenu(e: MouseEvent) {
      handleContextMenu(e, option)
    }
  }
}

const renderPrefix = ({ option }: { option: TreeOption }) => {
  const node = option as unknown as FolderTreeNode
  return h(
    'span',
    {
      class: 'tree-node-icon',
      style: node.color ? { color: node.color } : undefined,
    },
    [
      h(FileIcon, {
        filename: node.name,
        isFolder: true,
        size: 18,
      }),
    ],
  )
}


</script>

<style scoped>
.folder-tree {
  padding: 0 0.5rem;
}

.tree-empty {
  padding: 1rem;
  text-align: center;
}

.tree-empty-text {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.tree-node-icon {
  display: inline-flex;
  align-items: center;
  margin-right: 0.25rem;
  flex-shrink: 0;
}

/* Tree node hover & selected */
.folder-tree :deep(.n-tree-node) {
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}

.folder-tree :deep(.n-tree-node:hover) {
  background: var(--color-surface-hover);
}

.folder-tree :deep(.n-tree-node--selected) {
  background: var(--color-primary-light) !important;
  box-shadow: inset 3px 0 0 var(--color-primary);
}

.folder-tree :deep(.n-tree-node--selected .n-tree-node-content__text) {
  color: var(--color-primary);
  font-weight: 500;
}
</style>
