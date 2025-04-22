package models

import (
	"time"
)

// Post represents a post in the system
type Post struct {
	ID           string    `db:"id"`
	CharacterID  string    `db:"character_id"`
	IsAI         bool      `db:"is_ai"`
	WorldID      string    `db:"world_id"`
	Caption      string    `db:"caption"`
	MediaID      string    `db:"media_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	DisplayName  string    // Not stored in DB, populated from Character service
	MediaURL     string    // Not stored in DB, populated from Media service
	LikesCount   int32     // Not stored in DB, populated from Interaction service
	CommentsCount int32    // Not stored in DB, populated from Interaction service
}

// PostWithStats represents a post with statistics
type PostWithStats struct {
	Post
	LikesCount    int32
	CommentsCount int32
}