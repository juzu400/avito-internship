package config

import (
	"log"
	"os"
)

// Config holds application configuration values.
type Config struct {
	HTTPAddr string
	DBDSN    string
	LogLevel string
}

// MustLoad loads configuration from environment variables and exits the application
// with a fatal log message if required values are missing.
func MustLoad() Config {
	cfg := Config{
		HTTPAddr: getenv("HTTP_ADDR", ":8080"),
		DBDSN:    getenv("DB_DSN", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		LogLevel: getenv("LOG_LEVEL", "debug"),
	}

	if cfg.HTTPAddr == "" {
		log.Fatal("HTTP_ADDR is required")
	}

	return cfg
}

// getenv returns the value of the environment variable key,
// or def if the variable is not set or empty.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
