package dto

import "time"

// AuthResponse is the response for login/register.
type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"`
	User        UserResponse `json:"user"`
}

// UserResponse is the user data exposed to the client.
type UserResponse struct {
	ID                string     `json:"id"`
	Email             string     `json:"email"`
	FullName          string     `json:"full_name"`
	AvatarURL         *string    `json:"avatar_url"`
	Role              string     `json:"role"`
	IsVerified        bool       `json:"is_verified"`
	StorageQuotaBytes int64      `json:"storage_quota_bytes"`
	StorageUsedBytes  int64      `json:"storage_used_bytes"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse is the standard success response.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}
