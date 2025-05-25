package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/world-service/internal/models"
	"go.uber.org/zap"
)

// WorldRepository is the interface for world data storage
type WorldRepository interface {
	// World operations
	Create(ctx context.Context, world *models.World) error
	GetByID(ctx context.Context, id string) (*models.World, error)
	GetAll(ctx context.Context, limit, offset int, status string) ([]*models.World, int, error)
	GetByUser(ctx context.Context, userID string, limit, offset int, status string) ([]*models.World, int, error)
	UpdateStatus(ctx context.Context, id, status string) error
	Update(ctx context.Context, world *models.World) error

	// User world operations
	AddUserToWorld(ctx context.Context, userID, worldID string) error
	RemoveUserFromWorld(ctx context.Context, userID, worldID string) error
	GetUserWorlds(ctx context.Context, userID string) ([]*models.UserWorld, error)
	CheckUserWorld(ctx context.Context, userID, worldID string) (bool, error)

	// Generation tasks

	// UpdateGenerationTask(ctx context.Context, taskID, status, result string) error
	GetWorldStats(ctx context.Context, worldID string) (int, int, error) // usersCount, postsCount, error
}

// PostgresWorldRepository is the PostgreSQL implementation of WorldRepository
type PostgresWorldRepository struct {
	db *sqlx.DB
}

// NewWorldRepository creates a new WorldRepository
func NewWorldRepository(db *sqlx.DB) WorldRepository {
	return &PostgresWorldRepository{
		db: db,
	}
}

// Create inserts a new world into the database
func (r *PostgresWorldRepository) Create(ctx context.Context, world *models.World) error {
	query := `
		INSERT INTO worlds (name, description, prompt, creator_id, generation_status, status, image_uuid, icon_uuid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	row := r.db.QueryRowContext(
		ctx,
		query,
		world.Name,
		world.Description,
		world.Prompt,
		world.CreatorID,
		"",
		// world.GenerationStatus,
		world.Status,
		world.ImageUUID, // Use sql.NullString which will be NULL when not set
		world.IconUUID,  // Use sql.NullString which will be NULL when not set
	)

	err := row.Scan(&world.ID, &world.CreatedAt, &world.UpdatedAt)
	if err != nil {
		logger.Logger.Error("Failed to create world", zap.Error(err))
		return err
	}

	return nil
}

// GetByID retrieves a world by its ID
func (r *PostgresWorldRepository) GetByID(ctx context.Context, id string) (*models.World, error) {
	query := `
		SELECT id, name, description, prompt, creator_id, generation_status, status, image_uuid, icon_uuid, created_at, updated_at
		FROM worlds
		WHERE id = $1
	`

	var world models.World
	err := r.db.GetContext(ctx, &world, query, id)
	if err != nil {
		logger.Logger.Error("Failed to get world by ID", zap.Error(err), zap.String("id", id))
		return nil, err
	}

	return &world, nil
}

// GetAll retrieves all worlds with pagination
func (r *PostgresWorldRepository) GetAll(ctx context.Context, limit, offset int, status string) ([]*models.World, int, error) {
	var query string
	var args []interface{}

	if status != "" && status != "all" {
		query = `
			SELECT id, name, description, prompt, creator_id, generation_status, status, image_uuid, icon_uuid, created_at, updated_at
			FROM worlds
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{status, limit, offset}
	} else {
		query = `
			SELECT id, name, description, prompt, creator_id, generation_status, status, image_uuid, icon_uuid, created_at, updated_at
			FROM worlds
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	worlds := []*models.World{}
	err := r.db.SelectContext(ctx, &worlds, query, args...)
	if err != nil {
		logger.Logger.Error("Failed to get all worlds", zap.Error(err))
		return nil, 0, err
	}

	// Get total count
	var countQuery string
	var countArgs []interface{}

	if status != "" && status != "all" {
		countQuery = `SELECT COUNT(*) FROM worlds WHERE status = $1`
		countArgs = []interface{}{status}
	} else {
		countQuery = `SELECT COUNT(*) FROM worlds`
		countArgs = []interface{}{}
	}

	var total int
	err = r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		logger.Logger.Error("Failed to get worlds count", zap.Error(err))
		return nil, 0, err
	}

	return worlds, total, nil
}

// GetByUser retrieves worlds accessible to a user
func (r *PostgresWorldRepository) GetByUser(ctx context.Context, userID string, limit, offset int, status string) ([]*models.World, int, error) {
	var query string
	var args []interface{}

	if status != "" && status != "all" {
		query = `
			SELECT w.id, w.name, w.description, w.prompt, w.creator_id, w.generation_status, w.status, w.image_uuid, w.icon_uuid, w.created_at, w.updated_at
			FROM worlds w
			JOIN user_worlds uw ON w.id = uw.world_id
			WHERE uw.user_id = $1 AND w.status = $2
			ORDER BY w.created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{userID, status, limit, offset}
	} else {
		query = `
			SELECT w.id, w.name, w.description, w.prompt, w.creator_id, w.generation_status, w.status, w.image_uuid, w.icon_uuid, w.created_at, w.updated_at
			FROM worlds w
			JOIN user_worlds uw ON w.id = uw.world_id
			WHERE uw.user_id = $1
			ORDER BY w.created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{userID, limit, offset}
	}

	worlds := []*models.World{}
	err := r.db.SelectContext(ctx, &worlds, query, args...)
	if err != nil {
		logger.Logger.Error("Failed to get worlds by user", zap.Error(err), zap.String("user_id", userID))
		return nil, 0, err
	}

	// Get total count
	var countQuery string
	var countArgs []interface{}

	if status != "" && status != "all" {
		countQuery = `
			SELECT COUNT(*)
			FROM worlds w
			JOIN user_worlds uw ON w.id = uw.world_id
			WHERE uw.user_id = $1 AND w.status = $2
		`
		countArgs = []interface{}{userID, status}
	} else {
		countQuery = `
			SELECT COUNT(*)
			FROM worlds w
			JOIN user_worlds uw ON w.id = uw.world_id
			WHERE uw.user_id = $1
		`
		countArgs = []interface{}{userID}
	}

	var total int
	err = r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		logger.Logger.Error("Failed to get user worlds count", zap.Error(err))
		return nil, 0, err
	}

	return worlds, total, nil
}

// UpdateStatus updates the status of a world
func (r *PostgresWorldRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE worlds
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		logger.Logger.Error("Failed to update world status",
			zap.Error(err),
			zap.String("id", id),
			zap.String("status", status))
		return err
	}

	return nil
}

// AddUserToWorld adds a user to a world
func (r *PostgresWorldRepository) AddUserToWorld(ctx context.Context, userID, worldID string) error {
	// Check if relationship already exists
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM user_worlds
			WHERE user_id = $1 AND world_id = $2
		)
	`
	err := r.db.GetContext(ctx, &exists, checkQuery, userID, worldID)
	if err != nil {
		logger.Logger.Error("Failed to check if user-world relationship exists",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("world_id", worldID))
		return err
	}

	if exists {
		return nil // Relationship already exists
	}

	query := `
		INSERT INTO user_worlds (user_id, world_id)
		VALUES ($1, $2)
	`

	_, err = r.db.ExecContext(ctx, query, userID, worldID)
	if err != nil {
		logger.Logger.Error("Failed to add user to world",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("world_id", worldID))
		return err
	}

	return nil
}

