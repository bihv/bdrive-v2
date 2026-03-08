package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/biho/onedrive/internal/dto"
	jwtpkg "github.com/biho/onedrive/pkg/jwt"
)

// AuthMiddleware validates the JWT access token.
func AuthMiddleware(accessSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Success: false,
				Error:   "Authorization header is required",
				Code:    "MISSING_AUTH_HEADER",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Success: false,
				Error:   "Invalid authorization format",
				Code:    "INVALID_AUTH_FORMAT",
			})
		}

		claims, err := jwtpkg.ValidateAccessToken(parts[1], accessSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Success: false,
				Error:   "Invalid or expired token",
				Code:    "INVALID_TOKEN",
			})
		}

		// Store user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
