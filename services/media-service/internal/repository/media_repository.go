package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"instagram-clone/services/media-service/internal/models"
	"time"
)

// MediaRepository defines the interface for media data access
type MediaRepository interface {
	CreateMedia(ctx context.Context, media *models.Media) error
	GetMediaByID(ctx context.Context, id string) (*models.Media, error)
	GetMediaVariants(ctx context.Context, mediaID string) ([]*models.MediaVariant, error)
	CreateMediaVariant(ctx context.Context, variant *models.MediaVariant) error
}

// PostgresMediaRepository implements MediaRepository interface
type PostgresMediaRepository struct {
	db          *sqlx.DB
	minioClient *minio.Client
}

// NewPostgresMediaRepository creates a new PostgresMediaRepository
func NewPostgresMediaRepository(db *sqlx.DB, minioClient *minio.Client) *PostgresMediaRepository {
	return &PostgresMediaRepository{
		db:          db,
		minioClient: minioClient,
	}
}

// CreateMedia stores a new media record in the database
func (r *PostgresMediaRepository) CreateMedia(ctx context.Context, media *models.Media) error {
	// Set timestamps
	now := time.Now()
	media.CreatedAt = now
	media.UpdatedAt = now

	// Insert media record
	query := `
		INSERT INTO media (id, user_id, filename, content_type, size, bucket_name, object_name, created_at, updated_at)
		VALUES (:id, :user_id, :filename, :content_type, :size, :bucket_name, :object_name, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, media)
	return err
}

// GetMediaByID retrieves a media record by its ID
func (r *PostgresMediaRepository) GetMediaByID(ctx context.Context, id string) (*models.Media, error) {
	var media models.Media
	query := `
		SELECT id, user_id, filename, content_type, size, bucket_name, object_name, created_at, updated_at
		FROM media
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &media, query, id)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// GetMediaVariants retrieves all variants for a given media
func (r *PostgresMediaRepository) GetMediaVariants(ctx context.Context, mediaID string) ([]*models.MediaVariant, error) {
	variants := []*models.MediaVariant{}
	query := `
		SELECT id, media_id, name, url, width, height, created_at
		FROM media_variants
		WHERE media_id = $1
	`
	err := r.db.SelectContext(ctx, &variants, query, mediaID)
	if err != nil {
		return nil, err
	}
	return variants, nil
}

// CreateMediaVariant stores a new media variant record
func (r *PostgresMediaRepository) CreateMediaVariant(ctx context.Context, variant *models.MediaVariant) error {
	variant.CreatedAt = time.Now()
	query := `
		INSERT INTO media_variants (id, media_id, name, url, width, height, created_at)
		VALUES (:id, :media_id, :name, :url, :width, :height, :created_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, variant)
	return err
}