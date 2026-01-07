package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Load env vars from .env file directly
func init() {
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist (e.g. in production),
		// but we should log it just in case.
		// However, mostly we want to rely on environment variables being set.
		// If we are in local dev, this helps.
		log.Println("No .env file found or error loading it, using system environment variables")
	}
}

// GetAllowedOrigins returns a slice of allowed origins from the environment variable.
// It defaults to localhost:3000 if not set.
func GetAllowedOrigins() []string {
	originsStr := os.Getenv("ALLOWED_ORIGINS")
	if originsStr == "" {
		return []string{"http://localhost:3000"}
	}

	// Split by comma and trim spaces
	parts := strings.Split(originsStr, ",")
	var origins []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
