package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"instagram-clone/services/media-service/internal/models"
	"instagram-clone/services/media-service/internal/repository"
	"path/filepath"
	"time"
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
	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	// Generate object name for MinIO
	objectName := fmt.Sprintf("%s/%s%s", userID, id, filepath.Ext(filename))

	// Upload to MinIO
	_, err = s.minioClient.PutObject(ctx, s.bucket, objectName, data, int64(len(data)), minio.PutObjectOptions{
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

	// Generate presigned URL
	url, err := s.minioClient.PresignedGetObject(ctx, media.BucketName, objectName, expiresIn, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	expiresAt := time.Now().Add(expiresIn)
	return url.String(), expiresAt, nil
}

// generateID generates a unique ID for media
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}