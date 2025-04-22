package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/api-gateway/middleware"
	"go.uber.org/zap"

	characterpb "github.com/sdshorin/generia/api/grpc/character"
)

// CharacterHandler handles character-related HTTP requests
type CharacterHandler struct {
	characterClient characterpb.CharacterServiceClient
	timeout         time.Duration
}

// NewCharacterHandler creates a new CharacterHandler
func NewCharacterHandler(
	characterClient characterpb.CharacterServiceClient,
	timeout time.Duration,
) *CharacterHandler {
	return &CharacterHandler{
		characterClient: characterClient,
		timeout:         timeout,
	}
}

// CreateCharacterRequest represents a request to create a character
type CreateCharacterRequest struct {
	DisplayName   string `json:"display_name"`
	AvatarMediaID string `json:"avatar_media_id,omitempty"`
	Meta          string `json:"meta,omitempty"` // JSON string
}

// CreateCharacterResponse represents a response after creating a character
type CreateCharacterResponse struct {
	ID           string `json:"id"`
	WorldID      string `json:"world_id"`
	RealUserID   string `json:"real_user_id,omitempty"`
	IsAI         bool   `json:"is_ai"`
	DisplayName  string `json:"display_name"`
	AvatarMediaID string `json:"avatar_media_id,omitempty"`
	Meta         string `json:"meta,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// CreateCharacter handles character creation requests
func (h *CharacterHandler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	// Get world ID from URL
	vars := mux.Vars(r)
	worldID := vars["world_id"]
	if worldID == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req CreateCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		logger.Logger.Error("Failed to parse request body", zap.Error(err))
		return
	}

	// Validate request
	if req.DisplayName == "" {
		http.Error(w, "Display name is required", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	// Create character
	realUserID := userID // Use pointer to string to handle nil case in proto
	resp, err := h.characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
		WorldId:       worldID,
		RealUserId:    &realUserID,
		DisplayName:   req.DisplayName,
		AvatarMediaId: stringToOptionalString(req.AvatarMediaID),
		Meta:          stringToOptionalString(req.Meta),
	})
	if err != nil {
		http.Error(w, "Failed to create character", http.StatusInternalServerError)
		logger.Logger.Error("Failed to create character", zap.Error(err))
		return
	}

	// Prepare response
	response := CreateCharacterResponse{
		ID:           resp.Id,
		WorldID:      resp.WorldId,
		IsAI:         resp.IsAi,
		DisplayName:  resp.DisplayName,
		CreatedAt:    resp.CreatedAt,
	}

	// Add optional fields
	if resp.RealUserId != nil {
		response.RealUserID = *resp.RealUserId
	}
	if resp.AvatarMediaId != nil {
		response.AvatarMediaID = *resp.AvatarMediaId
	}
	if resp.Meta != nil {
		response.Meta = *resp.Meta
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// GetCharacterResponse represents a response for a character
type GetCharacterResponse struct {
	ID           string `json:"id"`
	WorldID      string `json:"world_id"`
	RealUserID   string `json:"real_user_id,omitempty"`
	IsAI         bool   `json:"is_ai"`
	DisplayName  string `json:"display_name"`
	AvatarMediaID string `json:"avatar_media_id,omitempty"`
	Meta         string `json:"meta,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// GetCharacter handles character retrieval requests
func (h *CharacterHandler) GetCharacter(w http.ResponseWriter, r *http.Request) {
	// Get character ID from URL
	vars := mux.Vars(r)
	characterID := vars["character_id"]
	if characterID == "" {
		http.Error(w, "Character ID is required", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	// Get character
	resp, err := h.characterClient.GetCharacter(ctx, &characterpb.GetCharacterRequest{
		CharacterId: characterID,
	})
	if err != nil {
		http.Error(w, "Failed to get character", http.StatusInternalServerError)
		logger.Logger.Error("Failed to get character", zap.Error(err))
		return
	}

	// Prepare response
	response := GetCharacterResponse{
		ID:           resp.Id,
		WorldID:      resp.WorldId,
		IsAI:         resp.IsAi,
		DisplayName:  resp.DisplayName,
		CreatedAt:    resp.CreatedAt,
	}

	// Add optional fields
	if resp.RealUserId != nil {
		response.RealUserID = *resp.RealUserId
	}
	if resp.AvatarMediaId != nil {
		response.AvatarMediaID = *resp.AvatarMediaId
	}
	if resp.Meta != nil {
		response.Meta = *resp.Meta
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// CharacterListResponse represents a response with a list of characters
type CharacterListResponse struct {
	Characters []GetCharacterResponse `json:"characters"`
}

// GetUserCharactersInWorld handles requests to get a user's characters in a world
func (h *CharacterHandler) GetUserCharactersInWorld(w http.ResponseWriter, r *http.Request) {
	// Get world ID and user ID from URL
	vars := mux.Vars(r)
	worldID := vars["world_id"]
	userID := vars["user_id"]

	if worldID == "" {
		http.Error(w, "World ID is required", http.StatusBadRequest)
		return
	}

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	// Get user's characters in world
	resp, err := h.characterClient.GetUserCharactersInWorld(ctx, &characterpb.GetUserCharactersInWorldRequest{
		UserId:  userID,
		WorldId: worldID,
	})
	if err != nil {
		http.Error(w, "Failed to get user characters", http.StatusInternalServerError)
		logger.Logger.Error("Failed to get user characters", zap.Error(err))
		return
	}

	// Prepare response
	characters := make([]GetCharacterResponse, 0, len(resp.Characters))
	for _, character := range resp.Characters {
		char := GetCharacterResponse{
			ID:           character.Id,
			WorldID:      character.WorldId,
			IsAI:         character.IsAi,
			DisplayName:  character.DisplayName,
			CreatedAt:    character.CreatedAt,
		}

		// Add optional fields
		if character.RealUserId != nil {
			char.RealUserID = *character.RealUserId
		}
		if character.AvatarMediaId != nil {
			char.AvatarMediaID = *character.AvatarMediaId
		}
		if character.Meta != nil {
			char.Meta = *character.Meta
		}

		characters = append(characters, char)
	}

	response := CharacterListResponse{
		Characters: characters,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// Helper function to convert empty string to nil
func stringToOptionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}