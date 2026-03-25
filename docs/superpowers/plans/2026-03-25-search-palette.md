# Search Palette Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Notion-inspired command palette for fuzzy search across all files/folders, accessible via `Cmd+K` or header button.

**Architecture:** Custom overlay shell (CSS positioning + backdrop blur) wrapping Naive UI components (`n-input` for search, custom list for results). Global singleton composable manages keyboard shortcuts and search state. Folder store extended with `highlightedId` to animate file highlight on navigate.

**Tech Stack:** Vue 3 Composition API, Nuxt 4, Naive UI, Pinia, TypeScript, CSS animations

---

## File Structure

```
frontend/app/
├── components/
│   └── search/
│       ├── SearchPalette.vue    ← Root: overlay + state orchestration
│       ├── SearchInput.vue     ← Input with icon + loading spinner
│       ├── SearchResults.vue   ← Scrollable results list
│       ├── SearchItem.vue      ← Single result row
│       └── SearchFooter.vue    ← Hint + result count
├── composables/
│   └── useSearchPalette.ts     ← Keyboard + search + state
├── styles/
│   └── search-palette.css      ← CSS variables + animations
└── types/
    └── folder.ts               ← Add SearchResult type

frontend/app/stores/
└── folder.ts                    ← Add highlightedId state

frontend/app/components/folders/
├── FileManagerGrid.vue         ← Add highlight state
└── FileManagerList.vue         ← Add highlight state

frontend/app/layouts/
└── default.vue                 ← Register palette globally
```

---

## Chunk 1: Styles & Types

### Task 1.1: Create Search Palette CSS

**Files:**
- Create: `frontend/app/styles/search-palette.css`

- [ ] **Step 1: Write the CSS file**

