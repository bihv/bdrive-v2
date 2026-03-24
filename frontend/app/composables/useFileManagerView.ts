import type { ViewMode } from '~/types/folder'

export type { ViewMode }

const STORAGE_KEY = 'fm-view-mode'
const VALID_MODES: ViewMode[] = ['grid', 'list']

const stored = localStorage.getItem(STORAGE_KEY) as ViewMode | null
const defaultMode: ViewMode = stored && VALID_MODES.includes(stored) ? stored : 'grid'

const state = reactive({ mode: defaultMode })

export function useFileManagerView() {
    function setViewMode(mode: ViewMode) {
        state.mode = mode
        localStorage.setItem(STORAGE_KEY, mode)
    }

    return {
        viewMode: computed(() => state.mode),
        setViewMode,
        viewModes: VALID_MODES,
    }
}
