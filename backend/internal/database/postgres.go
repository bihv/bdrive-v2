package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/biho/onedrive/internal/config"
)

// ConnectPostgres establishes a connection to PostgreSQL using GORM.
func ConnectPostgres(cfg *config.DBConfig, log *zap.Logger) (*gorm.DB, error) {
	gormLogLevel := logger.Silent
	if cfg.Host == "localhost" {
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	log.Info("Connected to PostgreSQL",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Name),
	)

	return db, nil
}
