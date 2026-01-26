package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Client wraps the S3 client for Cloudflare R2
type R2Client struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

// NewR2Client creates a new R2 storage client
func NewR2Client() (*R2Client, error) {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")
	publicURL := os.Getenv("R2_PUBLIC_URL")

	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing R2 configuration environment variables")
	}

	// R2 endpoint format
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	// Create S3 client configured for R2
	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
	})

	return &R2Client{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}, nil
}

// GeneratePresignedURL creates a presigned URL for uploading
func (r *R2Client) GeneratePresignedURL(ctx context.Context, key string, contentType string) (string, error) {
	presignClient := s3.NewPresignClient(r.client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return request.URL, nil
}

// GetPublicURL returns the public URL for an uploaded file
func (r *R2Client) GetPublicURL(key string) string {
	if r.publicURL != "" {
		return fmt.Sprintf("%s/%s", r.publicURL, key)
	}
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s",
		os.Getenv("R2_ACCOUNT_ID"), r.bucketName, key)
}

// DeleteObject deletes a file from R2
func (r *R2Client) DeleteObject(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	return err
}

// GetObject retrieves an object from R2
func (r *R2Client) GetObject(ctx context.Context, key string) ([]byte, error) {
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return data, nil
}

// PutObject uploads an object to R2
func (r *R2Client) PutObject(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	return nil
}

// MoveObject moves an object from one key to another (copy + delete)
func (r *R2Client) MoveObject(ctx context.Context, srcKey, dstKey string) error {
	// Copy to new location
	copySource := fmt.Sprintf("%s/%s", r.bucketName, srcKey)
	_, err := r.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(r.bucketName),
		Key:        aws.String(dstKey),
		CopySource: aws.String(copySource),
	})
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	// Delete original
	if err := r.DeleteObject(ctx, srcKey); err != nil {
		return fmt.Errorf("failed to delete original after copy: %w", err)
	}

	return nil
}

// GeneratePresignedURLWithMaxSize creates a presigned URL with content-length constraints
func (r *R2Client) GeneratePresignedURLWithMaxSize(ctx context.Context, key string, contentType string, maxSizeBytes int64) (string, error) {
	presignClient := s3.NewPresignClient(r.client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucketName),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(maxSizeBytes),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return request.URL, nil
}
