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

	postgresPort, err := strconv.Atoi(os.Getenv("POSTGRES_PORT"))
	if err != nil {
		log.Fatalf("Invalid POSTGRES_PORT: %v", err)
	}

	kafkaBrokers := []string{}
	brokers := os.Getenv("KAFKA_BROKERS")
	kafkaBrokers = append(kafkaBrokers, splitAndTrim(brokers, ",")...)

	return &Config{
		Port:             os.Getenv("PORT"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     postgresPort,
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		KafkaBrokers:     kafkaBrokers,
		RedisAddr:        os.Getenv("REDIS_ADDR"),
		AWSAccessKey:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretKey:     os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSS3Bucket:      os.Getenv("AWS_S3_BUCKET"),
		AWSRegion:        os.Getenv("AWS_REGION"),
		LogLevel:         os.Getenv("LOG_LEVEL"),
	}
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
