package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/api-gateway/middleware"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	worldpb "github.com/sdshorin/generia/api/grpc/world"
)

type WorldCreateRequest struct {
	Name            string `json:"name" validate:"required,min=3,max=100"`
	Description     string `json:"description"`
	Prompt          string `json:"prompt" validate:"required,min=10"`
	CharactersCount int32  `json:"characters_count" validate:"min=1,max=40"`
	PostsCount      int32  `json:"posts_count" validate:"min=1,max=250"`
}

// WorldHandler handles world-related requests
type WorldHandler struct {
	worldClient worldpb.WorldServiceClient
	timeout     time.Duration
	tracer      trace.Tracer
	jwtSecret   string
}

// NewWorldHandler creates a new WorldHandler
func NewWorldHandler(worldClient worldpb.WorldServiceClient, timeout time.Duration, jwtSecret string) *WorldHandler {
	return &WorldHandler{
		worldClient: worldClient,
		timeout:     timeout,
		jwtSecret:   jwtSecret,
	}
}

// GetWorlds handles GET /worlds
func (h *WorldHandler) GetWorlds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(middleware.UserIDKey)
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	status := r.URL.Query().Get("status")

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	resp, err := h.worldClient.GetWorlds(timeoutCtx, &worldpb.GetWorldsRequest{
		UserId: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
		Status: status,
	})

	if err != nil {
		httpStatus := grpcStatusToHTTP(err)
		logger.Logger.Error("Failed to get worlds", zap.Error(err), zap.Int("http_status", httpStatus))
		http.Error(w, "Failed to get worlds", httpStatus)
		return
	}

	// print full response
	logger.Logger.Info("GetWorlds response", zap.Any("response", resp))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

// CreateWorld handles POST /worlds
func (h *WorldHandler) CreateWorld(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(middleware.UserIDKey)
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req WorldCreateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	resp, err := h.worldClient.CreateWorld(timeoutCtx, &worldpb.CreateWorldRequest{
		UserId:          userID,
		Name:            req.Name,
		Description:     req.Description,
		Prompt:          req.Prompt,
		CharactersCount: req.CharactersCount,
		PostsCount:      req.PostsCount,
	})

	if err != nil {
		httpStatus := grpcStatusToHTTP(err)
		logger.Logger.Error("Failed to create world", zap.Error(err), zap.Int("http_status", httpStatus))
		http.Error(w, "Failed to create world", httpStatus)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetWorld handles GET /worlds/{world_id}
func (h *WorldHandler) GetWorld(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(middleware.UserIDKey)
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	worldID := vars["world_id"]
	if worldID == "" {
		http.Error(w, "world_id is required", http.StatusBadRequest)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	resp, err := h.worldClient.GetWorld(timeoutCtx, &worldpb.GetWorldRequest{
		WorldId: worldID,
		UserId:  userID,
	})

	if err != nil {
		httpStatus := grpcStatusToHTTP(err)
		logger.Logger.Error("Failed to get world", zap.Error(err), zap.String("world_id", worldID), zap.Int("http_status", httpStatus))
		http.Error(w, "Failed to get world", httpStatus)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// JoinWorld handles POST /worlds/{world_id}/join
func (h *WorldHandler) JoinWorld(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(middleware.UserIDKey)
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	worldID := vars["world_id"]
	if worldID == "" {
		http.Error(w, "world_id is required", http.StatusBadRequest)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	resp, err := h.worldClient.JoinWorld(timeoutCtx, &worldpb.JoinWorldRequest{
		UserId:  userID,
		WorldId: worldID,
	})

	if err != nil {
		httpStatus := grpcStatusToHTTP(err)
		logger.Logger.Error("Failed to join world",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("world_id", worldID),
			zap.Int("http_status", httpStatus))
		http.Error(w, "Failed to join world", httpStatus)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetWorldStatus handles GET /worlds/{world_id}/status
func (h *WorldHandler) GetWorldStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDValue := ctx.Value(middleware.UserIDKey)
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	worldID := vars["world_id"]
	if worldID == "" {
		http.Error(w, "world_id is required", http.StatusBadRequest)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	resp, err := h.worldClient.GetGenerationStatus(timeoutCtx, &worldpb.GetGenerationStatusRequest{
		WorldId: worldID,
	})

	if err != nil {
		httpStatus := grpcStatusToHTTP(err)
		logger.Logger.Error("Failed to get world generation status",
			zap.Error(err),
			zap.String("world_id", worldID),
			zap.Int("http_status", httpStatus))
		http.Error(w, "Failed to get world generation status", httpStatus)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// grpcStatusToHTTP converts gRPC status codes to HTTP status codes
func grpcStatusToHTTP(err error) int {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			return http.StatusNotFound
		case codes.InvalidArgument:
			return http.StatusBadRequest
		case codes.Unauthenticated:
			return http.StatusUnauthorized
		case codes.PermissionDenied:
			return http.StatusForbidden
		case codes.AlreadyExists:
			return http.StatusConflict
		case codes.ResourceExhausted:
			return http.StatusTooManyRequests
		case codes.FailedPrecondition:
			return http.StatusPreconditionFailed
		case codes.Unimplemented:
			return http.StatusNotImplemented
		case codes.Unavailable:
			return http.StatusServiceUnavailable
		case codes.DeadlineExceeded:
			return http.StatusRequestTimeout
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// validateTokenFromQuery validates JWT token from query parameters
func (h *WorldHandler) validateTokenFromQuery(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userID, ok := claims["user_id"].(string); ok {
			return userID, nil
		}
		return "", fmt.Errorf("user_id not found in token")
	}

	return "", fmt.Errorf("invalid token format")
}

// StreamWorldStatus handles SSE for world generation status
func (h *WorldHandler) StreamWorldStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// For SSE, we need to check token from query params since EventSource doesn't support custom headers
	token := r.URL.Query().Get("token")
	var userID string
	
	if token == "" {
		// Fallback to context (if middleware already handled it)
		userIDValue := ctx.Value(middleware.UserIDKey)
		if userIDValue == nil {
			http.Error(w, "Unauthorized: token required", http.StatusUnauthorized)
			return
		}
		var ok bool
		userID, ok = userIDValue.(string)
		if !ok || userID == "" {
			http.Error(w, "Unauthorized: invalid user context", http.StatusUnauthorized)
			return
		}
	} else {
		// Validate token manually for SSE
		var err error
		userID, err = h.validateTokenFromQuery(token)
		if err != nil {
			logger.Logger.Debug("SSE token validation failed", zap.Error(err))
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}
	}

	vars := mux.Vars(r)
	worldID := vars["world_id"]
	if worldID == "" {
		http.Error(w, "world_id is required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Send a ping immediately to establish connection
	fmt.Fprintf(w, "data: {\"type\": \"ping\"}\n\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get generation status
			timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
			resp, err := h.worldClient.GetGenerationStatus(timeoutCtx, &worldpb.GetGenerationStatusRequest{
				WorldId: worldID,
			})
			cancel()

			if err != nil {
				logger.Logger.Error("Failed to get world generation status in SSE",
					zap.Error(err),
					zap.String("world_id", worldID))
				continue
			}

			// Convert response to JSON
			jsonData, err := json.Marshal(resp)
			if err != nil {
				logger.Logger.Error("Failed to marshal generation status",
					zap.Error(err),
					zap.String("world_id", worldID))
				continue
			}

			// Send data
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			// Stop streaming if generation is completed
			if resp.Status == "completed" || resp.Status == "failed" {
				return
			}
		}
	}
}