```css
/* ============================================
   Search Palette — CSS Variables & Animations
   ============================================ */

/* Light Mode Variables */
:root {
  --sp-bg: #ffffff;
  --sp-border: #e5e5e5;
  --sp-border-focus: #3366ff;
  --sp-text-primary: #1a1a1a;
  --sp-text-secondary: #888888;
  --sp-text-placeholder: #aaaaaa;
  --sp-accent: #3366ff;
  --sp-highlight-bg: #e8efff;
  --sp-hover-bg: #f5f5f5;
  --sp-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  --sp-backdrop-bg: rgba(0, 0, 0, 0.4);
}

/* Dark Mode Variables */
html.dark,
.dark {
  --sp-bg: #1e1e1e;
  --sp-border: #333333;
  --sp-border-focus: #6699ff;
  --sp-text-primary: #e8e8e8;
  --sp-text-secondary: #666666;
  --sp-text-placeholder: #555555;
  --sp-accent: #6699ff;
  --sp-highlight-bg: #1a2844;
  --sp-hover-bg: #2a2a2a;
  --sp-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  --sp-backdrop-bg: rgba(0, 0, 0, 0.6);
}

/* ============================================
   Overlay & Modal
   ============================================ */

.sp-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 15vh;
  background: var(--sp-backdrop-bg);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  animation: sp-backdrop-in 200ms ease-out forwards;
}

.sp-overlay.is-closing {
  animation: sp-backdrop-out 200ms ease-in forwards;
}

.sp-modal {
  width: 100%;
  max-width: 560px;
  max-height: 480px;
  background: var(--sp-bg);
  border: 1px solid var(--sp-border);
  border-radius: 12px;
  box-shadow: var(--sp-shadow);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: sp-modal-in 250ms cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.sp-overlay.is-closing .sp-modal {
  animation: sp-modal-out 200ms ease-in forwards;
}

/* ============================================
   Input Section
   ============================================ */

.sp-input-wrapper {
  display: flex;
  align-items: center;
  padding: 0 16px;
  height: 52px;
  gap: 10px;
  border-bottom: 1px solid var(--sp-border);
  flex-shrink: 0;
}

.sp-input-icon {
  color: var(--sp-text-placeholder);
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.sp-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  font-size: 15px;
  font-weight: 400;
  color: var(--sp-text-primary);
  line-height: 1.5;
}

.sp-input::placeholder {
  color: var(--sp-text-placeholder);
}

.sp-input:focus {
  outline: none;
}

.sp-input:focus-visible {
  outline: none;
}

.sp-clear-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--sp-text-secondary);
  cursor: pointer;
  flex-shrink: 0;
  transition: all 150ms ease-out;
  padding: 0;
}

.sp-clear-btn:hover {
  background: var(--sp-hover-bg);
  color: var(--sp-text-primary);
}

.sp-loading-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid var(--sp-border);
  border-top-color: var(--sp-accent);
  border-radius: 50%;
  animation: sp-spin 600ms linear infinite;
  flex-shrink: 0;
}

/* ============================================
   Results List
   ============================================ */

.sp-results {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 6px 0;
  min-height: 0;
}

.sp-results::-webkit-scrollbar {
  width: 6px;
}

.sp-results::-webkit-scrollbar-track {
  background: transparent;
}

.sp-results::-webkit-scrollbar-thumb {
  background: var(--sp-border);
  border-radius: 3px;
}

.sp-results::-webkit-scrollbar-thumb:hover {
  background: var(--sp-text-placeholder);
}

/* Empty / Initial States */
.sp-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 16px;
  color: var(--sp-text-secondary);
  font-size: 13px;
  gap: 8px;
}

.sp-empty-icon {
  font-size: 32px;
  opacity: 0.3;
}

/* ============================================
   Search Item
   ============================================ */

.sp-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  gap: 10px;
  cursor: pointer;
  border-radius: 0;
  transition: background-color 150ms ease-out;
  border-left: 2px solid transparent;
  min-height: 52px;
  box-sizing: border-box;
}

.sp-item:first-child {
  border-radius: 6px 6px 0 0;
}

.sp-item:last-child {
  border-radius: 0 0 6px 6px;
}

.sp-item:only-child {
  border-radius: 6px;
}

.sp-item:hover {
  background: var(--sp-hover-bg);
}

.sp-item.selected {
  background: var(--sp-highlight-bg);
  border-left-color: var(--sp-accent);
}

.sp-item-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--sp-text-secondary);
}

.sp-item-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.sp-item-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--sp-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}

.sp-item-name .match {
  color: var(--sp-accent);
  font-weight: 600;
}

.sp-item-path {
  font-size: 12px;
  color: var(--sp-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}

/* ============================================
   Footer
   ============================================ */

.sp-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-top: 1px solid var(--sp-border);
  flex-shrink: 0;
  font-size: 12px;
  color: var(--sp-text-secondary);
  min-height: 36px;
}

.sp-footer-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sp-footer-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sp-shortcut {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.sp-shortcut kbd {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 1px 5px;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  font-size: 11px;
  font-weight: 500;
  border: 1px solid var(--sp-border);
  border-radius: 4px;
  background: var(--sp-hover-bg);
  color: var(--sp-text-secondary);
  line-height: 1.4;
}

/* ============================================
   Highlight animation for file manager items
   ============================================ */

.fm-item.highlighted {
  animation: sp-item-pulse 1.5s ease-out forwards;
}

@keyframes sp-item-pulse {
  0% {
    background: var(--sp-highlight-bg);
    border-color: var(--sp-accent);
    box-shadow: 0 0 0 3px rgba(51, 102, 255, 0.2);
  }
  70% {
    background: var(--sp-highlight-bg);
    border-color: var(--sp-accent);
    box-shadow: 0 0 0 3px rgba(51, 102, 255, 0.2);
  }
  100% {
    background: transparent;
    border-color: transparent;
    box-shadow: none;
  }
}

/* ============================================
   Animations
   ============================================ */

@keyframes sp-backdrop-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes sp-backdrop-out {
  from { opacity: 1; }
  to { opacity: 0; }
}

@keyframes sp-modal-in {
  from {
    opacity: 0;
    transform: translateY(8px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes sp-modal-out {
  from {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
  to {
    opacity: 0;
    transform: translateY(8px) scale(0.98);
  }
}

@keyframes sp-spin {
  to { transform: rotate(360deg); }
}

/* ============================================
   Reduced Motion
   ============================================ */

@media (prefers-reduced-motion: reduce) {
  .sp-overlay,
  .sp-overlay.is-closing,
  .sp-modal,
  .sp-overlay.is-closing .sp-modal,
  .fm-item.highlighted {
    animation: none !important;
    transition: none !important;
  }
}

/* ============================================
   Mobile Responsive
   ============================================ */

@media (max-width: 768px) {
  .sp-overlay {
    align-items: flex-end;
    padding-top: 0;
    padding-bottom: 0;
  }

  .sp-modal {
    max-width: 100%;
    margin: 0 16px 16px;
    max-height: 60vh;
    border-radius: 16px;
  }

  .sp-shortcut {
    display: none;
  }
}
```

