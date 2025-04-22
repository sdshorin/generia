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

	feedpb "github.com/sdshorin/generia/api/grpc/feed"
)

// FeedHandler handles feed-related HTTP requests
type FeedHandler struct {
	feedClient feedpb.FeedServiceClient
	tracer     trace.Tracer
}

// NewFeedHandler creates a new FeedHandler
func NewFeedHandler(
	feedClient feedpb.FeedServiceClient,
	tracer trace.Tracer,
) *FeedHandler {
	return &FeedHandler{
		feedClient: feedClient,
		tracer:     tracer,
	}
}

// FeedItemResponse represents a post in the feed response
type FeedItemResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	DisplayName   string    `json:"display_name"`
	Caption       string    `json:"caption"`
	MediaURL      string    `json:"media_url"`
	CreatedAt     time.Time `json:"created_at"`
	LikesCount    int       `json:"likes_count"`
	CommentsCount int       `json:"comments_count"`
	UserLiked     bool      `json:"user_liked,omitempty"`
}

// GlobalFeedResponse represents the response for a global feed request
type GlobalFeedResponse struct {
	Posts      []FeedItemResponse `json:"posts"`
	NextCursor string             `json:"next_cursor,omitempty"`
	HasMore    bool               `json:"has_more"`
}

// GetGlobalFeed handles requests to get the global feed
func (h *FeedHandler) GetGlobalFeed(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "FeedHandler.GetGlobalFeed")
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

	// Получаем world_id из URL параметров
	vars := mux.Vars(r)
	worldID := vars["world_id"]
	if worldID == "" {
		http.Error(w, "world_id is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Missing world_id parameter")
		return
	}

	// Get feed
	resp, err := h.feedClient.GetGlobalFeed(ctx, &feedpb.GetGlobalFeedRequest{
		UserId:  userID,
		Limit:   int32(limit),
		Cursor:  cursor,
		WorldId: worldID,
	})
	if err != nil {
		http.Error(w, "Failed to get feed", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get feed", zap.Error(err))
		return
	}

	// Prepare response
	posts := make([]FeedItemResponse, 0, len(resp.Posts))
	for _, post := range resp.Posts {
		// Parse created time
		createdAt := time.Unix(post.CreatedAt, 0)

		posts = append(posts, FeedItemResponse{
			ID:            post.Id,
			UserID:        post.UserId,
			DisplayName:   post.User.DisplayName,
			Caption:       post.Caption,
			MediaURL:      post.MediaUrl,
			CreatedAt:     createdAt,
			LikesCount:    int(post.Stats.LikesCount),
			CommentsCount: int(post.Stats.CommentsCount),
			UserLiked:     post.Stats.UserLiked,
		})
	}

	// Send response
	response := GlobalFeedResponse{
		Posts:      posts,
		NextCursor: resp.NextCursor,
		HasMore:    resp.HasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}
