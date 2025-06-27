package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Environment string
	DatabaseURL string
	RedisURL    string
	BaseURL     string
}

func Load() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/urlshortener?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