- [ ] **Step 2: Verify CSS file syntax**

Check file exists and is valid CSS.

---

### Task 1.2: Add SearchResult Type

**Files:**
- Modify: `frontend/app/types/folder.ts`

- [ ] **Step 1: Append SearchResult type to folder.ts**

Add at the end of the file, before the last export or after existing types:

```typescript
// Search result type for SearchPalette
export interface SearchResult {
  id: string
  name: string
  type: 'file' | 'folder'
  mime_type: string | null
  parent_id: string | null
  path: string // Human-readable path e.g. "Drive / Thư mục A"
  size?: number
  updated_at?: string
}
```

---

## Chunk 2: API Extension

### Task 2.1: Add search API method

**Files:**
- Modify: `frontend/app/composables/useApi.ts`

- [ ] **Step 1: Add searchItems method to useApi composable**

Find the `return { ... }` block at the end of `useApi` and add `searchItems`:

```typescript
    async function searchItems(query: string, limit = 20): Promise<SearchResult[]> {
        const results = await apiFetch<SearchResult[]>('/api/v1/items/search', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                ...(authStore.accessToken ? { 'Authorization': `Bearer ${authStore.accessToken}` } : {}),
            },
        })
        return results
    }
```

Also add import for SearchResult if needed at the top:

```typescript
import type { ApiResponse, ApiError } from '~/types/auth'
import type { Item, RestoreItemRequest, SearchResult } from '~/types/folder'
```

And update the return object to include `searchItems`:

```typescript
    return {
        apiFetch,
        post,
        get,
        put,
        uploadToURL,
        refreshToken,
        getTrash,
        getItem,
        restoreItem,
        permanentDeleteItem,
        searchItems,
    }
```

Note: The backend endpoint `GET /api/v1/items/search?q=<query>&limit=<limit>` should return `SearchResult[]` directly (not wrapped in `{success, data}` — the apiFetch already handles the unwrapping at line 73: `return response.data`).

- [ ] **Step 2: Verify the change compiles**

```bash
cd frontend && npx tsc --noEmit --pretty false 2>&1 | head -20
```

Expected: no new errors related to our changes.

---

## Chunk 3: Composable

### Task 3.1: Create useSearchPalette composable

**Files:**
- Create: `frontend/app/composables/useSearchPalette.ts`

- [ ] **Step 1: Write the composable**

