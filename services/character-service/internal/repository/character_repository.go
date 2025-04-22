package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/generia/pkg/logger"
	"github.com/generia/services/character-service/internal/models"
)

type CharacterRepository interface {
	CreateCharacter(ctx context.Context, params models.CreateCharacterParams) (*models.Character, error)
	GetCharacter(ctx context.Context, id string) (*models.Character, error)
	GetUserCharactersInWorld(ctx context.Context, userID, worldID string) ([]*models.Character, error)
}

type characterRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewCharacterRepository(db *sql.DB, logger logger.Logger) CharacterRepository {
	return &characterRepository{
		db:     db,
		logger: logger,
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
		r.logger.Error("Failed to create character", err)
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
		r.logger.Error("Failed to get character", err)
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
		r.logger.Error("Failed to get user characters in world", err)
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
			r.logger.Error("Failed to scan character", err)
			return nil, err
		}
		characters = append(characters, &character)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error iterating characters", err)
		return nil, err
	}

	return characters, nil
}