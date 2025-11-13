package config

import (
	"log"
	"os"
)

type Config struct {
	HTTPAddr string
	DBDSN    string
	LogLevel string
}

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

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
