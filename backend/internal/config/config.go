package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

// Load reads configuration from environment variables and returns a Config.
// It returns an error if any required variable is missing.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		Port:        port,
	}, nil
}
