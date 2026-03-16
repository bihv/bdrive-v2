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

  <!-- Upload Modal -->
  <n-modal
    :show="showUpload"
    preset="dialog"
    title="Tải lên file"
    positive-text="Đóng"
    negative-text=""
    @negative-click="$emit('update:showUpload', false)"
    @close="$emit('update:showUpload', false)"
  >
    <div class="upload-area">
      <n-upload
        ref="uploadRef"
        :max="1"
        :custom-request="handleUploadRequest"
        :show-file-list="false"
        accept="*"
        @change="handleUploadChange"
      >
        <n-upload-dragger>
          <div class="upload-dragger">
            <n-icon :size="48" color="#18a058">
              <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 24 24"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6z" fill="currentColor"></path><path d="M14 2v6h6" fill="none" stroke="currentColor" stroke-width="2"></path></svg>
            </n-icon>
            <n-text style="font-size: 16px">Nhấn hoặc kéo file vào đây để tải lên</n-text>
            <n-text depth="3" style="font-size: 12px">File nhỏ hơn 5MB sẽ tải lên nhanh chóng. File lớn hơn sẽ tải theo từng phần.</n-text>
          </div>
        </n-upload-dragger>
      </n-upload>

      <div v-if="uploadingFile" class="upload-progress">
        <n-text>{{ uploadingFile.name }}</n-text>
        <n-progress
          type="line"
          :percentage="uploadProgress"
          :indicator-placement="'inside'"
        />
        <n-text depth="3">{{ formatSize(uploadingFile.size) }}</n-text>
      </div>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import type { UploadCustomRequestOptions, UploadFileInfo } from 'naive-ui'

const props = defineProps<{
  showCreate: boolean
  showRename: boolean
  showDelete: boolean
  showUpload: boolean
  targetItem: { id: string; name: string } | null
  parentId?: string
}>()

const emit = defineEmits<{
  'update:showCreate': [value: boolean]
  'update:showRename': [value: boolean]
  'update:showDelete': [value: boolean]
  'update:showUpload': [value: boolean]
  create: [name: string, parentId?: string]
  rename: [name: string]
  delete: []
  upload: [file: File, parentId?: string, onProgress?: (progress: number) => void]
}>()

const createName = ref('')
const renameName = ref('')
const creating = ref(false)
const renaming = ref(false)
const deleting = ref(false)

// Upload state
const uploadingFile = ref<File | null>(null)
const uploadProgress = ref(0)

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

function handleUploadChange(options: { file: UploadFileInfo }) {
  const file = options.file.file
  if (file) {
    uploadingFile.value = file
    uploadProgress.value = 0
    emit('upload', file, props.parentId, (progress) => {
      uploadProgress.value = Math.round(progress)
    })
  }
}

function handleUploadRequest(options: UploadCustomRequestOptions) {
  // Prevent default upload behavior - we handle it manually
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<style scoped>
.delete-warning {
  font-size: var(--font-size-sm);
  color: var(--color-error);
}

.upload-area {
  padding: 16px 0;
}

.upload-dragger {
  padding: 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.upload-progress {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
