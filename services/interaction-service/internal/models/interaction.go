package models

import (
	"time"
)

// Like represents a like on a post
type Like struct {
	ID        string    `bson:"_id,omitempty"`
	PostID    string    `bson:"post_id"`
	UserID    string    `bson:"user_id"`
	CreatedAt time.Time `bson:"created_at"`
}

// Comment represents a comment on a post
type Comment struct {
	ID        string    `bson:"_id,omitempty"`
	PostID    string    `bson:"post_id"`
	UserID    string    `bson:"user_id"`
	Text      string    `bson:"text"`
	CreatedAt time.Time `bson:"created_at"`
}

// PostStats represents the stats (likes and comments) for a post
type PostStats struct {
	PostID        string    `bson:"_id"` // Using post_id as the _id
	LikesCount    int32     `bson:"likes_count"`
	CommentsCount int32     `bson:"comments_count"`
	UpdatedAt     time.Time `bson:"updated_at"`
}