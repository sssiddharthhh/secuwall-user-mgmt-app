package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBPath    string
	JWTSecret string
	JWTExpiry time.Duration
}

func Load() *Config {
	// .env is optional; falls back to environment variables and defaults.
	_ = godotenv.Load()

	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBPath:    getEnv("DB_PATH", "./data/users.db"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry: 24 * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
