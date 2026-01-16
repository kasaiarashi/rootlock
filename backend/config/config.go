package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	WebAuthn WebAuthnConfig
	Security SecurityConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port        string
	Host        string
	Environment string // "development", "production"
	CORSOrigins []string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// WebAuthnConfig holds WebAuthn configuration
type WebAuthnConfig struct {
	RPID      string
	RPName    string
	RPOrigins []string
	Timeout   time.Duration
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	BcryptCost         int
	RateLimitEnabled   bool
	RateLimitPerMinute int
	Argon2Memory       uint32
	Argon2Time         uint32
	Argon2Parallelism  uint8
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Environment: getEnv("ENV", "development"),
			CORSOrigins: getEnvSlice("CORS_ORIGINS", []string{"http://localhost:3000"}),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgresql://rootlock:rootlock_dev_password@localhost:5432/rootlock?sslmode=disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenDuration:  getEnvDuration("JWT_ACCESS_DURATION", 15*time.Minute),
			RefreshTokenDuration: getEnvDuration("JWT_REFRESH_DURATION", 30*24*time.Hour),
		},
		WebAuthn: WebAuthnConfig{
			RPID:      getEnv("WEBAUTHN_RP_ID", "localhost"),
			RPName:    getEnv("WEBAUTHN_RP_NAME", "RootLock Password Manager"),
			RPOrigins: getEnvSlice("WEBAUTHN_RP_ORIGINS", []string{"http://localhost:8080"}),
			Timeout:   getEnvDuration("WEBAUTHN_TIMEOUT", 60*time.Second),
		},
		Security: SecurityConfig{
			BcryptCost:         getEnvInt("BCRYPT_COST", 12),
			RateLimitEnabled:   getEnvBool("RATE_LIMIT_ENABLED", true),
			RateLimitPerMinute: getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
			Argon2Memory:       uint32(getEnvInt("ARGON2_MEMORY", 65536)), // 64MB
			Argon2Time:         uint32(getEnvInt("ARGON2_TIME", 3)),
			Argon2Parallelism:  uint8(getEnvInt("ARGON2_PARALLELISM", 4)),
		},
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}

	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.JWT.Secret == "" || c.JWT.Secret == "your-secret-key-change-in-production" {
		if c.Server.Environment == "production" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
	}

	if c.WebAuthn.RPID == "" {
		return fmt.Errorf("WEBAUTHN_RP_ID is required")
	}

	if c.Security.BcryptCost < 10 || c.Security.BcryptCost > 14 {
		return fmt.Errorf("BCRYPT_COST must be between 10 and 14")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		return parseCommaSeparated(value)
	}
	return defaultValue
}

func parseCommaSeparated(s string) []string {
	var result []string
	current := ""
	for _, char := range s {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
