package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/biho/onedrive/internal/model"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail finds a user by email.
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindByID finds a user by ID.
func (r *UserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Create creates a new user.
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// UpdateLastLogin updates the user's last login timestamp.
func (r *UserRepository) UpdateLastLogin(id uuid.UUID) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		Update("last_login_at", gorm.Expr("NOW()")).Error
}

// EmailExists checks if an email is already registered.
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
