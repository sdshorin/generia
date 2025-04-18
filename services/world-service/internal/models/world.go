package models

import (
	"time"
)

// World represents a world in the application
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

// UserWorld represents a user's access to a world
type UserWorld struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	WorldID   string    `db:"world_id"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
}

// WorldGenerationTask represents a task for generating content for a world
type WorldGenerationTask struct {
	ID        string    `db:"id"`
	WorldID   string    `db:"world_id"`
	TaskType  string    `db:"task_type"`
	Status    string    `db:"status"`
	Parameters string    `db:"parameters"`
	Result    string    `db:"result"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Constants for world generation status
const (
	GenerationStatusPending    = "pending"
	GenerationStatusInProgress = "in_progress"
	GenerationStatusCompleted  = "completed"
	GenerationStatusFailed     = "failed"
)

// Constants for world status
const (
	WorldStatusActive   = "active"
	WorldStatusArchived = "archived"
)

// Constants for task types
const (
	TaskTypeUsers = "users"
	TaskTypePosts = "posts"
	TaskTypeMedia = "media"
)

// Constants for task status
const (
	TaskStatusPending    = "pending"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)