```typescript
import type { SearchResult } from '~/types/folder'

// Singleton state — shared across the entire app
const isOpen = ref(false)
const query = ref('')
const results = ref<SearchResult[]>([])
const selectedIndex = ref(0)
const isLoading = ref(false)
const error = ref<string | null>(null)
const abortController = ref<AbortController | null>(null)

export function useSearchPalette() {
  const api = useApi()
  const folderStore = useFolderStore()
  const router = useRouter()

  // ── Open / Close ──────────────────────────────────────────

  function open() {
    isOpen.value = true
    query.value = ''
    results.value = []
    selectedIndex.value = 0
    isLoading.value = false
    error.value = null
  }

  function close() {
    isOpen.value = false
    query.value = ''
    results.value = []
    selectedIndex.value = 0
    isLoading.value = false
  }

  // ── Search ─────────────────────────────────────────────────

  async function search(q: string) {
    query.value = q

    // Cancel any in-flight request
    if (abortController.value) {
      abortController.value.abort()
    }

    if (!q.trim()) {
      results.value = []
      isLoading.value = false
      return
    }

    isLoading.value = true
    error.value = null

    abortController.value = new AbortController()

    try {
      const data = await api.searchItems(q.trim(), 20)
      results.value = data
      selectedIndex.value = 0
    } catch (e: any) {
      if (e.name === 'AbortError' || e?.cause?.name === 'AbortError') return
      error.value = 'Đã xảy ra lỗi khi tìm kiếm'
      results.value = []
    } finally {
      isLoading.value = false
    }
  }

  // Debounced version — 250ms
  const debouncedSearch = useDebounceFn(search, 250)

  // ── Keyboard Navigation ──────────────────────────────────

  function handleKeydown(e: KeyboardEvent) {
    if (!isOpen.value) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        selectedIndex.value = Math.min(
          selectedIndex.value + 1,
          results.value.length - 1
        )
        scrollSelectedIntoView()
        break
      case 'ArrowUp':
        e.preventDefault()
        selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
        scrollSelectedIntoView()
        break
      case 'Enter':
        e.preventDefault()
        if (results.value.length > 0) {
          navigateToResult(results.value[selectedIndex.value])
        }
        break
      case 'Escape':
        e.preventDefault()
        close()
        break
    }
  }

  function scrollSelectedIntoView() {
    nextTick(() => {
      const el = document.querySelector('.sp-item.selected')
      if (el) {
        el.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
      }
    })
  }

  // ── Global Keyboard Shortcut ──────────────────────────────

  function handleGlobalKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      if (isOpen.value) {
        close()
      } else {
        open()
      }
    }
  }

  // ── Navigate ──────────────────────────────────────────────

  async function navigateToResult(item: SearchResult) {
    close()

    // If it's a folder, navigate into it
    if (item.type === 'folder') {
      await router.push({ path: '/', query: item.parent_id ? { folder: item.parent_id } : {} })
      return
    }

    // For files: navigate to parent folder + highlight
    if (item.parent_id) {
      await router.push({ path: '/', query: { folder: item.parent_id } })

      // Wait for view to update, then highlight
      await nextTick()
      folderStore.setHighlightedId(item.id)
    } else {
      // Root-level file
      await router.push({ path: '/', query: {} })
      await nextTick()
      folderStore.setHighlightedId(item.id)
    }
  }

  // ── Lifecycle ────────────────────────────────────────────

  onMounted(() => {
    document.addEventListener('keydown', handleGlobalKeydown)
  })

  onUnmounted(() => {
    document.removeEventListener('keydown', handleGlobalKeydown)
  })

  // ── Computed helpers ──────────────────────────────────────

  const hasResults = computed(() => results.value.length > 0)
  const showEmpty = computed(() => query.value.trim().length > 0 && !isLoading.value && results.value.length === 0)
  const showInitial = computed(() => query.value.length === 0 && !isLoading.value)

  return {
    // State
    isOpen: readonly(isOpen),
    query,
    results: readonly(results),
    selectedIndex,
    isLoading: readonly(isLoading),
    error: readonly(error),

    // Computed
    hasResults,
    showEmpty,
    showInitial,

    // Actions
    open,
    close,
    search,
    debouncedSearch,
    navigateToResult,
    handleKeydown,
  }
}
```

- [ ] **Step 2: Verify the composable compiles**

```bash
cd frontend && npx tsc --noEmit --pretty false 2>&1 | head -20
```

---

## Chunk 4: UI Components

### Task 4.1: Create SearchInput component

**Files:**
- Create: `frontend/app/components/search/SearchInput.vue`

- [ ] **Step 1: Write the component**

```vue
<template>
  <div class="sp-input-wrapper">
    <span class="sp-input-icon">
      <Icon icon="mdi:magnify" width="18" height="18" />
    </span>

    <input
      ref="inputRef"
      v-model="localQuery"
      type="text"
      class="sp-input"
      :placeholder="placeholder"
      autocomplete="off"
      autocorrect="off"
      autocapitalize="off"
      spellcheck="false"
      @input="onInput"
      @keydown="onKeydown"
    />

    <!-- Clear button -->
    <button
      v-if="localQuery.length > 0 && !loading"
      class="sp-clear-btn"
      aria-label="Clear search"
      @click="onClear"
    >
      <Icon icon="mdi:close" width="14" height="14" />
    </button>

    <!-- Loading spinner -->
    <div v-if="loading" class="sp-loading-spinner" aria-label="Searching..." />
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'

const props = defineProps<{
  modelValue: string
  loading: boolean
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'clear'): void
  (e: 'keydown', event: KeyboardEvent): void
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const localQuery = ref(props.modelValue)

watch(() => props.modelValue, (val) => {
  if (val !== localQuery.value) localQuery.value = val
})

function onInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  emit('update:modelValue', val)
}

function onClear() {
  localQuery.value = ''
  emit('update:modelValue', '')
  emit('clear')
  inputRef.value?.focus()
}

function onKeydown(e: KeyboardEvent) {
  emit('keydown', e)
}

// Expose focus for parent component
defineExpose({
  focus: () => inputRef.value?.focus(),
  selectAll: () => inputRef.value?.select(),
})
</script>
```

