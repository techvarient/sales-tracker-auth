package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/sales-tracker/auth-service/internal/config"
	"github.com/sales-tracker/auth-service/internal/handler"
	"github.com/sales-tracker/auth-service/internal/middleware"
	"github.com/sales-tracker/auth-service/internal/repository"
	"github.com/sales-tracker/auth-service/internal/usecase"
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

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepository := repository.NewPostgresUserRepository(db)

	// Initialize usecase
	userUsecase := usecase.NewUserUsecase(userRepository)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg, *userUsecase)

	// Register middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Register routes
	api := e.Group("/api")
	auth := api.Group("/auth")

	// Public routes
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/forgot-password", authHandler.ForgotPassword)
	auth.POST("/reset-password", authHandler.ResetPassword)

	// Protected routes
	protected := api.Group("", middleware.JWTMiddleware(cfg))
	protected.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Start server with graceful shutdown
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: e,
	}

	// Create a context that listens for the interrupt signal from the OS
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Make the server listen on the specified address
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server shutdown: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown with a timeout
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
