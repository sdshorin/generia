package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/api-gateway/middleware"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

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
}

// NewWorldHandler creates a new WorldHandler
func NewWorldHandler(worldClient worldpb.WorldServiceClient, timeout time.Duration) *WorldHandler {
	return &WorldHandler{
		worldClient: worldClient,
		timeout:     timeout,
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
		logger.Logger.Error("Failed to get worlds", zap.Error(err))
		http.Error(w, "Failed to get worlds", http.StatusInternalServerError)
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
		logger.Logger.Error("Failed to create world", zap.Error(err))
		http.Error(w, "Failed to create world", http.StatusInternalServerError)
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
		logger.Logger.Error("Failed to get world", zap.Error(err), zap.String("world_id", worldID))
		http.Error(w, "Failed to get world", http.StatusInternalServerError)
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
		logger.Logger.Error("Failed to join world",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("world_id", worldID))
		http.Error(w, "Failed to join world", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetWorldStatus handles GET /worlds/{world_id}/status
// func (h *WorldHandler) GetWorldStatus(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	userIDValue := ctx.Value(middleware.UserIDKey)
// 	if userIDValue == nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}
// 	userID, ok := userIDValue.(string)
// 	if !ok || userID == "" {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	worldID := vars["world_id"]
// 	if worldID == "" {
// 		http.Error(w, "world_id is required", http.StatusBadRequest)
// 		return
// 	}

// 	timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
// 	defer cancel()

// 	resp, err := h.worldClient.GetGenerationStatus(timeoutCtx, &worldpb.GetGenerationStatusRequest{
// 		WorldId: worldID,
// 	})

// 	if err != nil {
// 		logger.Logger.Error("Failed to get world generation status",
// 			zap.Error(err),
// 			zap.String("world_id", worldID))
// 		http.Error(w, "Failed to get world generation status", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(resp)
// }
