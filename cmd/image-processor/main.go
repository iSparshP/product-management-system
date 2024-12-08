// cmd/image-processor/main.go

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/iSparshP/product-management-system/internal/imageprocessor/service"
	"github.com/iSparshP/product-management-system/internal/infrastructure/config"
	"github.com/iSparshP/product-management-system/internal/infrastructure/kafka"
	"github.com/iSparshP/product-management-system/internal/infrastructure/logger"
	"github.com/iSparshP/product-management-system/internal/infrastructure/postgres"
)

func main() {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Debug: Print raw environment variables
	log.Printf("Raw AWS_ACCESS_KEY_ID: %v", os.Getenv("AWS_ACCESS_KEY_ID"))
	log.Printf("Raw AWS_SECRET_ACCESS_KEY length: %d", len(os.Getenv("AWS_SECRET_ACCESS_KEY")))
	log.Printf("Raw AWS_REGION: %v", os.Getenv("AWS_REGION"))
	log.Printf("Raw AWS_S3_BUCKET: %v", os.Getenv("AWS_S3_BUCKET"))

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Logger
	logInstance, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logInstance.Sync()

	logInstance.Info("AWS Configuration",
		zap.String("AWS_REGION", cfg.AWSRegion),
		zap.String("AWS_S3_BUCKET", cfg.AWSS3Bucket),
		zap.String("AWS_ENDPOINT", cfg.AWSEndpoint),
		zap.Bool("AWS_ACCESS_KEY_SET", cfg.AWSAccessKey != ""),
		zap.Bool("AWS_SECRET_KEY_SET", cfg.AWSSecretKey != ""),
	)

	// Initialize PostgreSQL
	pgConfig := &postgres.Config{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		DBName:   cfg.PostgresDB,
	}
	dsn := postgres.BuildDSN(pgConfig)
	db := postgres.NewPostgresDB(dsn)

	// Initialize Kafka Consumer
	kafkaConsumer, err := kafka.NewConsumer(cfg.KafkaBrokers, "image_processing_group", "image_processing", logInstance)
	if err != nil {
		logInstance.Fatal("Failed to initialize Kafka consumer", zap.Error(err))
	}

	// Initialize Repositories
	productRepo := postgres.NewProductRepo(db)

	// Initialize Image Processor Service
	imgProcessor := service.NewImageProcessor(kafkaConsumer, productRepo, cfg, logInstance)

	// Start Image Processor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := imgProcessor.Start(ctx); err != nil {
			logInstance.Fatal("Image processor encountered an error", zap.Error(err))
		}
	}()
	logInstance.Info("Image Processor started")

	// Wait for interrupt signal to gracefully shutdown the service
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logInstance.Info("Shutting down Image Processor...")
}
