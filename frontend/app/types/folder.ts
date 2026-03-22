// Item types for virtual file system
export interface Item {
    id: string
    parent_id: string | null
    name: string
    is_folder: boolean
    depth: number
    path: string
    mime_type: string | null
    size: number
    color: string | null
    sort_order: number
    child_count: number
    created_at: string
    updated_at: string
    deleted_at?: string
}

export interface CreateFolderRequest {
    name: string
    parent_id?: string
    color?: string
}

export interface UpdateItemRequest {
    name?: string
    color?: string
    sort_order?: number
}

// Tree node for folder tree component
export interface FolderTreeNode {
    id: string
    name: string
    path: string
    depth: number
    color: string | null
    children: FolderTreeNode[]
}

// Breadcrumb item for navigation
export interface BreadcrumbItem {
    id: string | null
    name: string
    path: string
}

// Upload types
export interface GetUploadURLRequest {
    parent_id?: string
    name: string
    size: number
    content_type: string
    part_size?: number
}

export interface SimpleUploadResponse {
    upload_url: string
    storage_key: string
    item_id: string
    mime_type: string
    size: number
}

export interface InitiateLargeUploadResponse {
    upload_id: string
    storage_key: string
    item_id: string
    part_size: number
    total_parts: number
}

export interface CompleteLargeUploadRequest {
    upload_id: string
    storage_key: string
    parts: CompletedPart[]
}

export interface CompletedPart {
    part_number: number
    etag: string
}

export interface CompleteLargeUploadResponse {
    id: string
    name: string
    storage_key: string
    size: number
    mime_type: string
    download_url: string | null
}

// Trash types
export interface RestoreItemRequest {
    targetParentID?: string
    newName?: string
}
