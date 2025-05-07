// Package handlers provides HTTP handlers for the API Gateway
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/api-gateway/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	interactionpb "github.com/sdshorin/generia/api/grpc/interaction"
	mediapb "github.com/sdshorin/generia/api/grpc/media"
	postpb "github.com/sdshorin/generia/api/grpc/post"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postClient        postpb.PostServiceClient
	mediaClient       mediapb.MediaServiceClient
	interactionClient interactionpb.InteractionServiceClient
	tracer            trace.Tracer
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(
	postClient postpb.PostServiceClient,
	mediaClient mediapb.MediaServiceClient,
	interactionClient interactionpb.InteractionServiceClient,
	tracer trace.Tracer,
) *PostHandler {
	return &PostHandler{
		postClient:        postClient,
		mediaClient:       mediaClient,
		interactionClient: interactionClient,
		tracer:            tracer,
	}
}

// CreatePostRequest represents a request to create a post
type CreatePostRequest struct {
	Caption     string `json:"caption"`
	MediaID     string `json:"media_id"`
	CharacterID string `json:"character_id"` // Required character ID
}

// CreatePostResponse represents a response after creating a post
type CreatePostResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreatePost handles requests to create a new post
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PostHandler.CreatePost")
	defer span.End()

	// Get user ID from context
	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Parse request body
	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to decode request body", zap.Error(err))
		return
	}

	// Validate request
	if req.MediaID == "" {
		http.Error(w, "Media ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	if req.CharacterID == "" {
		http.Error(w, "Character ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Получаем world_id из URL параметров
	vars := mux.Vars(r)
	worldID := vars["world_id"]
	if worldID == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Create post
	resp, err := h.postClient.CreatePost(ctx, &postpb.CreatePostRequest{
		UserId:      userID,
		CharacterId: req.CharacterID,
		Caption:     req.Caption,
		MediaId:     req.MediaID,
		WorldId:     worldID,
	})
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to create post", zap.Error(err))
		return
	}

	// Parse created time
	createdAt, err := time.Parse(time.RFC3339, resp.CreatedAt)
	if err != nil {
		createdAt = time.Now() // Fallback
	}

	// Prepare response
	response := CreatePostResponse{
		ID:        resp.PostId,
		CreatedAt: createdAt,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// PostResponse represents a post in the API response
type PostResponse struct {
	ID            string    `json:"id"`
	CharacterID   string    `json:"character_id"`
	DisplayName   string    `json:"display_name"`
	Caption       string    `json:"caption"`
	MediaURL      string    `json:"media_url"`
	AvatarURL     string    `json:"avatar_url"`
	CreatedAt     time.Time `json:"created_at"`
	LikesCount    int       `json:"likes_count"`
	CommentsCount int       `json:"comments_count"`
	UserLiked     bool      `json:"user_liked,omitempty"`
	IsAI          bool      `json:"is_ai"`
}

// GetPost handles requests to get a post by ID
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PostHandler.GetPost")
	defer span.End()

	// Get post ID and world ID from URL path
	vars := mux.Vars(r)
	postID := vars["id"]
	worldID := vars["world_id"]

	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	if worldID == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Get post
	resp, err := h.postClient.GetPost(ctx, &postpb.GetPostRequest{
		PostId:  postID,
		WorldId: worldID,
	})
	if err != nil {
		http.Error(w, "Failed to get post", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get post", zap.Error(err), zap.String("post_id", postID))
		return
	}

	// Check if user has liked this post (if authenticated)
	userLiked := false
	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if ok && userID != "" {
		likeResp, err := h.interactionClient.CheckUserLiked(ctx, &interactionpb.CheckUserLikedRequest{
			PostId: postID,
			UserId: userID,
		})
		if err == nil {
			userLiked = likeResp.Liked
		}
	}

	// Parse created time
	createdAt, err := time.Parse(time.RFC3339, resp.CreatedAt)
	if err != nil {
		createdAt = time.Now() // Fallback
	}

	// Prepare response
	response := PostResponse{
		ID:            resp.PostId,
		CharacterID:   resp.CharacterId,
		DisplayName:   resp.DisplayName,
		Caption:       resp.Caption,
		MediaURL:      resp.MediaUrl,
		AvatarURL:     resp.AvatarUrl,
		CreatedAt:     createdAt,
		LikesCount:    int(resp.LikesCount),
		CommentsCount: int(resp.CommentsCount),
		UserLiked:     userLiked,
		IsAI:          resp.IsAi,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// FeedResponse represents a paginated list of posts
type FeedResponse struct {
	Posts      []PostResponse `json:"posts"`
	Total      int            `json:"total"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

// GetUserPosts handles requests to get posts by user ID
func (h *PostHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PostHandler.GetUserPosts")
	defer span.End()

	// Get user ID and world ID from URL path
	vars := mux.Vars(r)
	userID := vars["user_id"]
	worldID := vars["world_id"]

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	if worldID == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Parse pagination parameters
	limit := 10 // Default
	offset := 0 // Default

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get posts
	resp, err := h.postClient.GetUserPosts(ctx, &postpb.GetUserPostsRequest{
		UserId:  userID,
		WorldId: worldID,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		http.Error(w, "Failed to get user posts", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get user posts", zap.Error(err), zap.String("user_id", userID))
		return
	}

	// Check if current user has liked these posts (if authenticated)
	currentUserID, _ := ctx.Value(middleware.UserIDKey).(string)

	// Prepare response
	posts := make([]PostResponse, 0, len(resp.Posts))
	for _, post := range resp.Posts {
		// Check if current user has liked this post
		userLiked := false
		if currentUserID != "" {
			likeResp, err := h.interactionClient.CheckUserLiked(ctx, &interactionpb.CheckUserLikedRequest{
				PostId: post.PostId,
				UserId: currentUserID,
			})
			if err == nil {
				userLiked = likeResp.Liked
			}
		}

		// Parse created time
		createdAt, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			createdAt = time.Now() // Fallback
		}

		posts = append(posts, PostResponse{
			ID:            post.PostId,
			CharacterID:   post.CharacterId,
			DisplayName:   post.DisplayName,
			Caption:       post.Caption,
			MediaURL:      post.MediaUrl,
			AvatarURL:     post.AvatarUrl,
			CreatedAt:     createdAt,
			LikesCount:    int(post.LikesCount),
			CommentsCount: int(post.CommentsCount),
			UserLiked:     userLiked,
			IsAI:          post.IsAi,
		})
	}

	// Send response
	response := FeedResponse{
		Posts: posts,
		Total: int(resp.Total),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// GetCharacterPosts handles requests to get posts for a specific character in a world
func (h *PostHandler) GetCharacterPosts(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PostHandler.GetCharacterPosts")
	defer span.End()

	// Get parameters from URL
	vars := mux.Vars(r)
	worldID := vars["world_id"]
	characterID := vars["character_id"]

	if worldID == "" || characterID == "" {
		http.Error(w, "World ID and Character ID are required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Get pagination parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20 // Default limit
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Call PostService
	resp, err := h.postClient.GetCharacterPosts(ctx, &postpb.GetCharacterPostsRequest{
		CharacterId: characterID,
		WorldId:     worldID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})

	if err != nil {
		http.Error(w, "Failed to get character posts", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get character posts", zap.Error(err))
		return
	}

	// Convert posts to response format
	posts := make([]PostResponse, len(resp.Posts))
	for i, post := range resp.Posts {
		posts[i] = PostResponse{
			ID:            post.PostId,
			CharacterID:   post.CharacterId,
			DisplayName:   post.DisplayName,
			Caption:       post.Caption,
			MediaURL:      post.MediaUrl,
			AvatarURL:     post.AvatarUrl,
			CreatedAt:     time.Unix(0, 0), // TODO: Parse created_at from string
			LikesCount:    int(post.LikesCount),
			CommentsCount: int(post.CommentsCount),
			IsAI:          post.IsAi,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FeedResponse{
		Posts: posts,
		Total: len(posts),
	})
}

// GetGlobalPosts handles requests to get the global feed directly from post service
func (h *PostHandler) GetGlobalPosts(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PostHandler.GetGlobalPosts")
	defer span.End()

	// Get user ID from context (if authenticated)
	userID, _ := ctx.Value(middleware.UserIDKey).(string)

	// Parse pagination parameters
	limit := 10 // Default
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	cursor := r.URL.Query().Get("cursor")

	// Get world_id from URL parameters
	vars := mux.Vars(r)
	worldID := vars["world_id"]
	if worldID == "" {
		http.Error(w, "world_id is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Missing world_id parameter")
		return
	}

	// Get posts directly from post service
	resp, err := h.postClient.GetGlobalFeed(ctx, &postpb.GetGlobalFeedRequest{
		Limit:   int32(limit),
		Cursor:  cursor,
		WorldId: worldID,
	})
	if err != nil {
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get posts", zap.Error(err))
		return
	}

	// Prepare response
	posts := make([]PostResponse, 0, len(resp.Posts))
	for _, post := range resp.Posts {
		// Check if current user has liked this post
		userLiked := false
		if userID != "" {
			likeResp, err := h.interactionClient.CheckUserLiked(ctx, &interactionpb.CheckUserLikedRequest{
				PostId: post.PostId,
				UserId: userID,
			})
			if err == nil {
				userLiked = likeResp.Liked
			}
		}

		// Parse created time
		createdAt, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			createdAt = time.Now() // Fallback
		}

		posts = append(posts, PostResponse{
			ID:            post.PostId,
			CharacterID:   post.CharacterId,
			DisplayName:   post.DisplayName,
			Caption:       post.Caption,
			MediaURL:      post.MediaUrl,
			AvatarURL:     post.AvatarUrl,
			CreatedAt:     createdAt,
			LikesCount:    int(post.LikesCount),
			CommentsCount: int(post.CommentsCount),
			UserLiked:     userLiked,
			IsAI:          post.IsAi,
		})
	}

	// Send response
	response := FeedResponse{
		Posts:      posts,
		Total:      len(posts),
		NextCursor: resp.NextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}
