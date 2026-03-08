import type { Item, FolderTreeNode, CreateFolderRequest, UpdateItemRequest } from '~/types/folder'

export function useFolder() {
    const store = useFolderStore()
    const api = useApi()

    async function loadItems(parentId?: string | null) {
        store.setLoading(true)
        try {
            const endpoint = parentId
                ? `/api/v1/items?parent_id=${parentId}`
                : '/api/v1/items'
            const items = await api.get<Item[]>(endpoint)
            store.setItems(items)
        } catch (e) {
            console.error('Failed to load items:', e)
            store.setItems([])
        } finally {
            store.setLoading(false)
        }
    }

    async function loadFolderTree() {
        store.setTreeLoading(true)
        try {
            const tree = await api.get<FolderTreeNode[]>('/api/v1/items/tree')
            store.setFolderTree(tree)
        } catch (e) {
            console.error('Failed to load folder tree:', e)
            store.setFolderTree([])
        } finally {
            store.setTreeLoading(false)
        }
    }

    async function createFolder(data: CreateFolderRequest): Promise<Item | null> {
        try {
            const item = await api.post<Item>('/api/v1/items/folder', data)
            store.addItem(item)
            // Refresh tree after creating folder
            await loadFolderTree()
            return item
        } catch (e: any) {
            throw e
        }
    }

    async function updateItem(id: string, data: UpdateItemRequest): Promise<Item | null> {
        try {
            const item = await api.apiFetch<Item>(`/api/v1/items/${id}`, {
                method: 'PUT',
                body: JSON.stringify(data),
            })
            store.updateItem(item)
            if (data.name) {
                await loadFolderTree()
            }
            return item
        } catch (e: any) {
            throw e
        }
    }

    async function deleteItem(id: string): Promise<void> {
        try {
            await api.apiFetch<void>(`/api/v1/items/${id}`, {
                method: 'DELETE',
            })
            store.removeItem(id)
            await loadFolderTree()
        } catch (e: any) {
            throw e
        }
    }

    function navigateToFolder(folderId: string | null, path: string = '/') {
        store.setCurrentFolder(folderId, path)
        loadItems(folderId)
    }

    return {
        // Actions
        loadItems,
        loadFolderTree,
        createFolder,
        updateItem,
        deleteItem,
        navigateToFolder,

        // Reactive state
        items: computed(() => store.currentItems),
        folderTree: computed(() => store.folderTree),
        currentFolderId: computed(() => store.currentFolder),
        currentPath: computed(() => store.currentPath),
        breadcrumbs: computed(() => store.breadcrumbs),
        loading: computed(() => store.loading),
        treeLoading: computed(() => store.treeLoading),
    }
}
