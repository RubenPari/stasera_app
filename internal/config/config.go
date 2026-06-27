package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
	AIGatewayBaseURL       string
	AIMaxTokens            int
	AITemperature          float64
	AITimeoutSeconds       int
	DBMaxOpen              int
	DBMaxIdle              int
	DBConnMaxLifetimeSec   int
}

// Load reads environment variables and returns a validated Config.
// It attempts to load a local .env file; if one is missing the error is ignored,
// but any other load error is logged.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file") {
			log.Printf("config: .env load: %v", err)
		}
	}

	cfg := &Config{
		Port:                   getEnv("PORT", "8080"),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		JWTSecret:              getEnv("JWT_SECRET", ""),
		JWTAccessExpiryMinutes: getEnvAsInt("JWT_ACCESS_EXPIRY_MINUTES", 15),
		JWTRefreshExpiryDays:   getEnvAsInt("JWT_REFRESH_EXPIRY_DAYS", 30),
		AIModel:                getEnv("AI_MODEL", "anthropic/claude-sonnet-4-20250514"),
		AIGatewayAPIKey:        getEnv("VERCEL_AI_GATEWAY_API_KEY", ""),
		AIGatewayBaseURL:       getEnv("AI_GATEWAY_BASE_URL", "https://ai-gateway.vercel.sh/v1"),
		AIMaxTokens:            getEnvAsInt("AI_MAX_TOKENS", 2000),
		AITemperature:          getEnvAsFloat("AI_TEMPERATURE", 0.7),
		AITimeoutSeconds:       getEnvAsInt("AI_TIMEOUT_SECONDS", 30),
		DBMaxOpen:              getEnvAsInt("DB_MAX_OPEN", 25),
		DBMaxIdle:              getEnvAsInt("DB_MAX_IDLE", 5),
		DBConnMaxLifetimeSec:   getEnvAsInt("DB_CONN_MAX_LIFETIME_SEC", 300),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if !strings.Contains(cfg.DatabaseURL, "parseTime=true") {
		return nil, fmt.Errorf("DATABASE_URL must include parseTime=true")
	}
	if cfg.JWTAccessExpiryMinutes <= 0 {
		return nil, fmt.Errorf("JWT_ACCESS_EXPIRY_MINUTES must be > 0")
	}
	if cfg.JWTRefreshExpiryDays <= 0 {
		return nil, fmt.Errorf("JWT_REFRESH_EXPIRY_DAYS must be > 0")
	}
	if cfg.AIGatewayAPIKey == "" {
		log.Println("config: VERCEL_AI_GATEWAY_API_KEY is not set — AI endpoints will fail")
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
		log.Printf("config: %s=%q is not a valid integer, using default %d", key, v, defaultValue)
		return defaultValue
	}
	return n
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		log.Printf("config: %s=%q is not a valid float, using default %g", key, v, defaultValue)
		return defaultValue
	}
	return f
}