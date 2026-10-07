package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// Config contains process configuration only. Optional database and JWT values
// deliberately do not prevent the Phase 1 operational server from starting.
type Config struct {
	AppEnv      string
	AppPort     int
	FrontendURL string
	DatabaseURL string
	JWTSecret   string
	StorageDir  string
}

func Load() (Config, error) {
	port := 8080
	if rawPort, ok := os.LookupEnv("APP_PORT"); ok && rawPort != "" {
		parsedPort, err := strconv.Atoi(rawPort)
		if err != nil || parsedPort < 1 || parsedPort > 65535 {
			return Config{}, fmt.Errorf("APP_PORT must be an integer between 1 and 65535")
		}
		port = parsedPort
	}

	frontendURL := envOr("FRONTEND_URL", "http://localhost:3000")
	parsedFrontendURL, err := url.Parse(frontendURL)
	if err != nil || parsedFrontendURL.Scheme == "" || parsedFrontendURL.Host == "" {
		return Config{}, fmt.Errorf("FRONTEND_URL must be an absolute HTTP(S) URL")
	}
	if parsedFrontendURL.Scheme != "http" && parsedFrontendURL.Scheme != "https" {
		return Config{}, fmt.Errorf("FRONTEND_URL must use HTTP or HTTPS")
	}

	return Config{
		AppEnv:      envOr("APP_ENV", "development"),
		AppPort:     port,
		FrontendURL: frontendURL,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		StorageDir:  envOr("STORAGE_DIR", "./private-storage"),
	}, nil
}

func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
