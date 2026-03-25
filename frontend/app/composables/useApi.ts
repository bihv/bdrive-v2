import type { ApiResponse, ApiError } from '~/types/auth'
import type { Item, RecentItem, RestoreItemRequest, SearchResult } from '~/types/folder'
import { useAuthStore } from '~/stores/auth'

export function useApi() {
    const config = useRuntimeConfig()
    const authStore = useAuthStore()
    const router = useRouter()
    const baseURL = config.public.apiBase

    let isRefreshing = false
    let refreshPromise: Promise<boolean> | null = null

    async function refreshToken(): Promise<boolean> {
        if (isRefreshing) {
            return refreshPromise || Promise.resolve(false)
        }

        isRefreshing = true
        refreshPromise = (async () => {
            try {
                const response = await $fetch<ApiResponse<{ access_token: string }>>(`${baseURL}/api/v1/auth/refresh`, {
                    method: 'POST',
                    credentials: 'include', // Send cookies (refresh token)
                })

                if (response.success && response.data?.access_token) {
                    authStore.accessToken = response.data.access_token
                    return true
                }
                return false
            } catch (error) {
                console.error('Token refresh failed:', error)
                authStore.clearAuth()
                router.push('/auth/login')
                return false
            } finally {
                isRefreshing = false
                refreshPromise = null
            }
        })()

        return refreshPromise
    }

    async function apiFetch<T>(
        endpoint: string,
        options: RequestInit = {},
        retryCount = 0,
    ): Promise<T> {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
            ...(options.headers as Record<string, string>),
        }

        // Remove Content-Type for FormData (let browser set multipart boundary)
        if (options.body instanceof FormData) {
            delete headers['Content-Type']
        }

        // Attach access token if available
        if (authStore.accessToken) {
            headers['Authorization'] = `Bearer ${authStore.accessToken}`
        }

        try {
            const response = await $fetch<ApiResponse<T>>(`${baseURL}${endpoint}`, {
                ...options,
                method: options.method as any,
                headers,
                credentials: 'include', // Send cookies (refresh token)
            })

            return response.data
        } catch (error: any) {
            // Check if 401 Unauthorized and we haven't retried yet
            if (error?.response?.status === 401 && retryCount === 0) {
                const refreshed = await refreshToken()
                if (refreshed) {
                    // Retry the original request with new token
                    return apiFetch<T>(endpoint, options, retryCount + 1)
                }
            }

            // Re-throw the error for handling by caller
            throw error
        }
    }

    async function post<T>(endpoint: string, body: unknown): Promise<T> {
        return apiFetch<T>(endpoint, {
            method: 'POST',
            body: JSON.stringify(body),
        })
    }

    async function get<T>(endpoint: string): Promise<T> {
        return apiFetch<T>(endpoint, {
            method: 'GET',
        })
    }

    async function put<T>(endpoint: string, body: unknown): Promise<T> {
        return apiFetch<T>(endpoint, {
            method: 'PUT',
            body: body instanceof FormData ? body : JSON.stringify(body),
        })
    }

    async function uploadToURL(url: string, file: File | Blob, contentType: string, onProgress?: (progress: number) => void): Promise<string | null> {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest()
            
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable && onProgress) {
                    onProgress((e.loaded / e.total) * 100)
                }
            })

            xhr.addEventListener('load', () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    resolve(null)
                } else {
                    reject(new Error(`Upload failed with status ${xhr.status}`))
                }
            })

            xhr.addEventListener('error', () => {
                reject(new Error('Upload failed'))
            })

            xhr.open('PUT', url)
            xhr.setRequestHeader('Content-Type', contentType)
            xhr.send(file)
        })
    }

    async function getTrash(): Promise<Item[]> {
        return apiFetch<Item[]>('/api/v1/trash')
    }

    async function getItem(id: string): Promise<Item> {
        return apiFetch<Item>('/api/v1/items/' + id)
    }

    async function restoreItem(id: string, body?: RestoreItemRequest): Promise<Item> {
        return apiFetch<Item>('/api/v1/trash/' + id + '/restore', {
            method: 'POST',
            body: JSON.stringify(body || {}),
        })
    }

    async function permanentDeleteItem(id: string): Promise<void> {
        return apiFetch<void>('/api/v1/trash/' + id, {
            method: 'DELETE',
        })
    }

    async function searchItems(query: string, limit = 20): Promise<SearchResult[]> {
        return apiFetch<SearchResult[]>(
            `/api/v1/items/search?q=${encodeURIComponent(query)}&limit=${limit}`,
            { method: 'GET' }
        )
    }

    async function getRecentItems(limit = 20): Promise<RecentItem[]> {
        return apiFetch<RecentItem[]>(`/api/v1/items/recent?limit=${limit}`, {
            method: 'GET',
        })
    }

    async function getStarredItems(): Promise<Item[]> {
        return apiFetch<Item[]>('/api/v1/items/starred', {
            method: 'GET',
        })
    }

    async function starItem(id: string): Promise<void> {
        return apiFetch<void>(`/api/v1/items/${id}/star`, {
            method: 'POST',
            body: JSON.stringify({}),
        })
    }

    async function unstarItem(id: string): Promise<void> {
        return apiFetch<void>(`/api/v1/items/${id}/star`, {
            method: 'DELETE',
        })
    }

    async function trackItemActivity(id: string, type: 'open' | 'download' | 'update'): Promise<void> {
        return apiFetch<void>(`/api/v1/items/${id}/activity`, {
            method: 'POST',
            body: JSON.stringify({ type }),
        })
    }

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
        getRecentItems,
        getStarredItems,
        starItem,
        unstarItem,
        trackItemActivity,
    }
}
