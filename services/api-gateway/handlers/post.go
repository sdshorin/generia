package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"instagram-clone/pkg/logger"
	"instagram-clone/services/api-gateway/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	interactionpb "instagram-clone/api/grpc/interaction"
	mediapb "instagram-clone/api/grpc/media"
	postpb "instagram-clone/api/grpc/post"
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
	Caption string `json:"caption"`
	MediaID string `json:"media_id"`
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

	// Create post
	resp, err := h.postClient.CreatePost(ctx, &postpb.CreatePostRequest{
		UserId:  userID,
		Caption: req.Caption,
		MediaId: req.MediaID,
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
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Caption       string    `json:"caption"`
	MediaURL      string    `json:"media_url"`
	CreatedAt     time.Time `json:"created_at"`
	LikesCount    int       `json:"likes_count"`
	CommentsCount int       `json:"comments_count"`
	UserLiked     bool      `json:"user_liked,omitempty"`
}

// GetPost handles requests to get a post by ID
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "PostHandler.GetPost")
	defer span.End()

	// Get post ID from URL path
	vars := mux.Vars(r)
	postID := vars["id"]
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Get post
	resp, err := h.postClient.GetPost(ctx, &postpb.GetPostRequest{
		PostId: postID,
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
		UserID:        resp.UserId,
		Username:      resp.Username,
		Caption:       resp.Caption,
		MediaURL:      resp.MediaUrl,
		CreatedAt:     createdAt,
		LikesCount:    int(resp.LikesCount),
		CommentsCount: int(resp.CommentsCount),
		UserLiked:     userLiked,
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

	// Get user ID from URL path
	vars := mux.Vars(r)
	userID := vars["user_id"]
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
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
		UserId: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
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
			UserID:        post.UserId,
			Username:      post.Username,
			Caption:       post.Caption,
			MediaURL:      post.MediaUrl,
			CreatedAt:     createdAt,
			LikesCount:    int(post.LikesCount),
			CommentsCount: int(post.CommentsCount),
			UserLiked:     userLiked,
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