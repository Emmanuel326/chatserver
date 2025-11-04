package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application-wide configuration
type Config struct {
	DB_FILE      string // The path to the SQLite DB file
	
	JWT_SECRET   string
	JWT_EXPIRY   int // in minutes

	SERVER_PORT  string
}

// Load loads configuration from environment variables (or a .env file)
func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: No .env file found. Reading config from environment.")
	}

	getEnv := func(key string, defaultValue string) string {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
		return defaultValue
	}

	expiryStr := getEnv("JWT_EXPIRY_MINUTES", "10080")
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		log.Printf("Warning: Invalid JWT_EXPIRY_MINUTES (%s). Defaulting to 10080.\n", expiryStr)
		expiry = 10080
	}

	return &Config{
		// Database
		DB_FILE:     getEnv("DB_FILE", "chatserver.db"),
		
		// JWT
		JWT_SECRET:  getEnv("JWT_SECRET", "default_secret"),
		JWT_EXPIRY:  expiry,
		
		// Server
		SERVER_PORT: getEnv("SERVER_PORT", "8080"),
	}
}
