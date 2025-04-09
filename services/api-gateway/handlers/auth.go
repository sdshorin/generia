package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"instagram-clone/pkg/logger"
	"instagram-clone/services/api-gateway/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "instagram-clone/api/grpc/auth"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authClient authpb.AuthServiceClient
	tracer     trace.Tracer
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authClient authpb.AuthServiceClient, tracer trace.Tracer) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
		tracer:     tracer,
	}
}

// RegisterRequest represents a request to register a new user
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token     string    `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// UserResponse represents a user in the API response
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "AuthHandler.Register")
	defer span.End()

	// Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to decode request body", zap.Error(err))
		return
	}

	// Validate request
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Register user
	resp, err := h.authClient.Register(ctx, &authpb.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok && statusErr.Code() == codes.AlreadyExists {
			http.Error(w, "Username or email already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to register user", zap.Error(err))
		return
	}

	// Get created time
	var createdAt time.Time
	userInfo, err := h.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: resp.UserId,
	})
	if err == nil && userInfo.CreatedAt != "" {
		createdAt, _ = time.Parse(time.RFC3339, userInfo.CreatedAt)
	} else {
		createdAt = time.Now() // Fallback
	}

	// Prepare response
	expiresAt := time.Unix(resp.ExpiresAt, 0)
	response := AuthResponse{
		Token:     resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt: expiresAt,
		User: UserResponse{
			ID:        resp.UserId,
			Username:  req.Username,
			Email:     req.Email,
			CreatedAt: createdAt,
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username"`
	Password        string `json:"password"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "AuthHandler.Login")
	defer span.End()

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to decode request body", zap.Error(err))
		return
	}

	// Validate request
	if req.EmailOrUsername == "" || req.Password == "" {
		http.Error(w, "Email or username and password are required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Login user
	resp, err := h.authClient.Login(ctx, &authpb.LoginRequest{
		EmailOrUsername: req.EmailOrUsername,
		Password:        req.Password,
	})
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok && statusErr.Code() == codes.NotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else if ok && statusErr.Code() == codes.Unauthenticated {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "Failed to login", http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to login user", zap.Error(err))
		return
	}

	// Get user info
	userInfo, err := h.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: resp.UserId,
	})
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get user info", zap.Error(err))
		return
	}

	// Parse created time
	createdAt, err := time.Parse(time.RFC3339, userInfo.CreatedAt)
	if err != nil {
		createdAt = time.Now() // Fallback
	}

	// Prepare response
	expiresAt := time.Unix(resp.ExpiresAt, 0)
	response := AuthResponse{
		Token:     resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt: expiresAt,
		User: UserResponse{
			ID:        resp.UserId,
			Username:  userInfo.Username,
			Email:     userInfo.Email,
			CreatedAt: createdAt,
		},
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// Me handles retrieving the current user's information
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "AuthHandler.Me")
	defer span.End()

	// Get user ID from context
	userID, ok := ctx.Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Get user info
	userInfo, err := h.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to get user info", zap.Error(err))
		return
	}

	// Parse created time
	createdAt, err := time.Parse(time.RFC3339, userInfo.CreatedAt)
	if err != nil {
		createdAt = time.Now() // Fallback
	}

	// Prepare response
	response := UserResponse{
		ID:        userInfo.UserId,
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		CreatedAt: createdAt,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}

// RefreshTokenRequest represents a request to refresh an access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken handles refreshing access tokens
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "AuthHandler.RefreshToken")
	defer span.End()

	// Parse request body
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to decode request body", zap.Error(err))
		return
	}

	// Validate request
	if req.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		span.SetAttributes(attribute.Bool("error", true))
		return
	}

	// Refresh token
	resp, err := h.authClient.RefreshToken(ctx, &authpb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok && statusErr.Code() == codes.NotFound {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		} else if ok && statusErr.Code() == codes.Unauthenticated {
			http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		} else {
			http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
		}
		span.SetAttributes(attribute.Bool("error", true))
		logger.Logger.Error("Failed to refresh token", zap.Error(err))
		return
	}

	// Prepare response
	expiresAt := time.Unix(resp.ExpiresAt, 0)
	response := struct {
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		ExpiresAt    time.Time `json:"expires_at"`
	}{
		Token:        resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    expiresAt,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
	}
}