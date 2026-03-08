<template>
  <!-- Create Folder Modal -->
  <n-modal
    :show="showCreate"
    preset="dialog"
    title="Tạo thư mục mới"
    positive-text="Tạo"
    negative-text="Hủy"
    :loading="creating"
    @positive-click="handleCreate"
    @negative-click="$emit('update:showCreate', false)"
    @close="$emit('update:showCreate', false)"
  >
    <n-input
      v-model:value="createName"
      placeholder="Tên thư mục"
      autofocus
      @keydown.enter="handleCreate"
    />
  </n-modal>

  <!-- Rename Modal -->
  <n-modal
    :show="showRename"
    preset="dialog"
    title="Đổi tên"
    positive-text="Lưu"
    negative-text="Hủy"
    :loading="renaming"
    @positive-click="handleRename"
    @negative-click="$emit('update:showRename', false)"
    @close="$emit('update:showRename', false)"
  >
    <n-input
      v-model:value="renameName"
      placeholder="Tên mới"
      autofocus
      @keydown.enter="handleRename"
    />
  </n-modal>

  <!-- Delete Confirmation -->
  <n-modal
    :show="showDelete"
    preset="dialog"
    type="warning"
    title="Xác nhận xóa"
    positive-text="Xóa"
    negative-text="Hủy"
    :loading="deleting"
    @positive-click="handleDelete"
    @negative-click="$emit('update:showDelete', false)"
    @close="$emit('update:showDelete', false)"
  >
    <p>
      Bạn có chắc muốn xóa <strong>{{ targetItem?.name }}</strong>?
      <br />
      <span class="delete-warning">Tất cả nội dung bên trong sẽ bị xóa.</span>
    </p>
  </n-modal>
</template>

<script setup lang="ts">
const props = defineProps<{
  showCreate: boolean
  showRename: boolean
  showDelete: boolean
  targetItem: { id: string; name: string } | null
  parentId?: string
}>()

const emit = defineEmits<{
  'update:showCreate': [value: boolean]
  'update:showRename': [value: boolean]
  'update:showDelete': [value: boolean]
  create: [name: string, parentId?: string]
  rename: [name: string]
  delete: []
}>()

const createName = ref('')
const renameName = ref('')
const creating = ref(false)
const renaming = ref(false)
const deleting = ref(false)

// Sync rename name when target changes
watch(() => props.targetItem, (item) => {
  if (item) renameName.value = item.name
})

// Reset create name when modal opens
watch(() => props.showCreate, (show) => {
  if (show) createName.value = ''
})

function handleCreate() {
  if (!createName.value.trim()) return
  emit('create', createName.value.trim(), props.parentId)
}

function handleRename() {
  if (!renameName.value.trim()) return
  emit('rename', renameName.value.trim())
}

function handleDelete() {
  emit('delete')
}
</script>

<style scoped>
.delete-warning {
  font-size: var(--font-size-sm);
  color: var(--color-error);
}
</style>
