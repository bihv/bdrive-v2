package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"

	"github.com/biho/onedrive/internal/config"
)

const (
	// MaxSimpleUploadSize is the maximum size for simple upload (5MB)
	// B2 allows uploads up to 5GB in a single request, but for reliability
	// we use 5MB as the threshold as mentioned in B2 docs
	MaxSimpleUploadSize int64 = 5 * 1024 * 1024

	// DefaultPartSize is the default part size for multipart uploads (5MB)
	// B2 recommends using the recommendedPartSize from b2_authorize_account
	// Using 5MB as it's the minimum allowed by B2
	DefaultPartSize int64 = 5 * 1024 * 1024

	// MaxPartSize is the maximum part size (5GB as per B2 docs)
	MaxPartSize int64 = 5 * 1024 * 1024 * 1024
)

// UploadResult contains the result of an upload operation
type UploadResult struct {
	StorageKey string
	VersionID  *string
}

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

// UploadFile uploads a file to B2 storage.
// It automatically chooses between simple upload (<=5MB) and multipart upload (>5MB).
func (b *B2Client) UploadFile(ctx context.Context, key string, content []byte, contentType string) (*UploadResult, error) {
	size := int64(len(content))

	if size <= MaxSimpleUploadSize {
		return b.uploadSimple(ctx, key, content, contentType)
	}

	return b.uploadLarge(ctx, key, bytes.NewReader(content), size, contentType)
}

// UploadFileFromReader uploads a file from an io.Reader to B2 storage.
// It automatically chooses between simple upload (<=5MB) and multipart upload (>5MB).
func (b *B2Client) UploadFileFromReader(ctx context.Context, key string, reader io.ReaderAt, size int64, contentType string) (*UploadResult, error) {
	if size <= MaxSimpleUploadSize {
		buf := make([]byte, size)
		_, err := reader.ReadAt(buf, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to read content: %w", err)
		}
		return b.uploadSimple(ctx, key, buf, contentType)
	}

	return b.uploadLarge(ctx, key, io.NewSectionReader(reader, 0, size), size, contentType)
}

// uploadSimple uploads a file using simple PutObject (for files <= 5MB)
func (b *B2Client) uploadSimple(ctx context.Context, key string, content []byte, contentType string) (*UploadResult, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(b.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	}

	result, err := b.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadResult{
		StorageKey: key,
		VersionID:  result.VersionId,
	}, nil
}

// uploadLarge uploads a file using multipart upload (for files > 5MB)
func (b *B2Client) uploadLarge(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*UploadResult, error) {
	// Step 1: Create multipart upload
	createInput := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(b.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	createResult, err := b.client.CreateMultipartUpload(ctx, createInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart upload: %w", err)
	}

	uploadID := *createResult.UploadId

	// Calculate part size - use default or adjust for large files
	partSize := DefaultPartSize
	if size > MaxPartSize*100 {
		// For very large files, increase part size to reduce number of parts
		partSize = MaxPartSize
	}

	// Step 2: Upload parts
	completedParts := make([]types.CompletedPart, 0)
	partNumber := 1
	offset := int64(0)

	for offset < size {
		// Calculate remaining bytes
		remaining := size - offset
		currentPartSize := partSize
		if remaining < partSize {
			currentPartSize = remaining
		}

		// Read part data
		partData := make([]byte, currentPartSize)
		n, err := reader.Read(partData)
		if err != nil && err != io.EOF {
			// Abort on error
			b.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(b.bucketName),
				Key:      aws.String(key),
				UploadId: aws.String(uploadID),
			})
			return nil, fmt.Errorf("failed to read part data: %w", err)
		}

		// Upload part
		uploadPartInput := &s3.UploadPartInput{
			Bucket:     aws.String(b.bucketName),
			Key:        aws.String(key),
			UploadId:   aws.String(uploadID),
			PartNumber: aws.Int32(int32(partNumber)),
			Body:       bytes.NewReader(partData[:n]),
		}

		partResult, err := b.client.UploadPart(ctx, uploadPartInput)
		if err != nil {
			b.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(b.bucketName),
				Key:      aws.String(key),
				UploadId: aws.String(uploadID),
			})
			return nil, fmt.Errorf("failed to upload part %d: %w", partNumber, err)
		}

		completedParts = append(completedParts, types.CompletedPart{
			ETag:       partResult.ETag,
			PartNumber: aws.Int32(int32(partNumber)),
		})

		offset += currentPartSize
		partNumber++
	}

	// Step 3: Complete multipart upload
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(b.bucketName),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}

	_, err = b.client.CompleteMultipartUpload(ctx, completeInput)
	if err != nil {
		return nil, fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	return &UploadResult{
		StorageKey: key,
	}, nil
}

