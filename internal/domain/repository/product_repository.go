// internal/domain/repository/product_repository.go

package repository

import (
	"context"

	"github.com/iSparshP/product-management-system/internal/domain/model"
)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, id string) (*model.Product, error)
	GetAll(ctx context.Context, userID string, filters map[string]interface{}) ([]model.Product, error)
	UpdateCompressedImages(ctx context.Context, id string, images []string) error
}
