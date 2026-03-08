package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/biho/onedrive/internal/handler"
	"github.com/biho/onedrive/internal/middleware"
)

// Setup configures all application routes.
func Setup(app *fiber.App, authHandler *handler.AuthHandler, accessSecret string) {
	// API v1
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "1Drive API",
		})
	})

	// Auth routes - refresh uses general rate limit (not strict auth limiter)
	auth := api.Group("/auth")
	auth.Post("/refresh", authHandler.RefreshToken)

	// Login/Register use strict rate limiter (5 req / 15 min)
	auth.Use(middleware.SetupAuthRateLimiter())
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected auth routes
	authProtected := auth.Group("")
	authProtected.Use(middleware.AuthMiddleware(accessSecret))
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Get("/me", authHandler.GetMe)
}
