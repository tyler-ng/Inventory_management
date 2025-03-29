package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/yourusername/inventory-management-system/internal/config"
	"github.com/yourusername/inventory-management-system/internal/database"
	"github.com/yourusername/inventory-management-system/internal/handlers"
	"github.com/yourusername/inventory-management-system/internal/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer sqlDB.Close()

	// Initialize router
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.Logging)
	
	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	
	// Public routes
	public := apiRouter.PathPrefix("").Subrouter()
	handlers.RegisterPublicRoutes(public, db)
	
	// Protected routes
	protected := apiRouter.PathPrefix("").Subrouter()
	protected.Use(middleware.Authenticate(cfg.JWTSecret))
	handlers.RegisterProtectedRoutes(protected, db)
	
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	
	log.Printf("Server starting on port %s...", port)
	log.Fatal(srv.ListenAndServe())
}