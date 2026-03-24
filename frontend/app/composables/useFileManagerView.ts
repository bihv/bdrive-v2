export type ViewMode = 'grid' | 'list' | 'column'

const viewComponents: Record<ViewMode, string> = {
    grid: 'FileManagerGrid',
    list: 'FileManagerList',
    column: 'FileManagerColumnView',
}

export const viewModes: ViewMode[] = ['grid', 'list', 'column']

export function useFileManagerView() {
    const viewMode = useState<ViewMode>('fileManager.viewMode', () => 'grid')

    const currentComponent = computed(() => viewComponents[viewMode.value])

    function setViewMode(mode: ViewMode) {
        viewMode.value = mode
    }

    return {
        viewMode: viewMode as Readonly<typeof viewMode>,
        currentComponent,
        setViewMode,
        viewModes,
    }
}
