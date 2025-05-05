package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/sales-tracker/auth-service/internal/config"
	"github.com/sales-tracker/auth-service/internal/handler"
	"github.com/sales-tracker/auth-service/internal/middleware"
	"github.com/sales-tracker/auth-service/internal/repository"
	"github.com/sales-tracker/auth-service/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found")
	}

	// Initialize config
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Initialize Echo
	e := echo.New()

	// Validate database configuration
	if cfg.DatabaseURL == "" {
		log.Fatal("DatabaseURL configuration missing - check config.yaml")
	}
	println(cfg.DatabaseURL)

	// Initialize database
	dsn := cfg.DatabaseURL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get the underlying SQL connection
	dbSQL, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}

	// Set connection pool settings
	dbSQL.SetMaxIdleConns(10)
	dbSQL.SetMaxOpenConns(100)
	dbSQL.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := dbSQL.Ping(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Initialize repositories
	userRepository := repository.NewPostgresUserRepository(dbSQL)

	// Initialize usecase
	userUsecase := usecase.NewUserUsecase(userRepository)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg, *userUsecase)

	// Register middleware
	e.Use(middleware.Logger())

	// Register routes
	e.POST("/auth/login", authHandler.Login)
	e.POST("/auth/register", authHandler.Register)

	// Start server
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
