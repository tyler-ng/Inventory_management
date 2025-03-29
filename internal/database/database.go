package database

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/inventory-management-system/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	// Configure GORM logger
	gormLogger := logger.Default
	if cfg.Environment == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Connect to database with retry mechanism
	var db *gorm.DB
	var err error
	maxRetries := 5
	
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
		
		if err == nil {
			break
		}
		
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(time.Second * 5)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established successfully")
	return db, nil
}

// MigrateDB performs database migrations
func MigrateDB(db *gorm.DB) error {
	// Import model packages to auto-migrate
	// Note: for production, use a dedicated migration tool instead of AutoMigrate
	
	log.Println("Running database migrations...")
	// Add your model imports and migrations here
	
	return nil
}