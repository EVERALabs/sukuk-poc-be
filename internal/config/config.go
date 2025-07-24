package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Blockchain BlockchainConfig
	API        APIConfig
	Logger     LoggerConfig
	Email      EmailConfig // Low priority
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
	Port        int
	Debug       bool
	UploadDir   string
	MaxFileSize int64 // in bytes
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type IndexerDatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type BlockchainConfig struct {
	ChainID         int64  // Base Testnet: 84532
	RPCEndpoint     string // Base Testnet RPC
	WebSocketURL    string // Base Testnet WebSocket
	ContractAddress string // Your Sukuk contract
	StartBlock      int64  // Block to start indexing from
}

type APIConfig struct {
	APIKey          string
	RateLimitPerMin int
	AllowedOrigins  []string
	WebhookSecret   string
}

type LoggerConfig struct {
	Level  string
	Format string
}

type EmailConfig struct {
	Enabled  bool
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	config := &Config{}

	// App configuration
	config.App = AppConfig{
		Name:        getEnv("APP_NAME", "sukuk-poc-api"),
		Version:     getEnv("APP_VERSION", "1.0.0"),
		Environment: getEnv("APP_ENV", "development"),
		Port:        getEnvAsInt("APP_PORT", 8080),
		Debug:       getEnvAsBool("APP_DEBUG", true),
		UploadDir:   getEnv("APP_UPLOAD_DIR", "./uploads"),
		MaxFileSize: getEnvAsInt64("APP_MAX_FILE_SIZE", 10485760), // 10MB default
	}

	// Database configuration
	config.Database = DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnvAsInt("DB_PORT", 5432),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "sukuk_poc"),
		SSLMode:         getEnv("DB_SSL_MODE", "disable"),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", time.Hour),
	}

	// Blockchain configuration (Base Testnet defaults)
	config.Blockchain = BlockchainConfig{
		ChainID:         getEnvAsInt64("BLOCKCHAIN_CHAIN_ID", 84532), // Base Testnet
		RPCEndpoint:     getEnv("BLOCKCHAIN_RPC_ENDPOINT", "https://sepolia.base.org"),
		WebSocketURL:    getEnv("BLOCKCHAIN_WEBSOCKET_URL", "wss://sepolia.base.org"),
		ContractAddress: getEnv("BLOCKCHAIN_CONTRACT_ADDRESS", ""),
		StartBlock:      getEnvAsInt64("BLOCKCHAIN_START_BLOCK", 0),
	}

	// API configuration
	config.API = APIConfig{
		APIKey:          getEnv("API_API_KEY", ""),
		RateLimitPerMin: getEnvAsInt("API_RATE_LIMIT_PER_MIN", 100),
		AllowedOrigins:  getEnvAsSlice("API_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		WebhookSecret:   getEnv("API_WEBHOOK_SECRET", ""),
	}

	// Logger configuration
	config.Logger = LoggerConfig{
		Level:  getEnv("LOGGER_LEVEL", "info"),
		Format: getEnv("LOGGER_FORMAT", "json"),
	}

	// Email configuration (disabled by default)
	config.Email = EmailConfig{
		Enabled:  getEnvAsBool("EMAIL_ENABLED", false),
		Host:     getEnv("EMAIL_HOST", "smtp.gmail.com"),
		Port:     getEnvAsInt("EMAIL_PORT", 587),
		User:     getEnv("EMAIL_USER", ""),
		Password: getEnv("EMAIL_PASSWORD", ""),
		From:     getEnv("EMAIL_FROM", "noreply@sukuk-poc.com"),
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.App.Port <= 0 || config.App.Port > 65535 {
		return fmt.Errorf("invalid port: %d", config.App.Port)
	}

	if config.Blockchain.ChainID != 84532 {
		return fmt.Errorf("this project is configured for Base Testnet (chain ID 84532), got: %d", config.Blockchain.ChainID)
	}

	if config.API.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	return nil
}

// Helper functions
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsInt64(key string, defaultVal int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsSlice(key string, defaultVal []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	return strings.Split(valueStr, ",")
}