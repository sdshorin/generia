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
)

// InteractionHandler handles interaction-related HTTP requests
type InteractionHandler struct {
	interactionClient interactionpb.InteractionServiceClient
	tracer            trace.Tracer
}

// NewInteractionHandler creates a new InteractionHandler
func NewInteractionHandler(
	interactionClient interactionpb.InteractionServiceClient,
	tracer trace.Tracer,
) *InteractionHandler {
	return &InteractionHandler{
		interactionClient: interactionClient,
		tracer:            tracer,
	}
}

// LikePostResponse represents the response after liking a post
type LikePostResponse struct {
	Success    bool `json:"success"`
	LikesCount int  `json:"likes_count"`
}

// LikePost handles requests to like a post
func (h *InteractionHandler) LikePost(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "InteractionHandler.LikePost")
	defer span.End()

	// Get user ID from context
	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Get post ID from URL path
	vars := mux.Vars(r)
	postID := vars["id"]
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Like the post
	resp, err := h.interactionClient.LikePost(ctx, &interactionpb.LikePostRequest{
		PostId: postID,
		UserId: userID,
	})
	if err != nil {
		http.Error(w, "Failed to like post", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to like post", zap.Error(err), zap.String("post_id", postID))
		return
	}

	// Prepare response
	response := LikePostResponse{
		Success:    resp.Success,
		LikesCount: int(resp.LikesCount),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// UnlikePost handles requests to unlike a post
func (h *InteractionHandler) UnlikePost(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "InteractionHandler.UnlikePost")
	defer span.End()

	// Get user ID from context
	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Get post ID from URL path
	vars := mux.Vars(r)
	postID := vars["id"]
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Unlike the post
	resp, err := h.interactionClient.UnlikePost(ctx, &interactionpb.UnlikePostRequest{
		PostId: postID,
		UserId: userID,
	})
	if err != nil {
		http.Error(w, "Failed to unlike post", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to unlike post", zap.Error(err), zap.String("post_id", postID))
		return
	}

	// Prepare response
	response := LikePostResponse{
		Success:    resp.Success,
		LikesCount: int(resp.LikesCount),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// AddCommentRequest represents a request to add a comment
type AddCommentRequest struct {
	Text string `json:"text"`
}

// AddCommentResponse represents the response after adding a comment
type AddCommentResponse struct {
	CommentID string    `json:"comment_id"`
	CreatedAt time.Time `json:"created_at"`
}

// AddComment handles requests to add a comment to a post
func (h *InteractionHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "InteractionHandler.AddComment")
	defer span.End()

	// Get user ID from context
	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Get post ID from URL path
	vars := mux.Vars(r)
	postID := vars["id"]
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Parse request body
	var req AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to decode request body", zap.Error(err))
		return
	}

	// Validate request
	if req.Text == "" {
		http.Error(w, "Comment text is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Add comment
	resp, err := h.interactionClient.AddComment(ctx, &interactionpb.AddCommentRequest{
		PostId: postID,
		UserId: userID,
		Text:   req.Text,
	})
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to add comment", zap.Error(err), zap.String("post_id", postID))
		return
	}

	// Parse created time
	createdAt, err := time.Parse(time.RFC3339, resp.CreatedAt)
	if err != nil {
		createdAt = time.Now() // Fallback
	}

	// Prepare response
	response := AddCommentResponse{
		CommentID: resp.CommentId,
		CreatedAt: createdAt,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// CommentResponse represents a comment in the API response
type CommentResponse struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// CommentsResponse represents a paginated list of comments
type CommentsResponse struct {
	Comments []CommentResponse `json:"comments"`
	Total    int               `json:"total"`
}

// GetComments handles requests to get comments for a post
func (h *InteractionHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "InteractionHandler.GetComments")
	defer span.End()

	// Get post ID from URL path
	vars := mux.Vars(r)
	postID := vars["id"]
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
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

	// Get comments
	resp, err := h.interactionClient.GetPostComments(ctx, &interactionpb.GetPostCommentsRequest{
		PostId: postID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get comments", zap.Error(err), zap.String("post_id", postID))
		return
	}

	// Prepare response
	comments := make([]CommentResponse, 0, len(resp.Comments))
	for _, comment := range resp.Comments {
		// Parse created time
		createdAt, err := time.Parse(time.RFC3339, comment.CreatedAt)
		if err != nil {
			createdAt = time.Now() // Fallback
		}

		comments = append(comments, CommentResponse{
			ID:        comment.CommentId,
			PostID:    comment.PostId,
			UserID:    comment.UserId,
			Username:  comment.Username,
			Text:      comment.Text,
			CreatedAt: createdAt,
		})
	}

	// Send response
	response := CommentsResponse{
		Comments: comments,
		Total:    int(resp.Total),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// LikeResponse represents a like in the API response
type LikeResponse struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// LikesResponse represents a paginated list of likes
type LikesResponse struct {
	Likes []LikeResponse `json:"likes"`
	Total int            `json:"total"`
}

// GetLikes handles requests to get likes for a post
func (h *InteractionHandler) GetLikes(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "InteractionHandler.GetLikes")
	defer span.End()

	// Get post ID from URL path
	vars := mux.Vars(r)
	postID := vars["id"]
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
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

	// Get likes
	resp, err := h.interactionClient.GetPostLikes(ctx, &interactionpb.GetPostLikesRequest{
		PostId: postID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		http.Error(w, "Failed to get likes", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get likes", zap.Error(err), zap.String("post_id", postID))
		return
	}

	// Prepare response
	likes := make([]LikeResponse, 0, len(resp.Likes))
	for _, like := range resp.Likes {
		// Parse created time
		createdAt, err := time.Parse(time.RFC3339, like.CreatedAt)
		if err != nil {
			createdAt = time.Now() // Fallback
		}

		likes = append(likes, LikeResponse{
			UserID:    like.UserId,
			Username:  like.Username,
			CreatedAt: createdAt,
		})
	}

	// Send response
	response := LikesResponse{
		Likes: likes,
		Total: int(resp.Total),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}