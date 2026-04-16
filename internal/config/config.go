package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	LogLevel   string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env not found. Default values will be used.")
	}

	return &Config{
		Port:       getEnv("PORT", ":8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "sub_db"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
