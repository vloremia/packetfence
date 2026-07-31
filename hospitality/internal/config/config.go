package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort           string
	DatabaseURL        string
	SecretKey          string
	RateLimitPerMinute int
	GracePeriod        time.Duration
	LogLevel           string
	Environment        string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:           getEnv("HOSPITALITY_PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://hospitality:hospitality@localhost:5432/hospitality"),
		SecretKey:          getEnv("HOSPITALITY_SECRET_KEY", ""),
		RateLimitPerMinute: getEnvInt("HOSPITALITY_RATE_LIMIT_PER_MINUTE", 120),
		GracePeriod:        time.Duration(getEnvInt("HOSPITALITY_GRACE_PERIOD_MINUTES", 120)) * time.Minute,
		LogLevel:           getEnv("HOSPITALITY_LOG_LEVEL", "info"),
		Environment:        getEnv("HOSPITALITY_ENV", "development"),
	}
	if cfg.SecretKey == "" && cfg.Environment == "production" {
		return nil, fmt.Errorf("HOSPITALITY_SECRET_KEY must be set in production")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
