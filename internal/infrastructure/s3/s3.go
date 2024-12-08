// internal/infrastructure/s3/s3.go

package s3

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type Client struct {
	s3Client *s3.Client
	bucket   string
	logger   *zap.Logger
}

func NewS3Client(accessKey, secretKey, region, bucket, endpoint string, logger *zap.Logger) *Client {
	logger.Info("Initializing S3 client",
		zap.String("region", region),
		zap.String("bucket", bucket),
		zap.String("endpoint", endpoint),
		zap.Bool("hasAccessKey", accessKey != ""),
		zap.Bool("hasSecretKey", secretKey != ""))

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		logger.Fatal("Failed to create AWS config", zap.Error(err))
	}

	// Create S3 client with custom endpoint if provided
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	// Test the credentials
	_, err = s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		logger.Error("Failed to list buckets", zap.Error(err))
	} else {
		logger.Info("Successfully authenticated with AWS")
	}

	return &Client{
		s3Client: s3Client,
		bucket:   bucket,
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

	// Upload the file
	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String(getContentType(fileName)),
	})
	if err != nil {
		c.logger.Error("Failed to upload file to S3",
			zap.Error(err),
			zap.String("bucket", c.bucket),
			zap.String("key", key))
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate the S3 URL
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.bucket, "ap-south-1", key)

	c.logger.Info("Successfully uploaded file to S3",
		zap.String("bucket", c.bucket),
		zap.String("key", key),
		zap.String("url", url))

	return url, nil
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
