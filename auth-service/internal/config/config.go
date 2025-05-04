package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Port           string
	JWTSecret      string
	DatabaseURL    string
	DatabaseName   string
	DatabaseUser   string
	DatabasePassword string
	DatabaseHost   string
	DatabasePort   string
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPass       string
	SMTPFrom       string
	SMTPFromName   string
	BaseURL        string
	PasswordReset  string
	Verification   string
}

func NewConfig() (*Config, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found")
	}

	// Initialize Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Initialize config struct
	cfg := &Config{}

	// Override with environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUTH")

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Build DatabaseURL
	cfg.DatabaseURL = "postgresql://" + cfg.DatabaseUser + ":" + cfg.DatabasePassword +
		"@" + cfg.DatabaseHost + ":" + cfg.DatabasePort + "/" + cfg.DatabaseName

	return cfg, nil
}
