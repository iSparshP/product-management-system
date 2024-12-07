// cmd/api/main.go

package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/iSparshP/product-management-system/internal/api/handler"
	"github.com/iSparshP/product-management-system/internal/api/router"
	"github.com/iSparshP/product-management-system/internal/infrastructure/config"
	"github.com/iSparshP/product-management-system/internal/infrastructure/kafka"
	"github.com/iSparshP/product-management-system/internal/infrastructure/logger"
	"github.com/iSparshP/product-management-system/internal/infrastructure/postgres"
	"github.com/iSparshP/product-management-system/internal/infrastructure/redis"
	"github.com/iSparshP/product-management-system/internal/usecase/product"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Logger
	logInstance, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logInstance.Sync()

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

	// Initialize Redis
	redisClient := redis.NewRedisClient(cfg.RedisAddr)

	// Initialize Kafka Publisher
	kafkaPub, err := kafka.NewPublisher(cfg.KafkaBrokers, "image_processing", logInstance)
	if err != nil {
		logInstance.Fatal("Failed to initialize Kafka publisher", zap.Error(err))
	}

	// Initialize Repositories
	productRepo := postgres.NewProductRepo(db)

	// Initialize Usecases
	productUsecase := product.NewProductUsecase(productRepo, kafkaPub, redisClient, logInstance)

	// Initialize Handlers
	productHandler := handler.NewProductHandler(productUsecase, logInstance)

	// Setup Router
	r := router.SetupRouter(productHandler, logInstance)

	// Start Server
	go func() {
		if err := r.Run(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			logInstance.Fatal("Failed to run server", zap.Error(err))
		}
	}()
	logInstance.Info("API server started", zap.String("port", cfg.Port))

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logInstance.Info("Shutting down server...")

	// Here you can add graceful shutdown logic if needed
}
