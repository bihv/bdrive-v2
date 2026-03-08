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
