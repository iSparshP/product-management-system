// internal/imageprocessor/service/image_processor.go

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/disintegration/imaging"
	"github.com/iSparshP/product-management-system/internal/domain/model"
	"github.com/iSparshP/product-management-system/internal/domain/repository"
	"github.com/iSparshP/product-management-system/internal/infrastructure/config"
	"github.com/iSparshP/product-management-system/internal/infrastructure/kafka"
	"github.com/iSparshP/product-management-system/internal/infrastructure/s3"
	"go.uber.org/zap"
)

const (
	maxRetries    = 3
	retryInterval = 5 * time.Second
)

type ProcessError struct {
	OriginalError error
	RetryCount    int
	TaskID        string
}

func (e *ProcessError) Error() string {
	return fmt.Sprintf("processing error after %d retries: %v", e.RetryCount, e.OriginalError)
}

type ImageProcessor struct {
	Consumer    *kafka.Consumer
	ProductRepo repository.ProductRepository
	S3Client    *s3.Client
	Logger      *zap.Logger
	KafkaDLQ    *kafka.Publisher
}

func NewImageProcessor(consumer *kafka.Consumer, repo repository.ProductRepository, cfg *config.Config, logger *zap.Logger) *ImageProcessor {
	// Initialize S3 Client
	s3Client := s3.NewS3Client(cfg.AWSAccessKey, cfg.AWSSecretKey, cfg.AWSRegion, cfg.AWSS3Bucket, logger)
	dlqPublisher, _ := kafka.NewPublisher(cfg.KafkaBrokers, "image_processing_dlq", logger)

	return &ImageProcessor{
		Consumer:    consumer,
		ProductRepo: repo,
		S3Client:    s3Client,
		Logger:      logger,
		KafkaDLQ:    dlqPublisher,
	}
}

func (ip *ImageProcessor) Start(ctx context.Context) error {
	return ip.Consumer.Start(ctx, ip)
}

func (ip *ImageProcessor) ProcessImageTask(task model.ImageProcessingTask) error {
	ip.Logger.Info("Processing image task", zap.String("product_id", task.ProductID))

	var compressedURLs []string
	var processingErrors []error

	for _, url := range task.ImageURLs {
		processedURL, err := ip.processImageWithRetry(url, task.ProductID)
		if err != nil {
			processingErrors = append(processingErrors, err)
			continue
		}
		if processedURL != "" {
			compressedURLs = append(compressedURLs, processedURL)
		}
	}

	// Handle results
	if len(compressedURLs) > 0 {
		if err := ip.updateProductImages(task.ProductID, compressedURLs); err != nil {
			// If update fails, send to DLQ for manual review
			ip.sendToDLQ(task, err, compressedURLs)
			return err
		}
		ip.Logger.Info("Successfully updated product with compressed images",
			zap.String("product_id", task.ProductID),
			zap.Strings("compressed_urls", compressedURLs))
	}

	// If we had any errors but also some successes, log warning
	if len(processingErrors) > 0 && len(compressedURLs) > 0 {
		ip.Logger.Warn("Partial success processing images",
			zap.String("product_id", task.ProductID),
			zap.Int("success_count", len(compressedURLs)),
			zap.Int("error_count", len(processingErrors)))
	}

	// If all images failed, return error
	if len(processingErrors) == len(task.ImageURLs) {
		err := fmt.Errorf("all images failed to process: %v", processingErrors)
		ip.sendToDLQ(task, err, nil)
		return err
	}

	return nil
}

func (ip *ImageProcessor) processImageWithRetry(url, productID string) (string, error) {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		s3URL, err := ip.processImage(url)
		if err == nil {
			return s3URL, nil
		}

		lastErr = err
		ip.Logger.Warn("Image processing attempt failed",
			zap.String("url", url),
			zap.String("product_id", productID),
			zap.Int("attempt", attempt),
			zap.Error(err))

		if !ip.isRetryableError(err) {
			return "", err
		}

		if attempt < maxRetries {
			time.Sleep(retryInterval * time.Duration(attempt))
		}
	}

	return "", &ProcessError{
		OriginalError: lastErr,
		RetryCount:    maxRetries,
		TaskID:        productID,
	}
}

func (ip *ImageProcessor) processImage(url string) (string, error) {
	// Download Image
	resp, err := ip.downloadImage(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	// Decode and process image
	img, err := ip.decodeAndCompressImage(resp.Body)
	if err != nil {
		return "", fmt.Errorf("processing failed: %w", err)
	}

	// Save to temporary file
	tempFile, err := ip.saveToTempFile(img)
	if err != nil {
		return "", fmt.Errorf("save failed: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Upload to S3
	s3URL, err := ip.uploadToS3(tempFile)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	return s3URL, nil
}

func (ip *ImageProcessor) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types that should be retried
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, io.ErrUnexpectedEOF),
		errors.Is(err, syscall.ECONNRESET),
		errors.Is(err, syscall.ETIMEDOUT):
		return true
	}

	// Check for network-related errors
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary() || netErr.Timeout()
	}

	// Add more specific error checks as needed
	return false
}

func (ip *ImageProcessor) sendToDLQ(task model.ImageProcessingTask, err error, partialResults []string) {
	dlqMessage := model.DLQMessage{
		TaskID:         task.ProductID,
		OriginalTask:   task,
		Error:          err.Error(),
		PartialResults: partialResults,
		Timestamp:      time.Now(),
		RetryCount:     maxRetries,
	}

	if err := ip.publishToDLQ(dlqMessage); err != nil {
		ip.Logger.Error("Failed to publish to DLQ",
			zap.String("product_id", task.ProductID),
			zap.Error(err))
	}
}

func (ip *ImageProcessor) publishToDLQ(message model.DLQMessage) error {
	// Implement DLQ publishing logic here
	// This could be a separate Kafka topic, database table, or other storage
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ message: %w", err)
	}

	// Example using Kafka DLQ topic
	if err := ip.KafkaDLQ.Publish(context.Background(), messageBytes); err != nil {
		return fmt.Errorf("failed to publish to DLQ: %w", err)
	}

	return nil
}

func (ip *ImageProcessor) updateProductImages(productID string, compressedURLs []string) error {
	ctx := context.Background()
	return ip.ProductRepo.UpdateCompressedImages(ctx, productID, compressedURLs)
}

func (ip *ImageProcessor) saveToTempFile(img image.Image) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "compressed-*.jpg")
	if err != nil {
		return nil, err
	}

	if err := imaging.Save(img, tempFile.Name(), imaging.JPEGQuality(80)); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, err
	}

	return tempFile, nil
}

func (ip *ImageProcessor) decodeAndCompressImage(r io.Reader) (image.Image, error) {
	img, err := imaging.Decode(r)
	if err != nil {
		return nil, err
	}

	return imaging.Fit(img, 1200, 1200, imaging.Lanczos), nil
}

func (ip *ImageProcessor) uploadToS3(file *os.File) (string, error) {
	ctx := context.Background()
	return ip.S3Client.UploadFile(ctx, file.Name(), file.Name())
}

func (ip *ImageProcessor) downloadImage(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	return client.Get(url)
}
