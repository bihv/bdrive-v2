import type { ApiResponse } from '~/types/auth'
import type {
  AuthenticatePublicLinkRequest,
  AuthenticatePublicLinkResponse,
  PublicFolderListing,
  PublicLinkDetail,
} from '~/types/public-link'

export function usePublicLinkApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase

  async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      ...(options.headers as Record<string, string>),
    }

    if (options.body != null && !(options.body instanceof FormData)) {
      headers['Content-Type'] = headers['Content-Type'] || 'application/json'
    }

    const response = await $fetch<ApiResponse<T>>(`${baseURL}${endpoint}`, {
      ...options,
      method: options.method as any,
      headers,
    })

    return response.data
  }

  function getPublicLink(token: string, sessionToken?: string): Promise<PublicLinkDetail> {
    const headers = sessionToken
      ? { 'X-Public-Link-Session': sessionToken }
      : undefined

    return request<PublicLinkDetail>(`/api/v1/public-links/${token}`, {
      method: 'GET',
      headers,
    })
  }

  function authenticatePublicLink(token: string, body: AuthenticatePublicLinkRequest): Promise<AuthenticatePublicLinkResponse> {
    return request<AuthenticatePublicLinkResponse>(`/api/v1/public-links/${token}/authenticate`, {
      method: 'POST',
      body: JSON.stringify(body),
    })
  }

  function listSharedItems(token: string, sessionToken?: string, parentId?: string): Promise<PublicFolderListing> {
    const query = parentId ? `?parent_id=${encodeURIComponent(parentId)}` : ''
    const headers = sessionToken
      ? { 'X-Public-Link-Session': sessionToken }
      : undefined

    return request<PublicFolderListing>(`/api/v1/public-links/${token}/items${query}`, {
      method: 'GET',
      headers,
    })
  }

  function buildStreamUrl(token: string, itemId?: string, sessionToken?: string, download = false): string {
    const params = new URLSearchParams()
    if (itemId) params.set('item_id', itemId)
    if (download) params.set('download', 'true')
    if (sessionToken) params.set('session', sessionToken)

    const query = params.toString()
    return `${baseURL}/api/v1/public-links/${token}/stream${query ? `?${query}` : ''}`
  }

  return {
    getPublicLink,
    authenticatePublicLink,
    listSharedItems,
    buildStreamUrl,
  }
}
