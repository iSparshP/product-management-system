// internal/infrastructure/postgres/product_repository.go

package postgres

import (
	"context"
	"errors"

	"github.com/iSparshP/product-management-system/internal/domain/model"
	"github.com/iSparshP/product-management-system/internal/domain/repository"
	"gorm.io/gorm"
)

type ProductRepo struct {
	DB *gorm.DB
}

func NewProductRepo(db *gorm.DB) repository.ProductRepository {
	return &ProductRepo{
		DB: db,
	}
}

func (r *ProductRepo) Create(ctx context.Context, product *model.Product) error {
	return r.DB.WithContext(ctx).Create(product).Error
}

func (r *ProductRepo) GetByID(ctx context.Context, id string) (*model.Product, error) {
	var product model.Product
	if err := r.DB.WithContext(ctx).First(&product, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepo) GetAll(ctx context.Context, userID string, filters map[string]interface{}) ([]model.Product, error) {
	var products []model.Product
	query := r.DB.WithContext(ctx).Where("user_id = ?", userID)

	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("product_name ILIKE ?", "%"+name+"%")
	}

	if minPrice, ok := filters["min_price"].(float64); ok && minPrice > 0 {
		query = query.Where("product_price >= ?", minPrice)
	}

	if maxPrice, ok := filters["max_price"].(float64); ok && maxPrice > 0 {
		query = query.Where("product_price <= ?", maxPrice)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductRepo) UpdateCompressedImages(ctx context.Context, id string, images []string) error {
	return r.DB.WithContext(ctx).Model(&model.Product{}).Where("id = ?", id).
		Update("compressed_product_images", images).Error
}
