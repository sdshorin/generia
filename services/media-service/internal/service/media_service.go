package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/sdshorin/generia/services/media-service/internal/models"
	"github.com/sdshorin/generia/services/media-service/internal/repository"
	"go.uber.org/zap"
)

// MediaService provides business logic for media operations
type MediaService struct {
	repo        repository.MediaRepository
	minioClient *minio.Client
	bucket      string
	logger      *zap.Logger
}

// NewMediaService creates a new MediaService
func NewMediaService(repo repository.MediaRepository, minioClient *minio.Client, bucket string, logger *zap.Logger) *MediaService {
	return &MediaService{
		repo:        repo,
		minioClient: minioClient,
		bucket:      bucket,
		logger:      logger,
	}
}

// CreateMedia creates a new media record
func (s *MediaService) CreateMedia(ctx context.Context, userID, filename, contentType string, size int64, data []byte) (*models.Media, error) {
	// Generate a unique ID
	id, err := GenerateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	// Generate object name for MinIO
	objectName := fmt.Sprintf("%s/%s%s", userID, id, filepath.Ext(filename))

	// Upload to MinIO
	reader := bytes.NewReader(data)
	_, err = s.minioClient.PutObject(ctx, s.bucket, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	// Create media record
	media := &models.Media{
		ID:          id,
		UserID:      userID,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		BucketName:  s.bucket,
		ObjectName:  objectName,
	}

	// Store in database
	err = s.repo.CreateMedia(ctx, media)
	if err != nil {
		return nil, fmt.Errorf("failed to store media in database: %w", err)
	}

	return media, nil
}

// GeneratePresignedPutURL generates a presigned URL for client-side uploading
func (s *MediaService) GeneratePresignedPutURL(ctx context.Context, userID, filename, contentType string, size int64) (*models.Media, string, time.Time, error) {
	// Generate a unique ID
	id, err := GenerateID()
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("failed to generate ID: %w", err)
	}

	// Generate object name for MinIO
	objectName := fmt.Sprintf("%s/%s%s", userID, id, filepath.Ext(filename))

	// Generate presigned PUT URL
	expiry := time.Minute * 10 // 10 minutes expiry for upload
	presignedURL, err := s.minioClient.PresignedPutObject(ctx, s.bucket, objectName, expiry)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Create media record (status pending until confirmed after upload)
	media := &models.Media{
		ID:          id,
		UserID:      userID,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		BucketName:  s.bucket,
		ObjectName:  objectName,
	}

	// Store in database
	err = s.repo.CreateMedia(ctx, media)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("failed to store media in database: %w", err)
	}

	expiresAt := time.Now().Add(expiry)
	return media, presignedURL.String(), expiresAt, nil
}

// ConfirmMediaUpload confirms that a media file has been uploaded via presigned URL
func (s *MediaService) ConfirmMediaUpload(ctx context.Context, mediaID string) error {
	// Get media from database
	media, err := s.repo.GetMediaByID(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("failed to get media from database: %w", err)
	}

	// Check if object exists in MinIO
	_, err = s.minioClient.StatObject(ctx, media.BucketName, media.ObjectName, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to verify media in storage: %w", err)
	}

	// Generate variants if needed (thumbnails, etc.)
	// This could be triggered as an async process via Kafka

	return nil
}

// GetMedia retrieves a media by its ID
func (s *MediaService) GetMedia(ctx context.Context, id string) (*models.Media, []*models.MediaVariant, error) {
	// Get media from database
	media, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get media from database: %w", err)
	}

	// Get variants
	variants, err := s.repo.GetMediaVariants(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get media variants: %w", err)
	}

	return media, variants, nil
}

// GetPresignedURL generates a presigned URL for a media object
func (s *MediaService) GetPresignedURL(ctx context.Context, media *models.Media, variant string, expiresIn time.Duration) (string, time.Time, error) {
	// Determine the object name based on variant
	var objectName string
	if variant == "original" {
		objectName = media.ObjectName
	} else {
		objectName = fmt.Sprintf("%s/%s_%s%s", media.UserID, media.ID, variant, filepath.Ext(media.Filename))
	}

	// If no variants exist yet but original is requested, use the object name
	if variant != "original" {
		// Check if the variant exists
		_, err := s.minioClient.StatObject(ctx, media.BucketName, objectName, minio.StatObjectOptions{})
		if err != nil {
			// If variant doesn't exist, use the original
			objectName = media.ObjectName
		}
	}

	// Generate presigned URL
	reqParams := make(url.Values)
	if media.ContentType != "" {
		reqParams.Set("response-content-type", media.ContentType)
	}

	url, err := s.minioClient.PresignedGetObject(ctx, media.BucketName, objectName, expiresIn, reqParams)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	expiresAt := time.Now().Add(expiresIn)
	return url.String(), expiresAt, nil
}

// GenerateVariants creates different size variants of an image
func (s *MediaService) GenerateVariants(ctx context.Context, mediaID string, variantsToCreate []string) ([]*models.MediaVariant, error) {
	// In a real implementation, this would:
	// 1. Get the original media from MinIO
	// 2. Use an image processing library to create variants
	// 3. Upload variants to MinIO
	// 4. Store variant information in the database

	// For now, just return a placeholder implementation
	s.logger.Info("GenerateVariants called",
		zap.String("media_id", mediaID),
		zap.Strings("variants", variantsToCreate))

	// This method would typically be called asynchronously after an upload

	return []*models.MediaVariant{}, nil
}

// GenerateID generates a unique ID for media
// This function is exported so it can be used by other packages
func GenerateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
