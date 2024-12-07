// internal/domain/model/product.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Product struct {
	ID                      uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID                  uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	ProductName             string         `gorm:"type:varchar(255);not null" json:"product_name"`
	ProductDescription      string         `gorm:"type:text" json:"product_description"`
	ProductImages           datatypes.JSON `gorm:"type:jsonb;not null" json:"product_images"`
	CompressedProductImages datatypes.JSON `gorm:"type:jsonb" json:"compressed_product_images"`
	ProductPrice            float64        `gorm:"type:decimal(10,2);not null" json:"product_price"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

// CreateProductInput represents the input payload for creating a product.
type CreateProductInput struct {
	UserID             string   `json:"user_id" binding:"required,uuid"`
	ProductName        string   `json:"product_name" binding:"required"`
	ProductDescription string   `json:"product_description" binding:"required"`
	ProductImages      []string `json:"product_images" binding:"required,min=1,dive,url"`
	ProductPrice       float64  `json:"product_price" binding:"required,gt=0"`
}

// ImageProcessingTask represents the task for processing images.
type ImageProcessingTask struct {
	ProductID string   `json:"product_id"`
	ImageURLs []string `json:"image_urls"`
}
