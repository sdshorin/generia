package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/api-gateway/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	cdnpb "github.com/sdshorin/generia/api/grpc/cdn"
	mediapb "github.com/sdshorin/generia/api/grpc/media"
)

// MediaHandler handles media-related HTTP requests
type MediaHandler struct {
	mediaClient mediapb.MediaServiceClient
	cdnClient   cdnpb.CDNServiceClient
	tracer      trace.Tracer
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler(
	mediaClient mediapb.MediaServiceClient,
	cdnClient cdnpb.CDNServiceClient,
	tracer trace.Tracer,
) *MediaHandler {
	return &MediaHandler{
		mediaClient: mediaClient,
		cdnClient:   cdnClient,
		tracer:      tracer,
	}
}

// GetUploadURLRequest represents a request to get a presigned upload URL
type GetUploadURLRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	CharacterID string `json:"character_id,omitempty"` // Character ID who owns the media (optional for world-level media)
	WorldID     string `json:"world_id"`               // World ID where the media belongs
	MediaType   int32  `json:"media_type"`             // Media type enum (1=world_header, 2=world_icon, 3=character_avatar, 4=post_image)
}

// GetUploadURLResponse represents a response with a presigned upload URL
type GetUploadURLResponse struct {
	MediaID   string `json:"media_id"`
	UploadURL string `json:"upload_url"`
	ExpiresAt int64  `json:"expires_at"`
}

// ConfirmUploadRequest represents a request to confirm a direct upload
type ConfirmUploadRequest struct {
	MediaID string `json:"media_id"`
}

// UploadMediaResponse represents a response after uploading media
type UploadMediaResponse struct {
	MediaID  string            `json:"media_id"`
	Variants map[string]string `json:"variants"`
}

// GetMediaURLsResponse represents the response for a media URLs request
type GetMediaURLsResponse struct {
	MediaID     string            `json:"media_id"`
	CharacterID string            `json:"character_id"`
	WorldID     string            `json:"world_id,omitempty"`
	Variants    map[string]string `json:"variants"`
}

// GetUploadURL handles requests to get a presigned upload URL
func (h *MediaHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "MediaHandler.GetUploadURL")
	defer span.End()

	// Check user is authenticated (for authorization check only)
	_, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Parse request body
	var req GetUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to decode request body", zap.Error(err))
		return
	}

	// Validate request
	if req.Filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	if req.WorldID == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	if req.MediaType == 0 {
		http.Error(w, "Media type is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// For character-specific media types, character_id is required
	if (req.MediaType == 3 || req.MediaType == 4) && req.CharacterID == "" {
		http.Error(w, "Character ID is required for character-specific media", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	if req.ContentType == "" {
		// Try to determine content type from filename
		req.ContentType = http.DetectContentType([]byte{})
	}

	// Get presigned upload URL
	resp, err := h.mediaClient.GetPresignedUploadURL(ctx, &mediapb.GetPresignedUploadURLRequest{
		WorldId:     req.WorldID,
		CharacterId: req.CharacterID,
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Size:        req.Size,
		MediaType:   mediapb.MediaType(req.MediaType),
	})
	if err != nil {
		http.Error(w, "Failed to generate upload URL", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to generate upload URL", zap.Error(err))
		return
	}

	// Prepare response
	response := GetUploadURLResponse{
		MediaID:   resp.MediaId,
		UploadURL: resp.UploadUrl,
		ExpiresAt: resp.ExpiresAt,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// ConfirmUpload handles requests to confirm a direct upload
func (h *MediaHandler) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "MediaHandler.ConfirmUpload")
	defer span.End()

	// Check user is authenticated (for authorization check only)
	_, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Parse request body
	var req ConfirmUploadRequest
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

	// Confirm upload
	resp, err := h.mediaClient.ConfirmUpload(ctx, &mediapb.ConfirmUploadRequest{
		MediaId: req.MediaID,
	})
	if err != nil {
		http.Error(w, "Failed to confirm upload", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to confirm upload", zap.Error(err))
		return
	}

	// Convert variants to map
	variants := make(map[string]string)
	for _, variant := range resp.Variants {
		variants[variant.Name] = variant.Url
	}

	// Prepare response
	response := UploadMediaResponse{
		MediaID:  req.MediaID,
		Variants: variants,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// GetMediaURLs handles requests to get media URLs
func (h *MediaHandler) GetMediaURLs(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "MediaHandler.GetMediaURLs")
	defer span.End()

	// Get media ID from URL path
	vars := mux.Vars(r)
	mediaID := vars["id"]
	if mediaID == "" {
		http.Error(w, "Media ID is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Get media info
	mediaInfo, err := h.mediaClient.GetMedia(ctx, &mediapb.GetMediaRequest{
		MediaId: mediaID,
	})
	if err != nil {
		http.Error(w, "Failed to get media info", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get media info", zap.Error(err), zap.String("media_id", mediaID))
		return
	}

	// Get signed URLs for each variant
	variants := make(map[string]string)
	for _, variant := range mediaInfo.Variants {
		// Get signed URL
		urlResp, err := h.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
			MediaId:   mediaID,
			Variant:   variant.Name,
			ExpiresIn: 3600, // 1 hour
		})
		if err != nil {
			logger.Logger.Error("Failed to get signed URL",
				zap.Error(err),
				zap.String("media_id", mediaID),
				zap.String("variant", variant.Name))
			// Continue with other variants
			continue
		}
		variants[variant.Name] = urlResp.Url
	}

	// Prepare response
	response := GetMediaURLsResponse{
		MediaID:     mediaInfo.MediaId,
		CharacterID: mediaInfo.CharacterId,
		WorldID:     mediaInfo.WorldId,
		Variants:    variants,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}
