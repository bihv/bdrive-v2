<template>
  <Icon :icon="iconName" :width="size" :height="size" />
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import { getIconForFile, getIconForFolder } from 'vscode-icons-js'

const props = defineProps({
  filename: {
    type: String,
    default: ''
  },
  isFolder: {
    type: Boolean,
    default: false
  },
  size: {
    type: [Number, String],
    default: 24
  }
})

// Override map for cases where vscode-icons-js SVG names
// don't match the actual Iconify icon names
const overrideMap = {
  'file-type-pdf': 'file-type-pdf2',
  'file-type-webp': 'file-type-image',
}

/**
 * Convert vscode-icons-js SVG name to Iconify icon name.
 * Example: "file_type_pdf.svg" → "vscode-icons:file-type-pdf2"
 */
function toIconifyName(svgName) {
  if (!svgName) return 'vscode-icons:default-file'
  const baseName = svgName.replace('.svg', '').replaceAll('_', '-')
  const finalName = overrideMap[baseName] || baseName
  return 'vscode-icons:' + finalName
}

const iconName = computed(() => {
  const name = props.filename || ''

  if (props.isFolder) {
    const folderIcon = getIconForFolder(name)
    return toIconifyName(folderIcon || 'default_folder.svg')
  }

  const fileIcon = getIconForFile(name)
  return toIconifyName(fileIcon || 'default_file.svg')
})
</script>
