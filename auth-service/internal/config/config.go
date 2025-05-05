package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port          string
	JWTSecret     string
	DatabaseURL   string
	DBName        string `mapstructure:"database.name"`
	DBUser        string `mapstructure:"database.user"`
	DBPassword    string `mapstructure:"database.password"`
	DBHost        string `mapstructure:"database.host"`
	DBPort        int    `mapstructure:"database.port"`
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	SMTPFrom      string
	SMTPFromName  string
	BaseURL       string
	PasswordReset string
	Verification  string
	SSLMode       string `mapstructure:"database.sslmode"`
}

func NewConfig() (*Config, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found")
	}

	// Initialize Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("internal/config") // Look two levels up from the current directory

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Generate PostgresDSN if not provided
	if config.DatabaseURL == "" {
		config.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)
	}

	return &config, nil
}
