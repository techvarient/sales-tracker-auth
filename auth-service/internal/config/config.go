package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Pass     string `mapstructure:"pass"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
}

type Config struct {
	Port          string         `mapstructure:"port"`
	Database      DatabaseConfig `mapstructure:"database"`
	JWTSecret     string         `mapstructure:"jwt_secret"`
	SMTP          SMTPConfig     `mapstructure:"smtp"`
	BaseURL       string         `mapstructure:"base_url"`
	PasswordReset string         `mapstructure:"password_reset_path"`
	Verification  string         `mapstructure:"verification_path"`
	DatabaseURL   string         // This will be constructed
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

	// Construct database URL from configuration
	if config.Database.Host != "" && config.Database.User != "" && config.Database.Name != "" {
		// Force IPv4 and add connection parameters
		config.DatabaseURL = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=10&sslmode=require&target_session_attrs=read-write",
			config.Database.User,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Name,
			config.Database.SSLMode,
		)
	} else if config.DatabaseURL == "" {
		return nil, fmt.Errorf("database configuration is incomplete - check config.yaml")
	}

	return &config, nil
}
