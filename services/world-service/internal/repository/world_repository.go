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
	UpdateGenerationStatus(ctx context.Context, id, status string) error

	// User world operations
	AddUserToWorld(ctx context.Context, userID, worldID string) error
	RemoveUserFromWorld(ctx context.Context, userID, worldID string) error
	GetUserWorlds(ctx context.Context, userID string) ([]*models.UserWorld, error)
	SetActiveWorld(ctx context.Context, userID, worldID string) error
	GetActiveWorld(ctx context.Context, userID string) (*models.World, error)
	CheckUserWorld(ctx context.Context, userID, worldID string) (bool, error)

	// Generation tasks
	CreateGenerationTask(ctx context.Context, task *models.WorldGenerationTask) error
	GetGenerationTasks(ctx context.Context, worldID string) ([]*models.WorldGenerationTask, error)
	UpdateGenerationTask(ctx context.Context, taskID, status, result string) error
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
		INSERT INTO worlds (name, description, prompt, creator_id, generation_status, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	row := r.db.QueryRowContext(
		ctx,
		query,
		world.Name,
		world.Description,
		world.Prompt,
		world.CreatorID,
		world.GenerationStatus,
		world.Status,
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
		SELECT id, name, description, prompt, creator_id, generation_status, status, created_at, updated_at
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
			SELECT id, name, description, prompt, creator_id, generation_status, status, created_at, updated_at
			FROM worlds
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{status, limit, offset}
	} else {
		query = `
			SELECT id, name, description, prompt, creator_id, generation_status, status, created_at, updated_at
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
			SELECT w.id, w.name, w.description, w.prompt, w.creator_id, w.generation_status, w.status, w.created_at, w.updated_at
			FROM worlds w
			JOIN user_worlds uw ON w.id = uw.world_id
			WHERE uw.user_id = $1 AND w.status = $2
			ORDER BY w.created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{userID, status, limit, offset}
	} else {
		query = `
			SELECT w.id, w.name, w.description, w.prompt, w.creator_id, w.generation_status, w.status, w.created_at, w.updated_at
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

// UpdateGenerationStatus updates the generation status of a world
func (r *PostgresWorldRepository) UpdateGenerationStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE worlds
		SET generation_status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		logger.Logger.Error("Failed to update world generation status", 
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

// SetActiveWorld sets a world as active for a user
func (r *PostgresWorldRepository) SetActiveWorld(ctx context.Context, userID, worldID string) error {
	// First, check if the user has access to the world
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM user_worlds
			WHERE user_id = $1 AND world_id = $2
		)
	`
	err := r.db.GetContext(ctx, &exists, checkQuery, userID, worldID)
	if err != nil {
		logger.Logger.Error("Failed to check if user has access to world", 
			zap.Error(err), 
			zap.String("user_id", userID), 
			zap.String("world_id", worldID))
		return err
	}

	if !exists {
		// Add user to world if they don't have access
		err = r.AddUserToWorld(ctx, userID, worldID)
		if err != nil {
			return err
		}
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Logger.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	// Reset all active worlds for user
	resetQuery := `
		UPDATE user_worlds
		SET is_active = false
		WHERE user_id = $1
	`
	_, err = tx.ExecContext(ctx, resetQuery, userID)
	if err != nil {
		logger.Logger.Error("Failed to reset active worlds", 
			zap.Error(err), 
			zap.String("user_id", userID))
		return err
	}

	// Set the world as active
	setQuery := `
		UPDATE user_worlds
		SET is_active = true
		WHERE user_id = $1 AND world_id = $2
	`
	_, err = tx.ExecContext(ctx, setQuery, userID, worldID)
	if err != nil {
		logger.Logger.Error("Failed to set active world", 
			zap.Error(err), 
			zap.String("user_id", userID), 
			zap.String("world_id", worldID))
		return err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		logger.Logger.Error("Failed to commit transaction", zap.Error(err))
		return err
	}

	return nil
}

// GetActiveWorld gets the active world for a user
func (r *PostgresWorldRepository) GetActiveWorld(ctx context.Context, userID string) (*models.World, error) {
	query := `
		SELECT w.id, w.name, w.description, w.prompt, w.creator_id, w.generation_status, w.status, w.created_at, w.updated_at
		FROM worlds w
		JOIN user_worlds uw ON w.id = uw.world_id
		WHERE uw.user_id = $1 AND uw.is_active = true
		LIMIT 1
	`

	var world models.World
	err := r.db.GetContext(ctx, &world, query, userID)
	if err != nil {
		// If no active world is found, this is not an error
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		logger.Logger.Error("Failed to get active world", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	return &world, nil
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

// CreateGenerationTask creates a new generation task
func (r *PostgresWorldRepository) CreateGenerationTask(ctx context.Context, task *models.WorldGenerationTask) error {
	query := `
		INSERT INTO world_generation_tasks (world_id, task_type, status, parameters)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	row := r.db.QueryRowContext(
		ctx,
		query,
		task.WorldID,
		task.TaskType,
		task.Status,
		task.Parameters,
	)

	err := row.Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		logger.Logger.Error("Failed to create generation task", zap.Error(err))
		return err
	}

	return nil
}

// GetGenerationTasks gets all generation tasks for a world
func (r *PostgresWorldRepository) GetGenerationTasks(ctx context.Context, worldID string) ([]*models.WorldGenerationTask, error) {
	query := `
		SELECT id, world_id, task_type, status, parameters, result, created_at, updated_at
		FROM world_generation_tasks
		WHERE world_id = $1
		ORDER BY created_at DESC
	`

	tasks := []*models.WorldGenerationTask{}
	err := r.db.SelectContext(ctx, &tasks, query, worldID)
	if err != nil {
		logger.Logger.Error("Failed to get generation tasks", zap.Error(err), zap.String("world_id", worldID))
		return nil, err
	}

	return tasks, nil
}

// UpdateGenerationTask updates a generation task's status and result
func (r *PostgresWorldRepository) UpdateGenerationTask(ctx context.Context, taskID, status, result string) error {
	query := `
		UPDATE world_generation_tasks
		SET status = $1, result = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, status, result, taskID)
	if err != nil {
		logger.Logger.Error("Failed to update generation task", 
			zap.Error(err), 
			zap.String("task_id", taskID), 
			zap.String("status", status))
		return err
	}

	return nil
}

// GetWorldStats gets user and post counts for a world
func (r *PostgresWorldRepository) GetWorldStats(ctx context.Context, worldID string) (int, int, error) {
	// Get AI users count
	usersQuery := `
		SELECT COUNT(*) FROM users
		WHERE world_id = $1 AND is_ai = true
	`

	var usersCount int
	err := r.db.GetContext(ctx, &usersCount, usersQuery, worldID)
	if err != nil {
		logger.Logger.Error("Failed to get world users count", zap.Error(err), zap.String("world_id", worldID))
		return 0, 0, err
	}

	// Get posts count
	postsQuery := `
		SELECT COUNT(*) FROM posts
		WHERE world_id = $1
	`

	var postsCount int
	err = r.db.GetContext(ctx, &postsCount, postsQuery, worldID)
	if err != nil {
		logger.Logger.Error("Failed to get world posts count", zap.Error(err), zap.String("world_id", worldID))
		return 0, 0, err
	}

	return usersCount, postsCount, nil
}