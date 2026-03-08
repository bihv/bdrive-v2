package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/biho/onedrive/internal/model"
)

// RefreshTokenRepository handles database operations for refresh tokens.
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create stores a new refresh token.
func (r *RefreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

// FindByTokenHash finds a non-revoked, non-expired refresh token by hash.
func (r *RefreshTokenRepository) FindByTokenHash(hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	result := r.db.Where(
		"token_hash = ? AND revoked_at IS NULL AND expires_at > ?",
		hash, time.Now(),
	).First(&token)
	if result.Error != nil {
		return nil, result.Error
	}
	return &token, nil
}

// Revoke marks a refresh token as revoked.
func (r *RefreshTokenRepository) Revoke(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&model.RefreshToken{}).Where("id = ?", id).
		Update("revoked_at", &now).Error
}

// RevokeAllForUser revokes all refresh tokens for a user.
func (r *RefreshTokenRepository) RevokeAllForUser(userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now).Error
}

// CleanExpired removes expired tokens older than 30 days.
func (r *RefreshTokenRepository) CleanExpired() error {
	cutoff := time.Now().AddDate(0, 0, -30)
	return r.db.Where("expires_at < ?", cutoff).Delete(&model.RefreshToken{}).Error
}
