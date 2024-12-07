// internal/infrastructure/postgres/postgres.go

package postgres

import (
	"fmt"
	"log"

	"github.com/iSparshP/product-management-system/internal/domain/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func BuildDSN(cfg *Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
}

func NewPostgresDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Enable UUID extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")

	// Auto migrate models
	if err := db.AutoMigrate(&model.Product{}, &model.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}
