package database

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/inventory-management-system/internal/config"
	"github.com/yourusername/inventory-management-system/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// Construct the database connection string
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, 
		cfg.DBPort, 
		cfg.DBUser, 
		cfg.DBPassword, 
		cfg.DBName,
	)

	// Configure GORM logger
	gormLogger := logger.Default
	if cfg.Environment == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Attempt to connect to the database with retry mechanism
	var db *gorm.DB
	var err error
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		// Open database connection
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
		
		if err == nil {
			// Connection successful
			log.Println("Database connection established successfully")
			break
		}
		
		log.Printf("Database connection attempt %d failed: %v", i+1, err)
		time.Sleep(time.Second * 5)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection pool: %w", err)
	}
	
	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// MigrateDB performs database migrations
func MigrateDB(db *gorm.DB) error {
	log.Println("Running database migrations...")
	
	// Migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.ProductAttachment{},
		&models.ProductVariant{},
		&models.ProductBundle{},
		&models.ProductSupplier{},
		&models.ProductWarehouse{},
		&models.ProductCategory{},
		&models.Supplier{},
		&models.Warehouse{},
		&models.WarehouseLocation{},
		&models.InventoryTransaction{},
		&models.PurchaseOrder{},
		&models.PurchaseOrderItem{},
		&models.Customer{},
		&models.SalesOrder{},
		&models.SalesOrderItem{},
		&models.AuditLog{},
	)
	
	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}
	
	// Optional: Insert default admin user if not exists
	if err := seedAdminUser(db); err != nil {
		log.Printf("Seeding admin user failed: %v", err)
		return err
	}
	
	log.Println("Database migration completed successfully")
	return nil
}

// seedAdminUser creates a default admin user if no admin exists
func seedAdminUser(db *gorm.DB) error {
	var userCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&userCount)
	
	if userCount == 0 {
		defaultAdmin := models.User{
			Username:     "admin",
			Email:        "admin@example.com",
			FullName:     "System Administrator",
			Role:         "admin",
			Status:       "active",
			PasswordHash: "admin123", // This will trigger password hashing in the BeforeCreate hook
		}
		
		return db.Create(&defaultAdmin).Error
	}
	
	return nil
}