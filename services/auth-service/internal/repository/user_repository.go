package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/auth-service/internal/models"
	"go.uber.org/zap"
)

// UserRepository handles database operations for users
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create inserts a new user into the database
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	var id string
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)

	if err != nil {
		logger.Logger.Error("Failed to create user", zap.Error(err))
		return err
	}

	user.ID = id
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Logger.Error("Failed to get user by ID", zap.Error(err), zap.String("id", id))
		return nil, err
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, err
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Logger.Error("Failed to get user by username", zap.Error(err), zap.String("username", username))
		return nil, err
	}

	return &user, nil
}

// SaveRefreshToken saves a refresh token to the database
func (r *userRepository) SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	// First, clean up any existing tokens for this user that might be expired
	// This helps prevent token accumulation in the database
	cleanupQuery := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1 AND expires_at < $2
	`
	
	_, err := r.db.ExecContext(
		ctx,
		cleanupQuery,
		token.UserID,
		time.Now(),
	)
	
	if err != nil {
		logger.Logger.Warn("Failed to clean up expired refresh tokens", zap.Error(err))
		// Continue even if cleanup fails
	}

	// Insert the new token with ON CONFLICT handling for unique token hash
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4)
		ON CONFLICT (token_hash) DO NOTHING
		RETURNING id
	`

	now := time.Now()
	token.CreatedAt = now

	var id string
	err = r.db.QueryRowContext(
		ctx,
		query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	).Scan(&id)

	if err != nil {
		// Check if this is a no-rows-returned error, which can happen with the ON CONFLICT DO NOTHING
		if err == sql.ErrNoRows {
			logger.Logger.Warn("Token already exists, no update performed", 
				zap.String("token_hash", token.TokenHash))
			return nil
		}
		
		logger.Logger.Error("Failed to save refresh token", zap.Error(err))
		return err
	}

	token.ID = id
	return nil
}

// GetRefreshToken retrieves a refresh token by token hash
func (r *userRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token models.RefreshToken
	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Logger.Error("Failed to get refresh token", zap.Error(err))
		return nil, err
	}

	return &token, nil
}

// DeleteRefreshToken deletes a refresh token by token hash
func (r *userRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1
	`

	_, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		logger.Logger.Error("Failed to delete refresh token", zap.Error(err))
		return err
	}

	return nil
}