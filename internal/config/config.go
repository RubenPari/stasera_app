package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all environment-based configuration for the application.
type Config struct {
	Port                   string
	DatabaseURL            string
	JWTSecret              string
	JWTAccessExpiryMinutes int
	JWTRefreshExpiryDays   int
	AIModel                string
	AIGatewayAPIKey        string
}

// Load reads environment variables and returns a validated Config.
// It attempts to load a local .env file, but ignores the error if one is missing.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                   getEnv("PORT", "8080"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		JWTAccessExpiryMinutes: getEnvAsInt("JWT_ACCESS_EXPIRY_MINUTES", 15),
		JWTRefreshExpiryDays:   getEnvAsInt("JWT_REFRESH_EXPIRY_DAYS", 30),
		AIModel:                getEnv("AI_MODEL", "anthropic/claude-sonnet-4-20250514"),
		AIGatewayAPIKey:        os.Getenv("VERCEL_AI_GATEWAY_API_KEY"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}
