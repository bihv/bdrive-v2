package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/biho/onedrive/internal/config"
	"github.com/biho/onedrive/internal/dto"
	"github.com/biho/onedrive/internal/service"
	"github.com/biho/onedrive/pkg/validator"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	authService *service.AuthService
	validator   *validator.Validator
	cfg         *config.Config
	log         *zap.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	authService *service.AuthService,
	validator *validator.Validator,
	cfg *config.Config,
	log *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
		cfg:         cfg,
		log:         log,
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Code:    "INVALID_BODY",
		})
	}

	if errs := h.validator.Validate(&req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"errors":  errs,
			"code":    "VALIDATION_ERROR",
		})
	}

	resp, refreshToken, err := h.authService.Register(&req)
	if err != nil {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		if err.Error() == "email already registered" {
			status = fiber.StatusConflict
			code = "EMAIL_EXISTS"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    code,
		})
	}

	h.setRefreshTokenCookie(c, refreshToken)

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// Login handles user login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Code:    "INVALID_BODY",
		})
	}

	if errs := h.validator.Validate(&req); errs != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"errors":  errs,
			"code":    "VALIDATION_ERROR",
		})
	}

	ipAddress := c.IP()
	resp, refreshToken, err := h.authService.Login(&req, ipAddress)
	if err != nil {
		status := fiber.StatusUnauthorized
		code := "INVALID_CREDENTIALS"
		if err.Error() == "account is disabled" {
			status = fiber.StatusForbidden
			code = "ACCOUNT_DISABLED"
		}
		if err.Error() == "internal error" {
			status = fiber.StatusInternalServerError
			code = "INTERNAL_ERROR"
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
			Code:    code,
		})
	}

	h.setRefreshTokenCookie(c, refreshToken)

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// Logout handles user logout.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	if err := h.authService.Logout(userID); err != nil {
		h.log.Error("Failed to logout", zap.Error(err))
	}

	// Clear refresh token cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.App.Env != "development",
		SameSite: "Strict",
		Path:     "/api/v1/auth",
	})

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    fiber.Map{"message": "Logged out successfully"},
	})
}

// RefreshToken handles token refresh.
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Refresh token not found",
			Code:    "NO_REFRESH_TOKEN",
		})
	}

	ipAddress := c.IP()
	resp, newRefreshToken, err := h.authService.RefreshToken(refreshToken, ipAddress)
	if err != nil {
		// Clear invalid cookie
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			HTTPOnly: true,
			Secure:   h.cfg.App.Env != "development",
			SameSite: "Strict",
			Path:     "/api/v1/auth",
		})
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid or expired refresh token",
			Code:    "INVALID_REFRESH_TOKEN",
		})
	}

	h.setRefreshTokenCookie(c, newRefreshToken)

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// GetMe returns the current user's profile.
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	user, err := h.authService.GetMe(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "User not found",
			Code:    "USER_NOT_FOUND",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    user,
	})
}

// setRefreshTokenCookie sets the refresh token as an HTTP-only cookie.
func (h *AuthHandler) setRefreshTokenCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Expires:  time.Now().Add(h.cfg.JWT.RefreshExpiry),
		HTTPOnly: true,
		Secure:   h.cfg.App.Env != "development",
		SameSite: "Strict",
		Path:     "/api/v1/auth",
	})
}
