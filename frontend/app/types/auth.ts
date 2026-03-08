// Auth types
export interface User {
    id: string
    email: string
    full_name: string
    avatar_url: string | null
    role: string
    is_verified: boolean
    storage_quota_bytes: number
    storage_used_bytes: number
    last_login_at: string | null
    created_at: string
}

export interface LoginRequest {
    email: string
    password: string
}

export interface RegisterRequest {
    email: string
    password: string
    full_name: string
}

export interface AuthResponse {
    access_token: string
    token_type: string
    expires_in: number
    user: User
}

export interface ApiResponse<T> {
    success: boolean
    data: T
}

export interface ApiError {
    success: boolean
    error: string
    code: string
}