// DeleteFile deletes a file from B2 storage
func (b *B2Client) DeleteFile(ctx context.Context, key string) error {
	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetFileURL returns a pre-signed URL for downloading a file
func (b *B2Client) GetFileURL(ctx context.Context, key string, expirySeconds int64) (string, error) {
	presignClient := s3.NewPresignClient(b.client)

	result, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirySeconds) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return result.URL, nil
}

// GetUploadURL returns a pre-signed URL for simple upload (<=5MB)
func (b *B2Client) GetUploadURL(ctx context.Context, key string, contentType string, expirySeconds int64) (string, error) {
	presignClient := s3.NewPresignClient(b.client)

	result, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(b.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirySeconds) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return result.URL, nil
}

// CreateMultipartUploadResult contains the result of creating a multipart upload
type CreateMultipartUploadResult struct {
	UploadID string
	Key      string
}

// CreateMultipartUpload creates a multipart upload and returns the upload ID
func (b *B2Client) CreateMultipartUpload(ctx context.Context, key string, contentType string) (*CreateMultipartUploadResult, error) {
	result, err := b.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(b.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart upload: %w", err)
	}

	return &CreateMultipartUploadResult{
		UploadID: *result.UploadId,
		Key:      key,
	}, nil
}

// GetUploadPartURL returns a pre-signed URL for uploading a part
func (b *B2Client) GetUploadPartURL(ctx context.Context, key string, uploadID string, partNumber int, expirySeconds int64) (string, error) {
	presignClient := s3.NewPresignClient(b.client)

	result, err := presignClient.PresignUploadPart(ctx, &s3.UploadPartInput{
		Bucket:     aws.String(b.bucketName),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(int32(partNumber)),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirySeconds) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload part URL: %w", err)
	}

	return result.URL, nil
}

// CompleteMultipartUpload completes a multipart upload
func (b *B2Client) CompleteMultipartUpload(ctx context.Context, key string, uploadID string, parts []types.CompletedPart) error {
	_, err := b.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(b.bucketName),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	return nil
}

// ListParts lists all uploaded parts for a multipart upload
func (b *B2Client) ListParts(ctx context.Context, key string, uploadID string) ([]types.CompletedPart, error) {
	result, err := b.client.ListParts(ctx, &s3.ListPartsInput{
		Bucket:   aws.String(b.bucketName),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list parts: %w", err)
	}

	parts := make([]types.CompletedPart, len(result.Parts))
	for i, part := range result.Parts {
		parts[i] = types.CompletedPart{
			ETag:       part.ETag,
			PartNumber: part.PartNumber,
		}
	}

	return parts, nil
}

// AbortMultipartUpload aborts a multipart upload
func (b *B2Client) AbortMultipartUpload(ctx context.Context, key string, uploadID string) error {
	_, err := b.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(b.bucketName),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		return fmt.Errorf("failed to abort multipart upload: %w", err)
	}
	return nil
}

// GenerateStorageKey generates a unique storage key for a file
func GenerateStorageKey(userID string, parentPath string, filename string) string {
	// Generate key: {userID}/{parentPath}/{filename}
	return filepath.Join(userID, parentPath, filename)
}

// GetMimeType determines the MIME type from the file extension
func GetMimeType(filename string) string {
	ext := filepath.Ext(filename)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return contentType
}
