package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/post-service/internal/models"
	"go.uber.org/zap"
)

// PostRepository handles database operations for posts
type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id string) (*models.Post, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Post, int, error)
	GetByCharacterID(ctx context.Context, characterID string, limit, offset int) ([]*models.Post, int, error)
	GetGlobalFeed(ctx context.Context, limit int, cursor string, worldID string) ([]*models.Post, string, error)
	GetByIDs(ctx context.Context, ids []string) ([]*models.Post, error)
}

type postRepository struct {
	db *sqlx.DB
}

// NewPostRepository creates a new PostRepository
func NewPostRepository(db *sqlx.DB) PostRepository {
	return &postRepository{
		db: db,
	}
}

// Create inserts a new post into the database
func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
	query := `
		INSERT INTO posts (id, character_id, is_ai, world_id, caption, media_id, created_at, updated_at)
		VALUES (:id, :character_id, :is_ai, :world_id, :caption, :media_id, :created_at, :updated_at)
		RETURNING id
	`

	// Set timestamps
	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	// Generate UUID if not provided
	if post.ID == "" {
		post.ID = uuid.New().String()
	}

	// Use named parameters
	var id string
	rows, err := r.db.NamedQueryContext(ctx, query, post)
	if err != nil {
		logger.Logger.Error("Failed to create post", zap.Error(err))
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			logger.Logger.Error("Failed to scan post ID", zap.Error(err))
			return err
		}
		post.ID = id
	}

	return nil
}

// GetByID retrieves a post by ID
func (r *postRepository) GetByID(ctx context.Context, id string) (*models.Post, error) {
	query := `
		SELECT id, character_id, is_ai, world_id, caption, media_id, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	var post models.Post
	err := r.db.GetContext(ctx, &post, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Logger.Error("Failed to get post by ID", zap.Error(err), zap.String("id", id))
		return nil, err
	}

	return &post, nil
}

// GetByUserID retrieves posts by user ID with pagination
// This method needs to join with the character service
func (r *postRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Post, int, error) {
	query := `
		SELECT p.id, p.character_id, p.is_ai, p.world_id, p.caption, p.media_id, p.created_at, p.updated_at
		FROM posts p
		JOIN world_user_characters wuc ON p.character_id = wuc.id
		WHERE wuc.real_user_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*)
		FROM posts p
		JOIN world_user_characters wuc ON p.character_id = wuc.id
		WHERE wuc.real_user_id = $1
	`

	posts := []*models.Post{}
	err := r.db.SelectContext(ctx, &posts, query, userID, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get posts by user ID",
			zap.Error(err),
			zap.String("user_id", userID))
		return nil, 0, err
	}

	var total int
	err = r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		logger.Logger.Error("Failed to count posts by user ID",
			zap.Error(err),
			zap.String("user_id", userID))
		return nil, 0, err
	}

	return posts, total, nil
}

// GetByCharacterID retrieves posts by character ID with pagination
func (r *postRepository) GetByCharacterID(ctx context.Context, characterID string, limit, offset int) ([]*models.Post, int, error) {
	query := `
		SELECT id, character_id, is_ai, world_id, caption, media_id, created_at, updated_at
		FROM posts
		WHERE character_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*)
		FROM posts
		WHERE character_id = $1
	`

	posts := []*models.Post{}
	err := r.db.SelectContext(ctx, &posts, query, characterID, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get posts by character ID",
			zap.Error(err),
			zap.String("character_id", characterID))
		return nil, 0, err
	}

	var total int
	err = r.db.GetContext(ctx, &total, countQuery, characterID)
	if err != nil {
		logger.Logger.Error("Failed to count posts by character ID",
			zap.Error(err),
			zap.String("character_id", characterID))
		return nil, 0, err
	}

	return posts, total, nil
}

// GetGlobalFeed retrieves posts for the global feed with cursor-based pagination
func (r *postRepository) GetGlobalFeed(ctx context.Context, limit int, cursor string, worldID string) ([]*models.Post, string, error) {
	var query string
	var args []interface{}

	if worldID == "" {
		// Если мир не указан, возвращаем пустой результат
		return []*models.Post{}, "", nil
	}

	if cursor == "" {
		// First page
		query = `
			SELECT id, character_id, is_ai, world_id, caption, media_id, created_at, updated_at
			FROM posts
			WHERE world_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{worldID, limit}
	} else {
		// Subsequent pages
		query = `
			SELECT id, character_id, is_ai, world_id, caption, media_id, created_at, updated_at
			FROM posts
			WHERE world_id = $1
			AND created_at < (
				SELECT created_at
				FROM posts
				WHERE id = $2
			)
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{worldID, cursor, limit}
	}

	posts := []*models.Post{}
	err := r.db.SelectContext(ctx, &posts, query, args...)
	if err != nil {
		logger.Logger.Error("Failed to get global feed",
			zap.Error(err),
			zap.String("cursor", cursor),
			zap.Int("limit", limit),
			zap.String("world_id", worldID))
		return nil, "", err
	}

	var nextCursor string
	if len(posts) > 0 {
		nextCursor = posts[len(posts)-1].ID
	}

	return posts, nextCursor, nil
}

// GetByIDs retrieves posts by IDs
func (r *postRepository) GetByIDs(ctx context.Context, ids []string) ([]*models.Post, error) {
	if len(ids) == 0 {
		return []*models.Post{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT id, character_id, is_ai, world_id, caption, media_id, created_at, updated_at
		FROM posts
		WHERE id IN (?)
		ORDER BY created_at DESC
	`, ids)
	if err != nil {
		logger.Logger.Error("Failed to build query for GetByIDs", zap.Error(err))
		return nil, err
	}

	query = r.db.Rebind(query)
	posts := []*models.Post{}
	err = r.db.SelectContext(ctx, &posts, query, args...)
	if err != nil {
		logger.Logger.Error("Failed to get posts by IDs", zap.Error(err))
		return nil, err
	}

	return posts, nil
}