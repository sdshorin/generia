package models

import (
	"database/sql"
	"time"
)

// World represents a world in the application
type World struct {
	ID               string         `db:"id"`
	Name             string         `db:"name"`
	Description      string         `db:"description"`
	Prompt           string         `db:"prompt"`
	CreatorID        string         `db:"creator_id"`
	Status           string         `db:"status"`
	GenerationStatus string         `db:"generation_status"`
	ImageUUID        sql.NullString `db:"image_uuid"`
	IconUUID         sql.NullString `db:"icon_uuid"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
}

// UserWorld represents a user's access to a world
type UserWorld struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	WorldID   string    `db:"world_id"`
	CreatedAt time.Time `db:"created_at"`
}

// Constants for world status
const (
	WorldStatusActive   = "active"
	WorldStatusArchived = "archived"
)
