/*
 * Contract ID: CTR-006
 * Service Name: SupermarketService
 * Description: Configuration parsing, environment variables, and safe defaults.
 */

package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration properties for the SupermarketService.
// Implements Contract CTR-006.
type Config struct {
	// Server Port
	PORT string

	// Deployment environment (development, staging, production)
	APP_ENV string

	// Comma-separated list of origins allowed to make cross-origin requests.
	// Empty means CORS is disabled (no Access-Control-Allow-Origin header is sent).
	ALLOWED_ORIGINS string

	// Database Connection String
	MYSQL_DSN string

	// JWT signing secret
	JWT_SECRET string

	// Log Level (debug, info, warn, error)
	LOG_LEVEL string

	// Database Pool Configurations (Production-ready enhancements)
	DB_MAX_OPEN_CONNS         int
	DB_MAX_IDLE_CONNS         int
	DB_CONN_MAX_LIFETIME_MINS int

	// Ristretto Cache Configurations
	CACHE_NUM_COUNTERS int64
	CACHE_MAX_COST     int64 // Upper memory allocation limit in bytes
	CACHE_BUFFER_ITEMS int64 // Write concurrency optimization
}

// Load loads the configuration from environment variables and an optional .env file.
func Load() (*Config, error) {
	// Attempt to load .env file if it exists in the runtime directory
	_ = loadDotEnv(".env")

	// Parse database pool parameters from env or set robust defaults
	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 100)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetime := getEnvAsInt("DB_CONN_MAX_LIFETIME_MINS", 30)

	// Parse Ristretto parameters from env or set defaults per LLD Section 9
	cacheNumCounters := getEnvAsInt64("CACHE_NUM_COUNTERS", 10000)
	cacheMaxCost := getEnvAsInt64("CACHE_MAX_COST", 100*1024*1024) // 100MB in bytes
	cacheBufferItems := getEnvAsInt64("CACHE_BUFFER_ITEMS", 64)

	cfg := &Config{
		PORT:                      getEnv("PORT", "8080"),
		APP_ENV:                   getEnv("APP_ENV", "development"),
		ALLOWED_ORIGINS:           getEnv("ALLOWED_ORIGINS", ""),
		MYSQL_DSN:                 getEnv("MYSQL_DSN", ""),
		JWT_SECRET:                getEnv("JWT_SECRET", ""),
		LOG_LEVEL:                 getEnv("LOG_LEVEL", "info"),
		DB_MAX_OPEN_CONNS:         maxOpenConns,
		DB_MAX_IDLE_CONNS:         maxIdleConns,
		DB_CONN_MAX_LIFETIME_MINS: connMaxLifetime,
		CACHE_NUM_COUNTERS:        cacheNumCounters,
		CACHE_MAX_COST:            cacheMaxCost,
		CACHE_BUFFER_ITEMS:        cacheBufferItems,
	}

	// Validate required fields
	if strings.TrimSpace(cfg.PORT) == "" {
		return nil, fmt.Errorf("configuration error: PORT must not be empty")
	}

	if cfg.MYSQL_DSN == "" {
		return nil, fmt.Errorf("configuration error: MYSQL_DSN is required (e.g. root:password@tcp(127.0.0.1:3306)/supermarket_supermarket?charset=utf8mb4&parseTime=True&loc=Local)")
	}

	if cfg.JWT_SECRET == "" {
		// Provide a dev-safe default fallback if in development, otherwise error out in production
		if cfg.APP_ENV == "production" {
			return nil, fmt.Errorf("configuration error: JWT_SECRET is required in production environment")
		}
		// ⚠ Fallback dev secret (dev-safe fallback)
		cfg.JWT_SECRET = "supermarket-development-secret-key-change-in-production-32chars"
	}

	return cfg, nil
}

// GetJWTExpiration returns the token expiration duration, hardcoded to 24 hours per SAD Section 5.1
func (c *Config) GetJWTExpiration() time.Duration {
	return 24 * time.Hour
}

// GetJWTSecret returns the configured JWT signing secret.
func (c *Config) GetJWTSecret() string {
	return c.JWT_SECRET
}

// Helper to lookup string environment variables
func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

// Helper to parse integer environment variables
func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

// Helper to parse int64 environment variables
func getEnvAsInt64(key string, defaultVal int64) int64 {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return defaultVal
	}
	return val
}

// loadDotEnv loads key-value pairs from a .env file into the environment
func loadDotEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines or comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Strip optional surrounding quotes
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}

		// Only set if not already defined in env
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}

	return scanner.Err()
}
