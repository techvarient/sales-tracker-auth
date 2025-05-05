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
	viper.AddConfigPath(".") // Look in the current directory

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Generate PostgresDSN
	if config.DBHost == "" {
		config.DBHost = "localhost"
	}
	if config.DBPort == 0 {
		config.DBPort = 5432
	}
	if config.DBUser == "" {
		config.DBUser = "postgres"
	}
	if config.DBName == "" {
		config.DBName = "auth_db"
	}

	config.DatabaseURL = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.DBHost,
		config.DBPort,
		config.DBUser,
		config.DBPassword,
		config.DBName,
		config.SSLMode,
	)
	if config.DatabaseURL == "" {
		config.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)
	}

	return &config, nil
}
