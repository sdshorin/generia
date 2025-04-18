package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sdshorin/generia/pkg/config"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/ai-worker/internal/generators"
	"go.uber.org/zap"
)

// Task represents a generation task from the database
type Task struct {
	ID         string         `db:"id"`
	WorldID    string         `db:"world_id"`
	TaskType   string         `db:"task_type"`
	Status     string         `db:"status"`
	Parameters string         `db:"parameters"`
	Result     sql.NullString `db:"result"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

// World represents a world from the database
type World struct {
	ID               string    `db:"id"`
	Name             string    `db:"name"`
	Description      string    `db:"description"`
	Prompt           string    `db:"prompt"`
	CreatorID        string    `db:"creator_id"`
	GenerationStatus string    `db:"generation_status"`
	Status           string    `db:"status"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// TaskParameters represents the parameters for a generation task
type TaskParameters struct {
	Count       int    `json:"count"`
	WorldPrompt string `json:"world_prompt"`
}

// TaskResult represents the result of a generation task
type TaskResult struct {
	Message        string   `json:"message"`
	GeneratedCount int      `json:"generated_count"`
	GeneratedIDs   []string `json:"generated_ids"`
}

func main() {
	// Initialize logger
	if err := logger.InitProduction(); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize generators
	userGenerator := generators.NewUserGenerator()
	postGenerator := generators.NewPostGenerator()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdown
		logger.Logger.Info("Shutting down AI worker...")
		cancel()
	}()

	// Start the worker loop
	logger.Logger.Info("Starting AI worker...")
	for {
		select {
		case <-ctx.Done():
			logger.Logger.Info("AI worker stopped")
			return
		default:
			// Look for pending tasks
			task, err := fetchNextTask(ctx, db)
			if err != nil {
				logger.Logger.Error("Failed to fetch next task", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			if task == nil {
				// No tasks found, wait and try again
				logger.Logger.Debug("No pending tasks found")
				time.Sleep(5 * time.Second)
				continue
			}

			// Process the task
			logger.Logger.Info("Processing task", 
				zap.String("task_id", task.ID), 
				zap.String("task_type", task.TaskType),
				zap.String("world_id", task.WorldID))

			// Update task status to in progress
			err = updateTaskStatus(ctx, db, task.ID, "in_progress", `{"message": "Processing task..."}`)
			if err != nil {
				logger.Logger.Error("Failed to update task status", zap.Error(err), zap.String("task_id", task.ID))
				time.Sleep(1 * time.Second)
				continue
			}

			// Get world info
			world, err := getWorld(ctx, db, task.WorldID)
			if err != nil {
				logger.Logger.Error("Failed to get world info", zap.Error(err), zap.String("world_id", task.WorldID))
				_ = updateTaskStatus(ctx, db, task.ID, "failed", `{"message": "Failed to get world info"}`)
				continue
			}

			// Process task based on type
			var result string
			if task.TaskType == "users" {
				result, err = processUsersTask(ctx, db, task, world, userGenerator)
			} else if task.TaskType == "posts" {
				result, err = processPostsTask(ctx, db, task, world, postGenerator)
			} else {
				result = `{"message": "Unknown task type"}`
				err = fmt.Errorf("unknown task type: %s", task.TaskType)
			}

			if err != nil {
				logger.Logger.Error("Failed to process task", 
					zap.Error(err), 
					zap.String("task_id", task.ID),
					zap.String("task_type", task.TaskType))
				_ = updateTaskStatus(ctx, db, task.ID, "failed", 
					fmt.Sprintf(`{"message": "Failed to process task: %s"}`, err.Error()))
				continue
			}

			// Update task with result
			err = updateTaskStatus(ctx, db, task.ID, "completed", result)
			if err != nil {
				logger.Logger.Error("Failed to update task result", zap.Error(err), zap.String("task_id", task.ID))
				continue
			}

			// Check if all tasks for this world are completed
			allCompleted, err := checkAllTasksCompleted(ctx, db, task.WorldID)
			if err != nil {
				logger.Logger.Error("Failed to check all tasks completed", 
					zap.Error(err), 
					zap.String("world_id", task.WorldID))
				continue
			}

			if allCompleted {
				// Update world generation status to completed
				err = updateWorldStatus(ctx, db, task.WorldID, "completed")
				if err != nil {
					logger.Logger.Error("Failed to update world status", 
						zap.Error(err), 
						zap.String("world_id", task.WorldID))
					continue
				}
				logger.Logger.Info("All tasks completed for world", zap.String("world_id", task.WorldID))
			}

			// Small delay to prevent hammering the database
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// fetchNextTask fetches the next pending task from the database
func fetchNextTask(ctx context.Context, db *sqlx.DB) (*Task, error) {
	query := `
		SELECT id, world_id, task_type, status, parameters, result, created_at, updated_at
		FROM world_generation_tasks
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT 1
	`

	var task Task
	err := db.GetContext(ctx, &task, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &task, nil
}

// updateTaskStatus updates the status and result of a task
func updateTaskStatus(ctx context.Context, db *sqlx.DB, taskID, status, result string) error {
	query := `
		UPDATE world_generation_tasks
		SET status = $1, result = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := db.ExecContext(ctx, query, status, result, taskID)
	return err
}

// getWorld gets the world info from the database
func getWorld(ctx context.Context, db *sqlx.DB, worldID string) (*World, error) {
	query := `
		SELECT id, name, description, prompt, creator_id, generation_status, status, created_at, updated_at
		FROM worlds
		WHERE id = $1
	`

	var world World
	err := db.GetContext(ctx, &world, query)
	if err != nil {
		return nil, err
	}

	return &world, nil
}

// updateWorldStatus updates the generation status of a world
func updateWorldStatus(ctx context.Context, db *sqlx.DB, worldID, status string) error {
	query := `
		UPDATE worlds
		SET generation_status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := db.ExecContext(ctx, query, status, worldID)
	return err
}

// checkAllTasksCompleted checks if all tasks for a world are completed
func checkAllTasksCompleted(ctx context.Context, db *sqlx.DB, worldID string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM world_generation_tasks
		WHERE world_id = $1 AND status != 'completed'
	`

	var count int
	err := db.GetContext(ctx, &count, query, worldID)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// processUsersTask processes a users generation task
func processUsersTask(ctx context.Context, db *sqlx.DB, task *Task, world *World, generator *generators.UserGenerator) (string, error) {
	// Parse task parameters
	var params TaskParameters
	err := json.Unmarshal([]byte(task.Parameters), &params)
	if err != nil {
		return "", fmt.Errorf("failed to parse task parameters: %w", err)
	}

	// If no world prompt provided in parameters, use the one from the world
	if params.WorldPrompt == "" {
		params.WorldPrompt = world.Prompt
	}

	// Use a default count if not specified
	if params.Count <= 0 {
		params.Count = 10
	}

	// Generate users
	generatedIDs := make([]string, 0, params.Count)
	for i := 0; i < params.Count; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// Generate a user
			userID, username, description, err := generator.GenerateUser(ctx, world.ID, params.WorldPrompt)
			if err != nil {
				return "", fmt.Errorf("failed to generate user: %w", err)
			}

			// Insert the user into the database
			// In a real implementation, you might want to batch these inserts for better performance
			err = insertAIUser(ctx, db, userID, username, description, world.ID)
			if err != nil {
				return "", fmt.Errorf("failed to insert AI user: %w", err)
			}

			generatedIDs = append(generatedIDs, userID)
		}
	}

	// Create result JSON
	result := TaskResult{
		Message:        fmt.Sprintf("Generated %d AI users", len(generatedIDs)),
		GeneratedCount: len(generatedIDs),
		GeneratedIDs:   generatedIDs,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}

// processPostsTask processes a posts generation task
func processPostsTask(ctx context.Context, db *sqlx.DB, task *Task, world *World, generator *generators.PostGenerator) (string, error) {
	// Parse task parameters
	var params TaskParameters
	err := json.Unmarshal([]byte(task.Parameters), &params)
	if err != nil {
		return "", fmt.Errorf("failed to parse task parameters: %w", err)
	}

	// If no world prompt provided in parameters, use the one from the world
	if params.WorldPrompt == "" {
		params.WorldPrompt = world.Prompt
	}

	// Use a default count if not specified
	if params.Count <= 0 {
		params.Count = 50
	}

	// Get AI users for this world
	users, err := getAIUsers(ctx, db, world.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get AI users: %w", err)
	}

	if len(users) == 0 {
		return "", fmt.Errorf("no AI users found for world")
	}

	// Generate posts
	generatedIDs := make([]string, 0, params.Count)
	for i := 0; i < params.Count; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// Select a random user
			userID := users[i%len(users)]

			// Generate a post
			postID, caption, _, err := generator.GeneratePost(ctx, world.ID, params.WorldPrompt, userID)
			if err != nil {
				return "", fmt.Errorf("failed to generate post: %w", err)
			}

			// In a real implementation, you would generate an image here
			// For now, we'll just use a placeholder
			mediaID := fmt.Sprintf("placeholder-%s", postID[:8])

			// Insert the post into the database
			err = insertAIPost(ctx, db, postID, userID, world.ID, caption, mediaID)
			if err != nil {
				return "", fmt.Errorf("failed to insert AI post: %w", err)
			}

			generatedIDs = append(generatedIDs, postID)
		}
	}

	// Create result JSON
	result := TaskResult{
		Message:        fmt.Sprintf("Generated %d AI posts", len(generatedIDs)),
		GeneratedCount: len(generatedIDs),
		GeneratedIDs:   generatedIDs,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}

// insertAIUser inserts an AI user into the database
func insertAIUser(ctx context.Context, db *sqlx.DB, userID, username, description, worldID string) error {
	query := `
		INSERT INTO users (id, username, is_ai, world_id)
		VALUES ($1, $2, true, $3)
	`

	_, err := db.ExecContext(ctx, query, userID, username, worldID)
	return err
}

// getAIUsers gets all AI users for a world
func getAIUsers(ctx context.Context, db *sqlx.DB, worldID string) ([]string, error) {
	query := `
		SELECT id
		FROM users
		WHERE world_id = $1 AND is_ai = true
	`

	var userIDs []string
	err := db.SelectContext(ctx, &userIDs, query, worldID)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

// insertAIPost inserts an AI post into the database
func insertAIPost(ctx context.Context, db *sqlx.DB, postID, userID, worldID, caption, mediaID string) error {
	query := `
		INSERT INTO posts (id, user_id, world_id, caption, media_id)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := db.ExecContext(ctx, query, postID, userID, worldID, caption, mediaID)
	return err
}