export type PublicLinkStatus = 'active' | 'expired' | 'revoked'

export interface PublicLink {
  id: string
  item_id: string
  token: string
  status: PublicLinkStatus
  requires_password: boolean
  expires_at: string | null
  revoked_at: string | null
  access_count: number
  last_accessed_at: string | null
  session_ttl_seconds: number
  created_at: string
  updated_at: string
}

export interface CreatePublicLinkRequest {
  password?: string
  expires_at?: string
}

export interface UpdatePublicLinkRequest {
  password?: string
  password_enabled?: boolean
  expires_at?: string
  clear_expiry?: boolean
}

export interface PublicSharedItem {
  id: string
  name: string
  is_folder: boolean
  mime_type: string | null
  size: number
  updated_at: string
}

export interface PublicLinkDetail {
  id: string
  token: string
  status: PublicLinkStatus
  requires_password: boolean
  access_granted: boolean
  external_auth_available: boolean
  expires_at: string | null
  revoked_at: string | null
  session_expires_at: string | null
  item: PublicSharedItem | null
}

export interface PublicFolderListing {
  root_item: PublicSharedItem
  current_folder: PublicSharedItem
  breadcrumbs: PublicSharedItem[]
  items: PublicSharedItem[]
}

export interface AuthenticatePublicLinkRequest {
  password: string
}

export interface AuthenticatePublicLinkResponse {
  session_token: string
  expires_at: string
}
