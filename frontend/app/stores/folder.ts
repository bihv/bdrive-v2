import { defineStore } from 'pinia'
import type { Item, FolderTreeNode } from '~/types/folder'

export const useFolderStore = defineStore('folder', {
    state: () => ({
        items: [] as Item[],
        folderTree: [] as FolderTreeNode[],
        currentFolderId: null as string | null,
        currentPath: '/' as string,
        loading: false,
        treeLoading: false,
    }),

    getters: {
        currentItems: (state) => state.items,
        currentFolder: (state): string | null => state.currentFolderId,
        breadcrumbs: (state): string[] => {
            if (state.currentPath === '/') return ['Root']
            return ['Root', ...state.currentPath.split('/').filter(Boolean)]
        },
    },

    actions: {
        setItems(items: Item[]) {
            this.items = items
        },

        setFolderTree(tree: FolderTreeNode[]) {
            this.folderTree = tree
        },

        setCurrentFolder(folderId: string | null, path: string = '/') {
            this.currentFolderId = folderId
            this.currentPath = path
        },

        addItem(item: Item) {
            this.items.push(item)
            // Sort: folders first, then by name
            this.items.sort((a, b) => {
                if (a.is_folder !== b.is_folder) return a.is_folder ? -1 : 1
                return a.name.localeCompare(b.name)
            })
        },

        updateItem(updated: Item) {
            const index = this.items.findIndex(i => i.id === updated.id)
            if (index !== -1) {
                this.items[index] = updated
            }
        },

        removeItem(id: string) {
            this.items = this.items.filter(i => i.id !== id)
        },

        setLoading(v: boolean) {
            this.loading = v
        },

        setTreeLoading(v: boolean) {
            this.treeLoading = v
        },
    },
})
