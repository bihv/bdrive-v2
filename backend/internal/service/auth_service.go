package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/biho/onedrive/internal/config"
	"github.com/biho/onedrive/internal/dto"
	"github.com/biho/onedrive/internal/model"
	"github.com/biho/onedrive/internal/repository"
	"github.com/biho/onedrive/pkg/hash"
	jwtpkg "github.com/biho/onedrive/pkg/jwt"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	cfg              *config.Config
	log              *zap.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo *repository.UserRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	cfg *config.Config,
	log *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		cfg:              cfg,
		log:              log,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, string, error) {
	// Check if email exists
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		s.log.Error("Failed to check email existence", zap.Error(err))
		return nil, "", fmt.Errorf("internal error")
	}
	if exists {
		return nil, "", fmt.Errorf("email already registered")
	}

	// Hash password with Argon2id
	params := &hash.Argon2Params{
		Memory:      s.cfg.Argon2.Memory,
		Iterations:  s.cfg.Argon2.Iterations,
		Parallelism: s.cfg.Argon2.Parallelism,
		SaltLength:  s.cfg.Argon2.SaltLength,
		KeyLength:   s.cfg.Argon2.KeyLength,
	}

	hashedPassword, err := hash.HashPassword(req.Password, params)
	if err != nil {
		s.log.Error("Failed to hash password", zap.Error(err))
		return nil, "", fmt.Errorf("internal error")
	}

	user := &model.User{
		Email:    req.Email,
		Password: hashedPassword,
		FullName: req.FullName,
	}

	if err := s.userRepo.Create(user); err != nil {
		s.log.Error("Failed to create user", zap.Error(err))
		return nil, "", fmt.Errorf("failed to create account")
	}

	return s.generateTokens(user, "")
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(req *dto.LoginRequest, ipAddress string) (*dto.AuthResponse, string, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Generic message to prevent email enumeration
			return nil, "", fmt.Errorf("invalid credentials")
		}
		s.log.Error("Failed to find user", zap.Error(err))
		return nil, "", fmt.Errorf("internal error")
	}

	// Check if account is active
	if !user.IsActive {
		return nil, "", fmt.Errorf("account is disabled")
	}

	// Verify password with constant-time comparison
	match, err := hash.VerifyPassword(req.Password, user.Password)
	if err != nil {
		s.log.Error("Failed to verify password", zap.Error(err))
		return nil, "", fmt.Errorf("internal error")
	}
	if !match {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(user.ID)

	return s.generateTokens(user, ipAddress)
}

// RefreshToken generates new tokens using a valid refresh token.
func (s *AuthService) RefreshToken(refreshToken, ipAddress string) (*dto.AuthResponse, string, error) {
	tokenHash := jwtpkg.HashToken(refreshToken)

	// Find the stored refresh token
	storedToken, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", fmt.Errorf("invalid refresh token")
		}
		s.log.Error("Failed to find refresh token", zap.Error(err))
		return nil, "", fmt.Errorf("internal error")
	}

	// Revoke the old token (rotation)
	if err := s.refreshTokenRepo.Revoke(storedToken.ID); err != nil {
		s.log.Error("Failed to revoke old refresh token", zap.Error(err))
	}

	// Get the user
	user, err := s.userRepo.FindByID(storedToken.UserID)
	if err != nil {
		return nil, "", fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, "", fmt.Errorf("account is disabled")
	}

	return s.generateTokens(user, ipAddress)
}

// Logout revokes all refresh tokens for the user.
func (s *AuthService) Logout(userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}
	return s.refreshTokenRepo.RevokeAllForUser(uid)
}

// GetMe retrieves the current user's profile.
func (s *AuthService) GetMe(userID string) (*dto.UserResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("internal error")
	}

	return s.toUserResponse(user), nil
}

// generateTokens creates access and refresh tokens.
func (s *AuthService) generateTokens(user *model.User, ipAddress string) (*dto.AuthResponse, string, error) {
	// Generate access token
	accessToken, err := jwtpkg.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.Role,
		s.cfg.JWT.AccessSecret,
		s.cfg.JWT.AccessExpiry,
	)
	if err != nil {
		s.log.Error("Failed to generate access token", zap.Error(err))
		return nil, "", fmt.Errorf("internal error")
	}

	// Generate refresh token
	rawRefreshToken := jwtpkg.GenerateRefreshToken()
	tokenHash := jwtpkg.HashToken(rawRefreshToken)

	refreshTokenModel := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		IPAddress: ipAddress,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshExpiry),
	}

	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		s.log.Error("Failed to store refresh token", zap.Error(err))
		return nil, "", fmt.Errorf("internal error")
	}

	response := &dto.AuthResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.cfg.JWT.AccessExpiry.Seconds()),
		User:        *s.toUserResponse(user),
	}

	return response, rawRefreshToken, nil
}

// toUserResponse converts a User model to UserResponse DTO.
func (s *AuthService) toUserResponse(user *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:                user.ID.String(),
		Email:             user.Email,
		FullName:          user.FullName,
		AvatarURL:         user.AvatarURL,
		Role:              user.Role,
		IsVerified:        user.IsVerified,
		StorageQuotaBytes: user.StorageQuotaBytes,
		StorageUsedBytes:  user.StorageUsedBytes,
		LastLoginAt:       user.LastLoginAt,
		CreatedAt:         user.CreatedAt,
	}
}
