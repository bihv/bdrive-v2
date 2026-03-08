import type { ApiResponse, ApiError } from '~/types/auth'

export function useApi() {
    const config = useRuntimeConfig()
    const authStore = useAuthStore()
    const baseURL = config.public.apiBase

    async function apiFetch<T>(
        endpoint: string,
        options: RequestInit = {},
    ): Promise<T> {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
            ...(options.headers as Record<string, string>),
        }

        // Attach access token if available
        if (authStore.accessToken) {
            headers['Authorization'] = `Bearer ${authStore.accessToken}`
        }

        const response = await $fetch<ApiResponse<T>>(`${baseURL}${endpoint}`, {
            ...options,
            method: options.method as any,
            headers,
            credentials: 'include', // Send cookies (refresh token)
        })

        return response.data
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

    return {
        apiFetch,
        post,
        get,
    }
}
