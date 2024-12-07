// internal/imageprocessor/service/image_processor.go

package service

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/iSparshP/product-management-system/internal/domain/model"
	"github.com/iSparshP/product-management-system/internal/domain/repository"
	"github.com/iSparshP/product-management-system/internal/infrastructure/config"
	"github.com/iSparshP/product-management-system/internal/infrastructure/kafka"
	"github.com/iSparshP/product-management-system/internal/infrastructure/s3"
	"go.uber.org/zap"
)

type ImageProcessor struct {
	Consumer    *kafka.Consumer
	ProductRepo repository.ProductRepository
	S3Client    *s3.Client
	Logger      *zap.Logger
}

func NewImageProcessor(consumer *kafka.Consumer, repo repository.ProductRepository, cfg *config.Config, logger *zap.Logger) *ImageProcessor {
	// Initialize S3 Client
	s3Client := s3.NewS3Client(cfg.AWSAccessKey, cfg.AWSSecretKey, cfg.AWSRegion, cfg.AWSS3Bucket, logger)

	return &ImageProcessor{
		Consumer:    consumer,
		ProductRepo: repo,
		S3Client:    s3Client,
		Logger:      logger,
	}
}

func (ip *ImageProcessor) Start(ctx context.Context) error {
	return ip.Consumer.Start(ctx, ip)
}

func (ip *ImageProcessor) ProcessImageTask(task model.ImageProcessingTask) error {
	ip.Logger.Info("Processing image task", zap.String("product_id", task.ProductID))

	var compressedURLs []string

	for _, url := range task.ImageURLs {
		// Download Image
		resp, err := http.Get(url)
		if err != nil {
			ip.Logger.Error("Failed to download image", zap.String("url", url), zap.Error(err))
			continue // Optionally handle retries or send to DLQ
		}

		if resp.StatusCode != http.StatusOK {
			ip.Logger.Error("Non-OK HTTP status", zap.Int("status_code", resp.StatusCode), zap.String("url", url))
			resp.Body.Close()
			continue
		}

		// Save to Temporary File
		tempFile, err := os.CreateTemp("", "image-*.jpg")
		if err != nil {
			ip.Logger.Error("Failed to create temp file", zap.Error(err))
			resp.Body.Close()
			continue
		}

		_, err = io.Copy(tempFile, resp.Body)
		resp.Body.Close()
		if err != nil {
			ip.Logger.Error("Failed to save image to temp file", zap.Error(err))
			tempFile.Close()
			os.Remove(tempFile.Name())
			continue
		}

		tempFile.Close()

		// Compress Image (Implement actual compression logic here)
		// For simplicity, we're skipping compression

		// Upload to S3 with context
		ctx := context.Background() // or pass context from caller
		fileName := filepath.Base(tempFile.Name()) + ".jpg"
		s3URL, err := ip.S3Client.UploadFile(ctx, tempFile.Name(), fileName)
		if err != nil {
			ip.Logger.Error("Failed to upload image to S3", zap.Error(err))
			os.Remove(tempFile.Name())
			continue
		}

		compressedURLs = append(compressedURLs, s3URL)
		os.Remove(tempFile.Name())

		ip.Logger.Info("Successfully processed and uploaded image", zap.String("s3_url", s3URL))
	}

	// Update Product with Compressed Image URLs
	if len(compressedURLs) > 0 {
		if err := ip.ProductRepo.UpdateCompressedImages(context.Background(), task.ProductID, compressedURLs); err != nil {
			ip.Logger.Error("Failed to update product with compressed images", zap.String("product_id", task.ProductID), zap.Error(err))
			return err
		}
	}

	return nil
}
