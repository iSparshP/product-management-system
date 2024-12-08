// internal/infrastructure/config/config.go

package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	KafkaBrokers     []string
	RedisAddr        string
	AWSAccessKey     string
	AWSSecretKey     string
	AWSS3Bucket      string
	AWSRegion        string
	AWSEndpoint      string
	LogLevel         string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	// Try multiple possible locations for .env file
	envFiles := []string{
		"configs/.env",
		"../configs/.env",
		"./configs/.env",
		"/app/configs/.env",
	}

	var loaded bool
	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			loaded = true
			log.Printf("Loaded config from %s", file)
			break
		}
	}

	if !loaded {
		log.Printf("No .env file found, using environment variables")
	}

	// Log all environment variables for debugging (be careful with sensitive data in production)
	log.Printf("Environment Variables:")
	log.Printf("AWS_ACCESS_KEY_ID set: %v", os.Getenv("AWS_ACCESS_KEY_ID") != "")
	log.Printf("AWS_SECRET_ACCESS_KEY set: %v", os.Getenv("AWS_SECRET_ACCESS_KEY") != "")
	log.Printf("AWS_S3_BUCKET: %s", os.Getenv("AWS_S3_BUCKET"))
	log.Printf("AWS_REGION: %s", os.Getenv("AWS_REGION"))
	log.Printf("AWS_ENDPOINT: %s", os.Getenv("AWS_ENDPOINT"))

	config := &Config{
		Port:             getEnvOrDefault("PORT", "8080"),
		PostgresHost:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnvAsIntOrDefault("POSTGRES_PORT", 5432),
		PostgresUser:     getEnvOrDefault("POSTGRES_USER", "youruser"),
		PostgresPassword: getEnvOrDefault("POSTGRES_PASSWORD", "yourpassword"),
		PostgresDB:       getEnvOrDefault("POSTGRES_DB", "productdb"),
		KafkaBrokers:     strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		RedisAddr:        getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		AWSAccessKey:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretKey:     os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSS3Bucket:      os.Getenv("AWS_S3_BUCKET"),
		AWSRegion:        os.Getenv("AWS_REGION"),
		AWSEndpoint:      getEnvOrDefault("AWS_ENDPOINT", "https://s3.amazonaws.com"),
		LogLevel:         getEnvOrDefault("LOG_LEVEL", "info"),
	}

	// Validate required AWS configuration
	if config.AWSAccessKey == "" || config.AWSSecretKey == "" || config.AWSS3Bucket == "" || config.AWSRegion == "" {
		log.Printf("Warning: Missing required AWS configuration")
		log.Printf("AWS_ACCESS_KEY_ID set: %v", config.AWSAccessKey != "")
		log.Printf("AWS_SECRET_ACCESS_KEY set: %v", config.AWSSecretKey != "")
		log.Printf("AWS_S3_BUCKET: %s", config.AWSS3Bucket)
		log.Printf("AWS_REGION: %s", config.AWSRegion)
	}

	return config
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("Invalid %s: %v", key, err)
	}
	return intValue
}
