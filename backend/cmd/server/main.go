package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"

	"github.com/biho/onedrive/internal/config"
	"github.com/biho/onedrive/internal/database"
	"github.com/biho/onedrive/internal/handler"
	"github.com/biho/onedrive/internal/middleware"
	"github.com/biho/onedrive/internal/model"
	"github.com/biho/onedrive/internal/repository"
	"github.com/biho/onedrive/internal/router"
	"github.com/biho/onedrive/internal/service"
	"github.com/biho/onedrive/pkg/storage"
	"github.com/biho/onedrive/pkg/validator"
)

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Connect to PostgreSQL
	db, err := database.ConnectPostgres(&cfg.DB, logger)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}

	// Enable uuid-ossp extension
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	// Auto-migrate models
	if err := db.AutoMigrate(&model.User{}, &model.RefreshToken{}, &model.Item{}); err != nil {
		logger.Fatal("Failed to run auto-migration", zap.Error(err))
	}

	logger.Info("Database migration completed successfully")

	// Connect to Redis
	rdb, err := database.ConnectRedis(&cfg.Redis, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer rdb.Close()

	// Initialize B2 storage client (non-fatal if not configured)
	var b2Client *storage.B2Client
	if cfg.B2.KeyID != "" && cfg.B2.KeyID != "your_b2_application_key_id" {
		b2Client, err = storage.NewB2Client(&cfg.B2, logger)
		if err != nil {
			logger.Warn("Failed to initialize B2 storage client", zap.Error(err))
		} else {
			logger.Info("B2 storage client initialized",
				zap.String("bucket", cfg.B2.BucketName),
				zap.String("region", cfg.B2.Region),
			)
		}
	} else {
		logger.Warn("B2 storage not configured — skipping initialization")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	itemRepo := repository.NewItemRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg, logger)
	itemService := service.NewItemService(itemRepo, logger)

	// Initialize validator
	v := validator.New()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, v, cfg, logger)
	itemHandler := handler.NewItemHandler(itemService, v, logger)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		Prefork:      cfg.Server.Prefork,
		ErrorHandler: customErrorHandler(logger),
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.SetupSecurityHeaders())
	app.Use(middleware.SetupCORS(&cfg.CORS))
	app.Use(middleware.SetupRateLimiter())
	app.Use(middleware.SetupLogger(logger))

	// Setup routes
	router.Setup(app, authHandler, itemHandler, cfg.JWT.AccessSecret, b2Client)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Starting 1Drive API server",
		zap.String("port", cfg.Server.Port),
		zap.String("env", cfg.App.Env),
	)

	if err := app.Listen(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// customErrorHandler handles unhandled errors globally.
func customErrorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		log.Error("Unhandled error",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)

		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"error":   "Internal server error",
			"code":    "INTERNAL_ERROR",
		})
	}
}
