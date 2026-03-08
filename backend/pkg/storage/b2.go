package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"github.com/biho/onedrive/internal/config"
)

// B2Client wraps the S3-compatible client for Backblaze B2.
type B2Client struct {
	client     *s3.Client
	bucketName string
	log        *zap.Logger
}

// NewB2Client creates a new Backblaze B2 storage client using S3-compatible API.
func NewB2Client(cfg *config.B2Config, log *zap.Logger) (*B2Client, error) {
	if cfg.KeyID == "" || cfg.AppKey == "" || cfg.BucketName == "" {
		return nil, fmt.Errorf("B2 configuration incomplete: KeyID, AppKey, and BucketName are required")
	}

	// Build S3 client pointing to B2 endpoint
	s3Client := s3.New(s3.Options{
		Region:       cfg.Region,
		BaseEndpoint: aws.String(cfg.Endpoint),
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.KeyID,
			cfg.AppKey,
			"", // session token not used for B2
		),
	})

	return &B2Client{
		client:     s3Client,
		bucketName: cfg.BucketName,
		log:        log,
	}, nil
}

// HealthCheck verifies the connection to B2 by checking if the bucket exists.
func (b *B2Client) HealthCheck(ctx context.Context) error {
	_, err := b.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(b.bucketName),
	})
	if err != nil {
		return fmt.Errorf("B2 health check failed: %w", err)
	}
	return nil
}

// GetClient returns the underlying S3 client for advanced usage.
func (b *B2Client) GetClient() *s3.Client {
	return b.client
}

// GetBucketName returns the configured bucket name.
func (b *B2Client) GetBucketName() string {
	return b.bucketName
}
