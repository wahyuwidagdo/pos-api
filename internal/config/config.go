package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	DBUrl       string
	JWTSecret   string
	CORSOrigins string
	Environment string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dsn := getEnv("DATABASE_URL", "")
	if dsn == "" {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", ""),
			getEnv("DB_NAME", "pos_db"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_SSL_MODE", "disable"),
		)
	}

	return &Config{
		AppPort:     getEnv("PORT", "3000"),
		DBUrl:       dsn,
		JWTSecret:   getEnv("JWT_SECRET", "verysecretkey"), // Default for dev, warn in prod
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:5174"),
		Environment: getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (c *Config) GetCORSOrigins() []string {
	return strings.Split(c.CORSOrigins, ",")
}

func (c *Config) GetPortInt() int {
	port, _ := strconv.Atoi(c.AppPort)
	return port
}
