import { defineStore } from 'pinia'
import type { Item, FolderTreeNode, BreadcrumbItem } from '~/types/folder'

export const useFolderStore = defineStore('folder', {
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
    }),

    getters: {
        currentItems: (state) => state.items,
        currentFolder: (state): string | null => state.currentFolderId,
        breadcrumbs(state): BreadcrumbItem[] {
            if (state.folderTree.length === 0) return []
            const result: BreadcrumbItem[] = [{ id: null, name: 'Root', path: '/' }]
            if (!state.currentFolderId) return result

            const findPath = (nodes: FolderTreeNode[], targetId: string): FolderTreeNode[] | null => {
                for (const node of nodes) {
                    if (node.id === targetId) return [node]
                    const sub = findPath(node.children || [], targetId)
                    if (sub) return [node, ...sub]
                }
                return null
            }

            const trail = findPath(state.folderTree, state.currentFolderId)
            if (trail) {
                for (const node of trail) {
                    result.push({ id: node.id, name: node.name, path: node.path })
                }
            }
            return result
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

        findFolderById(folderId: string): FolderTreeNode | null {
            const search = (nodes: FolderTreeNode[]): FolderTreeNode | null => {
                for (const node of nodes) {
                    if (node.id === folderId) return node
                    const found = search(node.children || [])
                    if (found) return found
                }
                return null
            }
            return search(this.folderTree)
        },

        setTrashItems(items: Item[]) {
            this.trashItems = items
        },

        setTrashView(v: boolean) {
            this.isTrashView = v
        },

        setTrashLoading(v: boolean) {
            this.trashLoading = v
        },

        removeTrashItem(id: string) {
            this.trashItems = this.trashItems.filter(i => i.id !== id)
        },

        updateTrashItem(updated: Item) {
            const index = this.trashItems.findIndex(i => i.id === updated.id)
            if (index !== -1) {
                this.trashItems[index] = updated
            }
        },
    },
})
