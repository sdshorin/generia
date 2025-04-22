package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/character-service/internal/models"
	"go.uber.org/zap"
)

type CharacterRepository interface {
	CreateCharacter(ctx context.Context, params models.CreateCharacterParams) (*models.Character, error)
	GetCharacter(ctx context.Context, id string) (*models.Character, error)
	GetUserCharactersInWorld(ctx context.Context, userID, worldID string) ([]*models.Character, error)
}

type characterRepository struct {
	db *sql.DB
}

func NewCharacterRepository(db *sql.DB) CharacterRepository {
	return &characterRepository{
		db: db,
	}
}

func (r *characterRepository) CreateCharacter(ctx context.Context, params models.CreateCharacterParams) (*models.Character, error) {
	query := `
		INSERT INTO world_user_characters (world_id, real_user_id, display_name, avatar_media_id, meta)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, world_id, real_user_id, is_ai, display_name, avatar_media_id, meta, created_at
	`

	var character models.Character
	err := r.db.QueryRowContext(
		ctx,
		query,
		params.WorldID,
		params.RealUserID,
		params.DisplayName,
		params.AvatarMediaID,
		params.Meta,
	).Scan(
		&character.ID,
		&character.WorldID,
		&character.RealUserID,
		&character.IsAI,
		&character.DisplayName,
		&character.AvatarMediaID,
		&character.Meta,
		&character.CreatedAt,
	)

	if err != nil {
		logger.Logger.Error("Failed to create character", zap.Error(err))
		return nil, err
	}

	return &character, nil
}

func (r *characterRepository) GetCharacter(ctx context.Context, id string) (*models.Character, error) {
	query := `
		SELECT id, world_id, real_user_id, is_ai, display_name, avatar_media_id, meta, created_at
		FROM world_user_characters
		WHERE id = $1
	`

	var character models.Character
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&character.ID,
		&character.WorldID,
		&character.RealUserID,
		&character.IsAI,
		&character.DisplayName,
		&character.AvatarMediaID,
		&character.Meta,
		&character.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("character not found")
		}
		logger.Logger.Error("Failed to get character", zap.Error(err))
		return nil, err
	}

	return &character, nil
}

func (r *characterRepository) GetUserCharactersInWorld(ctx context.Context, userID, worldID string) ([]*models.Character, error) {
	query := `
		SELECT id, world_id, real_user_id, is_ai, display_name, avatar_media_id, meta, created_at
		FROM world_user_characters
		WHERE real_user_id = $1 AND world_id = $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, worldID)
	if err != nil {
		logger.Logger.Error("Failed to get user characters in world", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var characters []*models.Character
	for rows.Next() {
		var character models.Character
		err := rows.Scan(
			&character.ID,
			&character.WorldID,
			&character.RealUserID,
			&character.IsAI,
			&character.DisplayName,
			&character.AvatarMediaID,
			&character.Meta,
			&character.CreatedAt,
		)
		if err != nil {
			logger.Logger.Error("Failed to scan character", zap.Error(err))
			return nil, err
		}
		characters = append(characters, &character)
	}

	if err = rows.Err(); err != nil {
		logger.Logger.Error("Error iterating characters", zap.Error(err))
		return nil, err
	}

	return characters, nil
}
