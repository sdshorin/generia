package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"instagram-clone/pkg/logger"
	"instagram-clone/services/api-gateway/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	cdnpb "instagram-clone/api/grpc/cdn"
	mediapb "instagram-clone/api/grpc/media"
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

// UploadMediaRequest represents a request to upload media
type UploadMediaRequest struct {
	MediaData   string `json:"media_data"` // Base64-encoded media
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
}

// UploadMediaResponse represents a response after uploading media
type UploadMediaResponse struct {
	MediaID  string            `json:"media_id"`
	Variants map[string]string `json:"variants"`
}

// UploadMedia handles media upload requests
func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "MediaHandler.UploadMedia")
	defer span.End()

	// Get user ID from context
	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Parse request body
	var req UploadMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to decode request body", zap.Error(err))
		return
	}

	// Validate request
	if req.MediaData == "" {
		http.Error(w, "Media data is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Decode base64 data
	mediaData := req.MediaData
	if strings.Contains(mediaData, ";base64,") {
		// Handle data URLs (e.g., "data:image/jpeg;base64,...")
		parts := strings.SplitN(mediaData, ";base64,", 2)
		if len(parts) != 2 {
			http.Error(w, "Invalid media data format", http.StatusBadRequest)
			span.SetAttributes(attribute.Bool("error", true))
			return
		}
		mediaData = parts[1]
	}

	decodedData, err := base64.StdEncoding.DecodeString(mediaData)
	if err != nil {
		http.Error(w, "Invalid base64 encoding", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to decode base64 data", zap.Error(err))
		return
	}

	// Determine content type if not provided
	contentType := req.ContentType
	if contentType == "" {
		contentType = http.DetectContentType(decodedData)
	}

	// Ensure we have a filename
	filename := req.Filename
	if filename == "" {
		// Generate a default filename based on current time
		extension := ""
		switch {
		case strings.HasPrefix(contentType, "image/jpeg"):
			extension = "jpg"
		case strings.HasPrefix(contentType, "image/png"):
			extension = "png"
		case strings.HasPrefix(contentType, "image/gif"):
			extension = "gif"
		default:
			extension = "bin"
		}
		filename = "upload_" + time.Now().Format("20060102_150405") + "." + extension
	}

	// Initialize upload stream
	stream, err := h.mediaClient.UploadMedia(ctx)
	if err != nil {
		http.Error(w, "Failed to initialize upload stream", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to initialize upload stream", zap.Error(err))
		return
	}

	// Send metadata
	err = stream.Send(&mediapb.UploadMediaRequest{
		Data: &mediapb.UploadMediaRequest_Metadata{
			Metadata: &mediapb.MediaMetadata{
				UserId:      userID,
				Filename:    filename,
				ContentType: contentType,
				Size:        int64(len(decodedData)),
			},
		},
	})
	if err != nil {
		http.Error(w, "Failed to send metadata", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to send metadata", zap.Error(err))
		return
	}

	// Send file data in chunks
	chunkSize := 1024 * 1024 // 1MB chunks
	buffer := bytes.NewBuffer(decodedData)

	for {
		chunk := make([]byte, chunkSize)
		n, err := buffer.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Failed to read file chunk", http.StatusInternalServerError)
			span.SetAttributes(attribute.Bool("error", true))
			logger.Logger.Error("Failed to read file chunk", zap.Error(err))
			return
		}

		// Send chunk
		err = stream.Send(&mediapb.UploadMediaRequest{
			Data: &mediapb.UploadMediaRequest_Chunk{
				Chunk: chunk[:n],
			},
		})
		if err != nil {
			http.Error(w, "Failed to send file chunk", http.StatusInternalServerError)
			span.SetAttributes(attribute.Bool("error", true))
			logger.Logger.Error("Failed to send file chunk", zap.Error(err))
			return
		}
	}

	// Finalize upload
	resp, err := stream.CloseAndRecv()
	if err != nil {
		http.Error(w, "Failed to finalize upload", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to finalize upload", zap.Error(err))
		return
	}

	// Convert variants to map
	variants := make(map[string]string)
	for _, variant := range resp.Variants {
		variants[variant.Name] = variant.Url
	}

	// Prepare response
	response := UploadMediaResponse{
		MediaID:  resp.MediaId,
		Variants: variants,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// GetMediaURLsResponse represents the response for a media URLs request
type GetMediaURLsResponse struct {
	MediaID  string            `json:"media_id"`
	UserID   string            `json:"user_id"`
	Variants map[string]string `json:"variants"`
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
		MediaID:  mediaInfo.MediaId,
		UserID:   mediaInfo.UserId,
		Variants: variants,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}