// RemoveUserFromWorld removes a user from a world
func (r *PostgresWorldRepository) RemoveUserFromWorld(ctx context.Context, userID, worldID string) error {
	query := `
		DELETE FROM user_worlds
		WHERE user_id = $1 AND world_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, userID, worldID)
	if err != nil {
		logger.Logger.Error("Failed to remove user from world",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("world_id", worldID))
		return err
	}

	return nil
}

// GetUserWorlds gets all worlds a user has access to
func (r *PostgresWorldRepository) GetUserWorlds(ctx context.Context, userID string) ([]*models.UserWorld, error) {
	query := `
		SELECT id, user_id, world_id, created_at
		FROM user_worlds
		WHERE user_id = $1
	`

	userWorlds := []*models.UserWorld{}
	err := r.db.SelectContext(ctx, &userWorlds, query, userID)
	if err != nil {
		logger.Logger.Error("Failed to get user worlds", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	return userWorlds, nil
}

// CheckUserWorld checks if a user has access to a world
func (r *PostgresWorldRepository) CheckUserWorld(ctx context.Context, userID, worldID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_worlds
			WHERE user_id = $1 AND world_id = $2
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, userID, worldID)
	if err != nil {
		logger.Logger.Error("Failed to check user access to world",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("world_id", worldID))
		return false, err
	}

	return exists, nil
}

// GetWorldStats gets user and post counts for a world
func (r *PostgresWorldRepository) GetWorldStats(ctx context.Context, worldID string) (int, int, error) {

	var usersCount int
	var postsCount int

	// TODO: Request character service for users count
	// TODO: Request post service for posts count

	// TODO: move this to service layer

	return usersCount, postsCount, nil
}

// Update updates all fields of a world in the database
func (r *PostgresWorldRepository) Update(ctx context.Context, world *models.World) error {
	query := `
		UPDATE worlds
		SET name = $1,
		    description = $2,
		    prompt = $3,
		    status = $4,
		    generation_status = $5,
		    image_uuid = $6,
			icon_uuid = $7,
		    updated_at = NOW()
		WHERE id = $8
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		world.Name,
		world.Description,
		world.Prompt,
		world.Status,
		world.GenerationStatus,
		world.ImageUUID,
		world.IconUUID,
		world.ID,
	)
	if err != nil {
		logger.Logger.Error("Failed to update world",
			zap.Error(err),
			zap.String("id", world.ID))
		return err
	}

	return nil
}