### Task 4.2: Create SearchItem component

**Files:**
- Create: `frontend/app/components/search/SearchItem.vue`

- [ ] **Step 1: Write the component**

```vue
<template>
  <div
    class="sp-item"
    :class="{ selected }"
    role="option"
    :aria-selected="selected"
    @click="$emit('click')"
    @mouseenter="$emit('mouseenter')"
  >
    <!-- Icon -->
    <div class="sp-item-icon">
      <FileIcon
        v-if="type === 'folder'"
        :filename="name"
        :is-folder="true"
        :size="24"
      />
      <FileIcon
        v-else
        :filename="name"
        :is-folder="false"
        :size="24"
      />
    </div>

    <!-- Name + Path -->
    <div class="sp-item-content">
      <!-- eslint-disable-next-line vue/no-v-html -->
      <div class="sp-item-name" v-html="highlightedName" />
      <div class="sp-item-path">{{ displayPath }}</div>
    </div>

    <!-- Folder arrow hint -->
    <div v-if="type === 'folder'" class="sp-item-icon" style="opacity: 0.4;">
      <Icon icon="mdi:chevron-right" width="16" height="16" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import FileIcon from '~/components/folders/FileIcon.vue'
import type { SearchResult } from '~/types/folder'

const props = defineProps<{
  item: SearchResult
  selected: boolean
  searchQuery: string
}>()

defineEmits<{
  (e: 'click'): void
  (e: 'mouseenter'): void
}>()

// Highlight matching characters in the name
const highlightedName = computed(() => {
  const q = props.searchQuery.trim()
  if (!q) return escapeHtml(props.item.name)

  const idx = props.item.name.toLowerCase().indexOf(q.toLowerCase())
  if (idx === -1) return escapeHtml(props.item.name)

  const before = escapeHtml(props.item.name.slice(0, idx))
  const match = escapeHtml(props.item.name.slice(idx, idx + q.length))
  const after = escapeHtml(props.item.name.slice(idx + q.length))
  return `${before}<span class="match">${match}</span>${after}`
})

// Shortened path display
const displayPath = computed(() => {
  if (!props.item.path) return ''
  // Show last 2 segments max
  const parts = props.item.path.split(' / ').filter(Boolean)
  if (parts.length <= 2) return props.item.path
  return '... / ' + parts.slice(-2).join(' / ')
})

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}
</script>
```

### Task 4.3: Create SearchResults component

**Files:**
- Create: `frontend/app/components/search/SearchResults.vue`

- [ ] **Step 1: Write the component**

```vue
<template>
  <div
    class="sp-results"
    role="listbox"
    :aria-label="`${results.length} search results`"
  >
    <!-- Results -->
    <template v-if="results.length > 0">
      <SearchItem
        v-for="(item, index) in results"
        :key="item.id"
        :item="item"
        :selected="index === selectedIndex"
        :search-query="searchQuery"
        @click="$emit('select', index)"
        @mouseenter="$emit('hover', index)"
      />
    </template>

    <!-- Empty state (query but no results) -->
    <div v-else-if="showEmpty" class="sp-empty">
      <span class="sp-empty-icon">🔍</span>
      <span>Không tìm thấy kết quả nào</span>
    </div>

    <!-- Initial state (no query) -->
    <div v-else-if="showInitial" class="sp-empty">
      <span class="sp-empty-icon">⌘</span>
      <span>Nhập từ khóa để tìm kiếm</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SearchResult } from '~/types/folder'

defineProps<{
  results: SearchResult[]
  selectedIndex: number
  searchQuery: string
  showEmpty: boolean
  showInitial: boolean
}>()

defineEmits<{
  (e: 'select', index: number): void
  (e: 'hover', index: number): void
}>()
</script>
```

### Task 4.4: Create SearchFooter component

**Files:**
- Create: `frontend/app/components/search/SearchFooter.vue`

- [ ] **Step 1: Write the component**

```vue
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
```

### Task 4.5: Create SearchPalette (root) component

**Files:**
- Create: `frontend/app/components/search/SearchPalette.vue`

- [ ] **Step 1: Write the root component**

