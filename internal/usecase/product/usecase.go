// internal/usecase/product/usecase.go

package product

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iSparshP/product-management-system/internal/domain/model"
	"github.com/iSparshP/product-management-system/internal/domain/repository"
	"github.com/iSparshP/product-management-system/internal/infrastructure/kafka"
	"github.com/iSparshP/product-management-system/internal/infrastructure/redis"
	"github.com/iSparshP/product-management-system/pkg/utils"
	"go.uber.org/zap"
)

type Usecase interface {
	CreateProduct(ctx context.Context, input model.CreateProductInput) (*model.Product, error)
	GetProductByID(ctx context.Context, id string) (*model.Product, error)
	GetProducts(ctx context.Context, userID string, filters map[string]interface{}) ([]model.Product, error)
}

type usecase struct {
	repo        repository.ProductRepository
	kafkaPub    *kafka.Publisher
	redisClient *redis.Client
	logger      *zap.Logger
}

func NewProductUsecase(repo repository.ProductRepository, kafkaPub *kafka.Publisher, redisClient *redis.Client, logger *zap.Logger) Usecase {
	return &usecase{
		repo:        repo,
		kafkaPub:    kafkaPub,
		redisClient: redisClient,
		logger:      logger,
	}
}

func (u *usecase) CreateProduct(ctx context.Context, input model.CreateProductInput) (*model.Product, error) {
	// Validate user_id
	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		u.logger.Error("Invalid user ID format", zap.String("user_id", input.UserID), zap.Error(err))
		return nil, fmt.Errorf("invalid user_id format: %w", err)
	}

	// Create Product
	product := &model.Product{
		ID:                 uuid.New(),
		UserID:             userUUID,
		ProductName:        input.ProductName,
		ProductDescription: input.ProductDescription,
		ProductImages:      utils.StringSliceToJSON(input.ProductImages),
		ProductPrice:       input.ProductPrice,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Save to Database
	if err := u.repo.Create(ctx, product); err != nil {
		u.logger.Error("Failed to create product in database",
			zap.Error(err),
			zap.String("product_id", product.ID.String()),
			zap.String("user_id", userUUID.String()))
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Publish to Kafka for image processing
	task := model.ImageProcessingTask{
		ProductID: product.ID.String(),
		ImageURLs: input.ProductImages,
	}

	taskData, err := json.Marshal(task)
	if err != nil {
		u.logger.Error("Failed to marshal image processing task",
			zap.Error(err),
			zap.String("product_id", product.ID.String()))
		// Continue execution as image processing is not critical for product creation
	} else {
		// Only attempt to publish if marshaling succeeded
		if err := u.kafkaPub.Publish(ctx, taskData); err != nil {
			u.logger.Error("Failed to publish image processing task",
				zap.Error(err),
				zap.String("product_id", product.ID.String()))
			// Consider implementing retry logic here
			// For now, we'll continue as image processing is not critical
		}
	}

	return product, nil
}

func (u *usecase) GetProductByID(ctx context.Context, id string) (*model.Product, error) {
	// Check Redis Cache
	cacheKey := fmt.Sprintf("product:%s", id)
	cachedProduct, err := u.redisClient.Get(ctx, cacheKey)
	if err == nil {
		var product model.Product
		if err := json.Unmarshal([]byte(cachedProduct), &product); err == nil {
			u.logger.Debug("Cache hit for product", zap.String("id", id))
			return &product, nil
		}
		u.logger.Warn("Failed to unmarshal cached product",
			zap.Error(err),
			zap.String("id", id))
	}

	// Fetch from Database
	product, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.logger.Error("Failed to get product by ID",
			zap.Error(err),
			zap.String("id", id))
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Cache the Product
	productData, err := json.Marshal(product)
	if err == nil {
		if err := u.redisClient.Set(ctx, cacheKey, productData, 10*time.Minute); err != nil {
			u.logger.Warn("Failed to cache product",
				zap.Error(err),
				zap.String("id", id))
		}
	}

	return product, nil
}

func (u *usecase) GetProducts(ctx context.Context, userID string, filters map[string]interface{}) ([]model.Product, error) {
	products, err := u.repo.GetAll(ctx, userID, filters)
	if err != nil {
		u.logger.Error("Failed to get products",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.Any("filters", filters))
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	return products, nil
}
