package router

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/biho/onedrive/internal/handler"
	"github.com/biho/onedrive/internal/middleware"
	"github.com/biho/onedrive/pkg/storage"
)

// Setup configures all application routes.
func Setup(app *fiber.App, authHandler *handler.AuthHandler, itemHandler *handler.ItemHandler, accessSecret string, b2Client *storage.B2Client) {
	// API v1
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		storageStatus := "not_configured"
		if b2Client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := b2Client.HealthCheck(ctx); err != nil {
				storageStatus = "disconnected"
			} else {
				storageStatus = "connected"
			}
		}

		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "1Drive API",
			"storage": storageStatus,
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

	// Item routes (protected)
	items := api.Group("/items")
	items.Use(middleware.AuthMiddleware(accessSecret))
	items.Get("/tree", itemHandler.GetFolderTree) // Must be before /:id
	items.Post("/folder", itemHandler.CreateFolder)
	items.Post("/upload", itemHandler.UploadFile)
	items.Post("/upload-url", itemHandler.GetPreSignedUploadURL)    // Get pre-signed URL for direct upload
	items.Post("/upload-part-url", itemHandler.GetUploadPartURL)    // Get pre-signed URL for part
	items.Post("/complete-upload", itemHandler.CompleteLargeUpload) // Complete multipart upload
	items.Get("/", itemHandler.ListItems)
	items.Get("/:id", itemHandler.GetItem)
	items.Put("/:id", itemHandler.UpdateItem)
	items.Delete("/:id", itemHandler.DeleteItem)
}