```vue
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
          @clear="results = []"
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
  debouncedSearch,
  navigateToResult,
  handleKeydown,
} = useSearchPalette()

const modalRef = ref<HTMLElement | null>(null)

function onSelect(index: number) {
  if (results.value[index]) {
    navigateToResult(results.value[index])
  }
}

// Focus input when opened
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
```

---

## Chunk 5: Store & Highlight

### Task 5.1: Add highlightedId to folder store

**Files:**
- Modify: `frontend/app/stores/folder.ts`

- [ ] **Step 1: Add highlightedId state and action**

In the `state()` function, add `highlightedId`:

```typescript
state: () => ({
    items: [] as Item[],
    folderTree: [] as FolderTreeNode[],
    currentFolderId: null as string | null,
    currentPath: '/' as string,
    loading: false,
    treeLoading: false,
    trashItems: [] as Item[],
    isTrashView: false as boolean,
    trashLoading: false as boolean,
    highlightedId: null as string | null,  // ← ADD THIS
}),
```

In the `actions` object, add:

```typescript
setHighlightedId(id: string | null) {
    this.highlightedId = id
    // Auto-clear highlight after 2s
    if (id !== null) {
        setTimeout(() => {
            if (this.highlightedId === id) {
                this.highlightedId = null
            }
        }, 2000)
    }
},
```

- [ ] **Step 2: Verify TypeScript**

```bash
cd frontend && npx tsc --noEmit --pretty false 2>&1 | head -20
```

### Task 5.2: Add highlight state to FileManagerGrid

**Files:**
- Modify: `frontend/app/components/folders/FileManagerGrid.vue`

- [ ] **Step 1: Add highlighted item support**

Add import and computed at the top of `<script setup>`:

```typescript
import { Icon } from '@iconify/vue'
import FileIcon from './FileIcon.vue'
import type { Item } from '~/types/folder'
import { useFolderStore } from '~/stores/folder'  // ADD

defineProps<{ ... }>()
defineEmits<{ ... }>()

const folderStore = useFolderStore()  // ADD
const highlightedId = computed(() => folderStore.highlightedId)  // ADD
```

In the template, update the `fm-item` div:

```vue
<div
  v-for="item in displayItems"
  :key="item.id"
  class="fm-item"
  :class="{
    'is-folder': item.is_folder,
    'is-trash': isTrashView,
    'highlighted': item.id === highlightedId  // ADD THIS
  }"
  ...
>
```

### Task 5.3: Add highlight state to FileManagerList

**Files:**
- Modify: `frontend/app/components/folders/FileManagerList.vue`

- [ ] **Step 1: Add highlighted item support**

Read the file to find the template's item row, then add similar highlight logic.

First, add to `<script setup>`:

```typescript
import { Icon } from '@iconify/vue'
import FileIcon from './FileIcon.vue'
import type { Item } from '~/types/folder'
import { useFolderStore } from '~/stores/folder'  // ADD

defineProps<{ ... }>()
defineEmits<{ ... }>()

const folderStore = useFolderStore()  // ADD
const highlightedId = computed(() => folderStore.highlightedId)  // ADD
```

Find the `tr` element (item row) and add class:

```vue
<tr
  :class="{
    'highlighted': item.id === highlightedId  // ADD THIS
  }"
  ...
>
```

Add to the `<style scoped>` section of `FileManagerList.vue`:

```css
.highlighted td {
  background: var(--sp-highlight-bg) !important;
  box-shadow: inset 3px 0 0 var(--sp-accent);
  animation: sp-list-pulse 1.5s ease-out forwards;
}

@keyframes sp-list-pulse {
  0%, 70% {
    background: var(--sp-highlight-bg);
    box-shadow: inset 3px 0 0 var(--sp-accent);
  }
  100% {
    background: transparent;
    box-shadow: none;
  }
}
```

---

## Chunk 6: Integration

### Task 6.1: Register SearchPalette globally in layout

**Files:**
- Modify: `frontend/app/layouts/default.vue`

- [ ] **Step 1: Add SearchPalette to default layout**

In the `<template>` section, after `<FilePreviewModal />`:

```vue
<!-- Global File Preview Modal -->
<FilePreviewModal />

<!-- Global Search Palette -->
<SearchPalette />
```

Import it at the top of `<script setup>`:

```typescript
import { Icon } from '@iconify/vue'
import { storeToRefs } from 'pinia'
import { useFolderStore } from '~/stores/folder'
import type { FolderTreeNode } from '~/types/folder'
import SearchPalette from '~/components/search/SearchPalette.vue'  // ADD
```

