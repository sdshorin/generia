package models

import (
	"fmt"
	"path/filepath"
	"time"
)

// Media type constants (matching proto enum)
const (
	MediaTypeUnknown         = 0
	MediaTypeWorldHeader     = 1
	MediaTypeWorldIcon       = 2
	MediaTypeCharacterAvatar = 3
	MediaTypePostImage       = 4
)

// Helper functions for pointer handling
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// GenerateObjectName generates the object name based on media type and parameters
func GenerateObjectName(worldId, characterId, mediaId, filename string, mediaType int32) string {
	ext := filepath.Ext(filename)

	switch mediaType {
	case MediaTypeWorldHeader:
		return fmt.Sprintf("%s/world_data/%s%s", worldId, mediaId, ext)
	case MediaTypeWorldIcon:
		return fmt.Sprintf("%s/world_data/%s%s", worldId, mediaId, ext)
	case MediaTypeCharacterAvatar:
		return fmt.Sprintf("%s/%s/avatars/%s%s", worldId, characterId, mediaId, ext)
	case MediaTypePostImage:
		return fmt.Sprintf("%s/%s/posts/%s%s", worldId, characterId, mediaId, ext)
	default:
		// Fallback to old format for unknown types
		return fmt.Sprintf("%s/%s%s", characterId, mediaId, ext)
	}
}

// Media represents a media entity in the database
type Media struct {
	ID          string    `db:"id" json:"id"`
	CharacterId *string   `db:"character_id" json:"character_id"` // Nullable for world-level media
	WorldId     string    `db:"world_id" json:"world_id"`
	Filename    string    `db:"filename" json:"filename"`
	ContentType string    `db:"content_type" json:"content_type"`
	Size        int64     `db:"size" json:"size"`
	BucketName  string    `db:"bucket" json:"bucket"`
	ObjectName  string    `db:"object_name" json:"object_name"`
	MediaType   int32     `db:"media_type" json:"media_type"` // Corresponds to proto MediaType enum
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// MediaVariant represents a variant of a media (e.g., thumbnail, medium, original)
type MediaVariant struct {
	ID        string    `db:"id" json:"id"`
	MediaID   string    `db:"media_id" json:"media_id"`
	Name      string    `db:"name" json:"name"`
	URL       string    `db:"url" json:"url"`
	Width     int32     `db:"width" json:"width"`
	Height    int32     `db:"height" json:"height"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
