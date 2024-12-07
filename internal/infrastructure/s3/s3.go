// internal/infrastructure/s3/s3.go

package s3

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
)

type Client struct {
	Uploader *s3manager.Uploader
	Bucket   string
	logger   *zap.Logger
}

func NewS3Client(accessKey, secretKey, region, bucket string, logger *zap.Logger) *Client {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		logger.Fatal("Failed to create AWS session", zap.Error(err))
	}

	uploader := s3manager.NewUploader(sess)

	return &Client{
		Uploader: uploader,
		Bucket:   bucket,
		logger:   logger,
	}
}

func (c *Client) UploadFile(ctx context.Context, filePath, fileName string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		c.logger.Error("Failed to open file for S3 upload",
			zap.Error(err),
			zap.String("file_path", filePath))
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Generate a unique key for the file
	key := fmt.Sprintf("products/%s", filepath.Base(fileName))

	result, err := c.Uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(c.Bucket),
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String(getContentType(fileName)),
	})
	if err != nil {
		c.logger.Error("Failed to upload file to S3",
			zap.Error(err),
			zap.String("bucket", c.Bucket),
			zap.String("key", key))
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	c.logger.Info("Successfully uploaded file to S3",
		zap.String("bucket", c.Bucket),
		zap.String("key", key),
		zap.String("location", result.Location))

	return result.Location, nil
}

// getContentType determines the content type based on file extension
func getContentType(fileName string) string {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
