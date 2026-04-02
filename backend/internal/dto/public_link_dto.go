package dto

type CreatePublicLinkRequest struct {
	Password  *string `json:"password" validate:"omitempty,min=8,max=128"`
	ExpiresAt *string `json:"expires_at" validate:"omitempty"`
}

type UpdatePublicLinkRequest struct {
	Password        *string `json:"password" validate:"omitempty,min=8,max=128"`
	PasswordEnabled *bool   `json:"password_enabled" validate:"omitempty"`
	ExpiresAt       *string `json:"expires_at" validate:"omitempty"`
	ClearExpiry     bool    `json:"clear_expiry"`
}

type PublicLinkResponse struct {
	ID                string  `json:"id"`
	ItemID            string  `json:"item_id"`
	Token             string  `json:"token"`
	Status            string  `json:"status"`
	RequiresPassword  bool    `json:"requires_password"`
	ExpiresAt         *string `json:"expires_at"`
	RevokedAt         *string `json:"revoked_at"`
	AccessCount       int64   `json:"access_count"`
	LastAccessedAt    *string `json:"last_accessed_at"`
	SessionTTLSeconds int64   `json:"session_ttl_seconds"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type AuthenticatePublicLinkRequest struct {
	Password string `json:"password" validate:"required,min=1,max=128"`
}

type AuthenticatePublicLinkResponse struct {
	SessionToken string `json:"session_token"`
	ExpiresAt    string `json:"expires_at"`
}

type PublicSharedItemResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	IsFolder  bool    `json:"is_folder"`
	MimeType  *string `json:"mime_type"`
	Size      int64   `json:"size"`
	UpdatedAt string  `json:"updated_at"`
}

type PublicLinkDetailResponse struct {
	ID                    string                    `json:"id"`
	Token                 string                    `json:"token"`
	Status                string                    `json:"status"`
	RequiresPassword      bool                      `json:"requires_password"`
	AccessGranted         bool                      `json:"access_granted"`
	ExternalAuthAvailable bool                      `json:"external_auth_available"`
	ExpiresAt             *string                   `json:"expires_at"`
	RevokedAt             *string                   `json:"revoked_at"`
	SessionExpiresAt      *string                   `json:"session_expires_at"`
	Item                  *PublicSharedItemResponse `json:"item"`
}

type PublicFolderListingResponse struct {
	RootItem      PublicSharedItemResponse   `json:"root_item"`
	CurrentFolder PublicSharedItemResponse   `json:"current_folder"`
	Breadcrumbs   []PublicSharedItemResponse `json:"breadcrumbs"`
	Items         []PublicSharedItemResponse `json:"items"`
}
