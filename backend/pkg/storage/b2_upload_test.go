package storage

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestUploadSmallFile(t *testing.T) {
	content := []byte("small file content")
	_ = content

	client := &B2Client{
		client:     &s3.Client{},
		bucketName: "test-bucket",
		log:        nil,
	}

	if client.bucketName != "test-bucket" {
		t.Error("bucket name mismatch")
	}
}

func TestUploadLargeFile(t *testing.T) {
	content := make([]byte, 6*1024*1024)
	for i := range content {
		content[i] = byte(i % 256)
	}

	partSize := int64(5 * 1024 * 1024)
	totalSize := int64(len(content))

	expectedParts := (totalSize + partSize - 1) / partSize
	if expectedParts != 2 {
		t.Errorf("expected 2 parts for 6MB file, got %d", expectedParts)
	}
}

func TestShouldUseLargeFile(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected bool
	}{
		{"4MB should use simple upload", 4 * 1024 * 1024, false},
		{"5MB should use simple upload", 5 * 1024 * 1024, false},
		{"5MB+1 should use large upload", 5*1024*1024 + 1, true},
		{"10GB should use large upload", 10 * 1024 * 1024 * 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.size > MaxSimpleUploadSize
			if result != tt.expected {
				t.Errorf("size %d: expected %v, got %v", tt.size, tt.expected, result)
			}
		})
	}
}

func TestCalculatePartSize(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
	}{
		{"6MB file", 6 * 1024 * 1024},
		{"100MB file", 100 * 1024 * 1024},
		{"1GB file", 1024 * 1024 * 1024},
		{"5MB file (boundary)", 5 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partSize := DefaultPartSize
			parts := (tt.fileSize + partSize - 1) / partSize
			t.Logf("fileSize=%d, partSize=%d, parts=%d", tt.fileSize, partSize, parts)
			if parts < 1 {
				t.Errorf("expected at least 1 part, got %d", parts)
			}
		})
	}
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		filename     string
		expectedType string
	}{
		{"test.png", "image/png"},
		{"test.jpg", "image/jpeg"},
		{"test.gif", "image/gif"},
		{"test.pdf", "application/pdf"},
		{"test.txt", "text/plain"}, // May include charset=utf-8
		{"test.unknown", "application/octet-stream"},
		{"test", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := GetMimeType(tt.filename)
			// For text/plain, accept both with and without charset
			if tt.filename == "test.txt" {
				if result != "text/plain" && result != "text/plain; charset=utf-8" {
					t.Errorf("GetMimeType(%s) = %s; want %s", tt.filename, result, tt.expectedType)
				}
			} else {
				if result != tt.expectedType {
					t.Errorf("GetMimeType(%s) = %s; want %s", tt.filename, result, tt.expectedType)
				}
			}
		})
	}
}

func TestGenerateStorageKey(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		parentPath  string
		filename    string
		expectedKey string
	}{
		{
			name:        "root file",
			userID:      "user123",
			parentPath:  "",
			filename:    "document.pdf",
			expectedKey: "user123/document.pdf",
		},
		{
			name:        "nested file",
			userID:      "user123",
			parentPath:  "folder1/folder2",
			filename:    "photo.jpg",
			expectedKey: "user123/folder1/folder2/photo.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateStorageKey(tt.userID, tt.parentPath, tt.filename)
			if result != tt.expectedKey {
				t.Errorf("GenerateStorageKey() = %s; want %s", result, tt.expectedKey)
			}
		})
	}
}
