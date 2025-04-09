package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sdshorin/generia/pkg/logger"
	"go.uber.org/zap"
)

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status string `json:"status"`
}

// HealthCheckHandler handles health check requests
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Create response
	response := HealthCheckResponse{
		Status: "UP",
	}

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode health check response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ReadinessCheckHandler handles readiness check requests
func ReadinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// This is a simple readiness check, but you may want to check connections to other services
	// like databases, caches, etc.
	
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Create response
	response := HealthCheckResponse{
		Status: "READY",
	}

	// Encode response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode readiness check response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}