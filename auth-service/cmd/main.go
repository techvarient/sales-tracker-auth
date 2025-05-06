package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/sales-tracker/auth-service/internal/config"
	"github.com/sales-tracker/auth-service/internal/handler"
	"github.com/sales-tracker/auth-service/internal/repository"
	"github.com/sales-tracker/auth-service/internal/service"
	"github.com/sales-tracker/auth-service/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var migrationsDir = "migrations"

type migrationFile struct {
	Name string
	Path string
}

func runMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Read migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []migrationFile
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			migrations = append(migrations, migrationFile{
				Name: file.Name(),
				Path: filepath.Join("migrations", file.Name()),
			})
		}
	}

	// Sort migrations by name
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	// Apply migrations
	for _, migration := range migrations {
		// Extract version from filename (e.g., 000001_initial.up.sql -> 1)
		var version int64
		_, err := fmt.Sscanf(migration.Name, "%d_", &version)
		if err != nil {
			return fmt.Errorf("invalid migration filename format: %s", migration.Name)
		}

		// Check if migration is already applied
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count == 0 {
			// Read migration file
			content, err := os.ReadFile(filepath.Join(migrationsDir, migration.Name))
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w", migration.Name, err)
			}

			// Start transaction
			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			// Execute migration
			if _, err := tx.Exec(string(content)); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %w", migration.Name, err)
			}

			// Record migration
			if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to record migration: %w", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}

			log.Printf("Applied migration: %s\n", migration.Name)
		}
	}

	return nil
}

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

	// Run migrations
	log.Println("Running database migrations...")
	if err := runMigrations(dbSQL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize repositories
	userRepository := repository.NewPostgresUserRepository(dbSQL)

	// Initialize usecase
	userUsecase := usecase.NewUserUsecase(userRepository)

	// Initialize email service
	emailService := service.NewSMTPService(cfg)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg, *userUsecase, emailService)

	// Register middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{}))
	// Add CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Register routes
	e.POST("/auth/login", authHandler.Login)
	e.POST("/auth/register", authHandler.Register)
	e.GET("/auth/verify", authHandler.VerifyEmail)
	e.POST("/auth/resend-verification", authHandler.ResendVerificationEmail)
	e.POST("/auth/reset-password", authHandler.ResetPassword)
	e.POST("/auth/forgot-password", authHandler.ForgotPassword)

	// Start server
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
