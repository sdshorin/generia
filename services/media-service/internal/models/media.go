package models

import (
	"time"
)

// Media represents a media entity in the database
type Media struct {
	ID          string    `db:"id" json:"id"`
	UserID      string    `db:"user_id" json:"user_id"`
	Filename    string    `db:"filename" json:"filename"`
	ContentType string    `db:"content_type" json:"content_type"`
	Size        int64     `db:"size" json:"size"`
	BucketName  string    `db:"bucket_name" json:"bucket_name"`
	ObjectName  string    `db:"object_name" json:"object_name"`
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