### Task 6.2: Add search button to FileManagerHeader

**Files:**
- Modify: `frontend/app/components/folders/FileManagerHeader.vue`

- [ ] **Step 1: Add search button to header actions**

In the `<template>`, inside `.fm-actions` div, add before the view switcher:

```vue
<div class="fm-actions">
  <!-- Search button — opens SearchPalette -->
  <n-tooltip trigger="hover">
    <template #trigger>
      <n-button
        quaternary
        size="small"
        @click="openSearchPalette"
      >
        <template #icon>
          <n-icon><Icon icon="mdi:magnify" /></n-icon>
        </template>
      </n-button>
    </template>
    Search <kbd style="font-size:10px; margin-left: 4px;">⌘K</kbd>
  </n-tooltip>

  <div class="fm-view-switcher">
    ...
  </div>
  ...
</div>
```

Add to `<script setup>`:

```typescript
import { Icon } from '@iconify/vue'
import type { BreadcrumbItem } from '~/types/folder'
import { useFileManagerView } from '~/composables/useFileManagerView'
import { useSearchPalette } from '~/composables/useSearchPalette'  // ADD

const { viewMode, setViewMode } = useFileManagerView()
const { open: openSearchPalette } = useSearchPalette()  // ADD
```

Also add `<style>` for the search button:

```css
/* Keep it minimal — inherits from fm-actions layout */
```

### Task 6.3: Import SearchPalette CSS globally

**Files:**
- Modify: `frontend/app/app.vue` or `frontend/app/layouts/default.vue`

Add to the `<script setup>` of `default.vue`:

```typescript
// Import search palette styles
import '~/styles/search-palette.css'
```

---

## Chunk 7: Verification

### Task 7.1: TypeScript check

- [ ] **Step 1: Run TypeScript check**

```bash
cd frontend && npx tsc --noEmit --pretty false 2>&1 | head -30
```

Expected: No new TypeScript errors. Fix any type mismatches.

### Task 7.2: Manual testing checklist

- [ ] `Cmd+K` opens palette, Escape closes it
- [ ] Typing in search input triggers debounced search
- [ ] Results appear with correct icons and paths
- [ ] Arrow Up/Down navigates between results
- [ ] Enter navigates to the selected result's folder and highlights the file
- [ ] Clicking backdrop closes the palette
- [ ] Search button in header opens the palette
- [ ] Highlight animation plays on the file item for 1.5s
- [ ] Light/dark mode colors match spec
- [ ] Mobile layout renders correctly (full-width, bottom-aligned)
- [ ] Empty state and initial state show correct messages
- [ ] Clear button resets search

---

## Backend Requirement

**This implementation requires a new backend endpoint:**

```
GET /api/v1/items/search?q=<query>&limit=20
Authorization: Bearer <token>

Response: SearchResult[]
```

The backend should implement fuzzy matching on the `name` field (case-insensitive) and return results sorted by relevance. Search results should NOT include items in trash.

If the backend does not support this endpoint yet, the search will fail with a 404 — handle gracefully (the composable catches errors and shows `error` message).

---

## Commit Strategy

After completing each chunk, commit with a meaningful message:

```bash
# Chunk 1
git add frontend/app/styles/search-palette.css frontend/app/types/folder.ts
git commit -m "feat(search): add search palette CSS variables, animations, and types"

# Chunk 2
git add frontend/app/composables/useApi.ts
git commit -m "feat(search): add searchItems API method to useApi composable"

# Chunk 3
git add frontend/app/composables/useSearchPalette.ts
git commit -m "feat(search): add useSearchPalette composable with keyboard handling and debounced search"

# Chunk 4
git add frontend/app/components/search/
git commit -m "feat(search): add SearchPalette UI components (SearchPalette, SearchInput, SearchResults, SearchItem, SearchFooter)"

# Chunk 5
git add frontend/app/stores/folder.ts frontend/app/components/folders/FileManagerGrid.vue frontend/app/components/folders/FileManagerList.vue
git commit -m "feat(search): add highlightedId state to folder store and highlight animation to grid/list views"

# Chunk 6
git add frontend/app/layouts/default.vue frontend/app/components/folders/FileManagerHeader.vue
git commit -m "feat(search): integrate SearchPalette globally and add search button to header"
```
