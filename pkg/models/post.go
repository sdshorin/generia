package models

import (
	"time"
)

// Post представляет модель поста
type Post struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username,omitempty"`
	Caption      string    `json:"caption"`
	ImageURL     string    `json:"image_url"`
	LikesCount   int       `json:"likes_count"`
	CommentsCount int      `json:"comments_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	UserLiked    bool      `json:"user_liked,omitempty"`
}

// CreatePostRequest представляет тело запроса на создание поста
type CreatePostRequest struct {
	Caption string `json:"caption"`
	Image   string `json:"image"` // Base64 encoded image
}

// Comment представляет модель комментария
type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateCommentRequest представляет тело запроса на создание комментария
type CreateCommentRequest struct {
	Text string `json:"text" validate:"required,min=1,max=500"`
}

// Like представляет модель лайка
type Like struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}