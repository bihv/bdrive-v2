import type { SearchResult } from '~/types/folder'

// Singleton state — module-level so it's shared across all composable calls
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
      // Ignore abort errors
      if (e?.name === 'AbortError') return
      error.value = 'Đã xảy ra lỗi khi tìm kiếm'
      results.value = []
    } finally {
      isLoading.value = false
    }
  }

  // Debounced version — 250ms
  let debounceTimer: ReturnType<typeof setTimeout> | null = null
  function debouncedSearch(q: string) {
    if (debounceTimer) clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => search(q), 250)
  }

  // ── Clear results ─────────────────────────────────────────

  function clearResults() {
    results.value = []
    query.value = ''
    selectedIndex.value = 0
  }

  // ── Keyboard Navigation ──────────────────────────────────

  function handleKeydown(e: KeyboardEvent) {
    if (!isOpen.value) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        if (results.value.length > 0) {
          selectedIndex.value = Math.min(
            selectedIndex.value + 1,
            results.value.length - 1,
          )
          scrollSelectedIntoView()
        }
        break
      case 'ArrowUp':
        e.preventDefault()
        selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
        scrollSelectedIntoView()
        break
      case 'Enter':
        e.preventDefault()
        if (results.value.length > 0) {
          navigateToResult(results.value[selectedIndex.value]!)
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

  async   function navigateToResult(item: SearchResult) {
    if (!item) return
    close()

    // If it's a folder, navigate into it (open the folder itself)
    if (item.type === 'folder') {
      await router.push({
        path: '/',
        query: { folder: item.id },
      })
      return
    }

    // For files: navigate to parent folder + highlight
    await router.push({
      path: '/',
      query: item.parent_id ? { folder: item.parent_id } : {},
    })

    // Wait for view to update, then highlight
    await nextTick()
    folderStore.setHighlightedId(item.id)
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
  const showEmpty = computed(
    () =>
      query.value.trim().length > 0 &&
      !isLoading.value &&
      results.value.length === 0,
  )
  const showInitial = computed(
    () => query.value.length === 0 && !isLoading.value,
  )

  return {
    // Read-only state
    isOpen: readonly(isOpen),
    isLoading: readonly(isLoading),
    error: readonly(error),

    // State (mutable refs)
    results,
    query,
    selectedIndex,

    // Computed
    hasResults,
    showEmpty,
    showInitial,

    // Actions
    open,
    close,
    clearResults,
    search,
    debouncedSearch,
    navigateToResult,
    handleKeydown,
  }
}
