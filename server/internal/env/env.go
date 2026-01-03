package env

import (
	"os"

	"download-server/internal/logger"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	env := godotenv.Load()

	if env != nil {
		logger.Logger.Fatal("Error loading .env file")
	}
}

func GetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		logger.Logger.Fatalf("Environment variable %s not set", key)
	}
	return value
}
