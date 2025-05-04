package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port           string
	JWTSecret      string
	DatabaseURL    string
	DatabaseName   string `mapstructure:"database.name"`
	DatabaseUser   string `mapstructure:"database.user"`
	DatabasePassword string `mapstructure:"database.password"`
	DatabaseHost   string `mapstructure:"database.host"`
	DatabasePort   string `mapstructure:"database.port"`
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPass       string
	SMTPFrom       string
	SMTPFromName   string
	BaseURL        string
	PasswordReset  string
	Verification   string
	SSLMode        string `mapstructure:"database.sslmode"`
	DBHost         string `mapstructure:"database.host"`
	DBPort         int    `mapstructure:"database.port"`
	DBUser         string `mapstructure:"database.user"`
	DBPassword     string `mapstructure:"database.password"`
	DBName         string `mapstructure:"database.name"`
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

	// Initialize config struct
	cfg := &Config{}

	// Override with environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUTH")

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.SSLMode)
}
