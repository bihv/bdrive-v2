import type { Item, FolderTreeNode, CreateFolderRequest, UpdateItemRequest, GetUploadURLRequest, SimpleUploadResponse, InitiateLargeUploadResponse, CompleteLargeUploadRequest, CompleteLargeUploadResponse } from '~/types/folder'

const LARGE_FILE_THRESHOLD = 5 * 1024 * 1024 // 5MB

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

    // Upload file directly to B2 using pre-signed URLs
    async function uploadFile(file: File, parentId?: string | null, onProgress?: (progress: number) => void): Promise<Item | null> {
        try {
            const contentType = file.type || 'application/octet-stream'
            const size = file.size

            // Step 1: Get pre-signed URL from backend
            const request: GetUploadURLRequest = {
                parent_id: parentId || undefined,
                name: file.name,
                size: size,
                content_type: contentType,
            }

            let response: SimpleUploadResponse | InitiateLargeUploadResponse

            if (size <= LARGE_FILE_THRESHOLD) {
                // Simple upload
                response = await api.post<SimpleUploadResponse>('/api/v1/items/upload-url', request)
            } else {
                // Large file - initiate multipart upload
                response = await api.post<InitiateLargeUploadResponse>('/api/v1/items/upload-url', {
                    ...request,
                    part_size: 5 * 1024 * 1024, // 5MB parts
                })
            }

            // Step 2: Upload to B2
            if ('upload_url' in response) {
                // Simple upload
                await api.uploadToURL(response.upload_url, file, contentType, onProgress)
            } else {
                // Large file - multipart upload
                await uploadLargeFile(file, response as InitiateLargeUploadResponse, contentType, onProgress)
            }

            // Refresh items list
            await loadItems(parentId)

            // Return a placeholder item (actual item is created in backend)
            return {
                id: response.item_id,
                parent_id: parentId || null,
                name: file.name,
                is_folder: false,
                depth: 0,
                path: '',
                mime_type: contentType,
                size: file.size,
                color: null,
                sort_order: 0,
                child_count: 0,
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
            }
        } catch (e: any) {
            console.error('Upload failed:', e)
            throw e
        }
    }

    // Upload large file using multipart upload
    async function uploadLargeFile(
        file: File, 
        response: InitiateLargeUploadResponse, 
        contentType: string,
        onProgress?: (progress: number) => void
    ): Promise<void> {
        const { upload_id, storage_key, part_size, total_parts } = response
        const parts: { part_number: number; etag: string }[] = []

        // Upload each part
        for (let i = 1; i <= total_parts; i++) {
            try {
                // Get pre-signed URL for this part (this might fail if token expired)
                const partUrlResponse = await api.post<{ upload_url: string }>('/api/v1/items/upload-part-url', {
                    storage_key,
                    upload_id,
                    part_number: i,
                })

                // Calculate part data
                const start = (i - 1) * part_size
                const end = Math.min(start + part_size, file.size)
                const blob = file.slice(start, end)

                // Upload the part (no ETag needed for B2)
                await api.uploadToURL(partUrlResponse.upload_url, blob, contentType)
                
                // B2 doesn't require ETag for completion, use placeholder
                parts.push({
                    part_number: i,
                    etag: `part-${i}`,
                })

                if (onProgress) {
                    onProgress((i / total_parts) * 100)
                }
            } catch (error: any) {
                // If we get 401, try to refresh token and retry
                if (error?.response?.status === 401) {
                    console.log('Token expired during upload, refreshing...')
                    const refreshed = await api.refreshToken()
                    if (refreshed) {
                        // Retry this part with new token
                        i-- // Decrement to retry this part
                        continue
                    }
                }
                throw error
            }
        }

        // Complete the multipart upload
        await api.post<CompleteLargeUploadResponse>('/api/v1/items/complete-upload', {
            upload_id,
            storage_key,
            parts,
        })
    }

    return {
        // Actions
        loadItems,
        loadFolderTree,
        createFolder,
        updateItem,
        deleteItem,
        navigateToFolder,
        uploadFile,

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
