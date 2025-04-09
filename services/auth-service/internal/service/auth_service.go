package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/auth-service/internal/models"
	"github.com/sdshorin/generia/services/auth-service/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/sdshorin/generia/api/grpc/auth"
)

// AuthService implements the auth gRPC service
type AuthService struct {
	authpb.UnimplementedAuthServiceServer
	userRepo       repository.UserRepository
	jwtSecret      string
	jwtExpiration  time.Duration
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExpiration time.Duration) authpb.AuthServiceServer {
	return &AuthService{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

// Register handles user registration
func (s *AuthService) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username, email, and password are required")
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		logger.Logger.Error("Failed to check if email exists", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to check if email exists")
	}
	if existingUser != nil {
		return nil, status.Errorf(codes.AlreadyExists, "email already exists")
	}

	// Check if username already exists
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		logger.Logger.Error("Failed to check if username exists", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to check if username exists")
	}
	if existingUser != nil {
		return nil, status.Errorf(codes.AlreadyExists, "username already exists")
	}

	// Hash password
	passwordHash, err := models.HashPassword(req.Password)
	if err != nil {
		logger.Logger.Error("Failed to hash password", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		logger.Logger.Error("Failed to create user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	// Generate access token
	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		logger.Logger.Error("Failed to generate access token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate access token")
	}

	// Generate refresh token
	refreshToken, refreshTokenHash, err := s.generateRefreshToken()
	if err != nil {
		logger.Logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	// Save refresh token - retry mechanism for potential DB contention
	var saveErr error
	for attempts := 0; attempts < 3; attempts++ {
		saveErr = s.userRepo.SaveRefreshToken(ctx, &models.RefreshToken{
			UserID:    user.ID,
			TokenHash: refreshTokenHash,
			ExpiresAt: time.Now().Add(s.jwtExpiration * 30), // Refresh token lasts 30 times longer than access token
		})
		if saveErr == nil {
			break // Success, exit the retry loop
		}
		
		logger.Logger.Warn("Attempt to save refresh token failed, retrying...", 
			zap.Error(saveErr), 
			zap.Int("attempt", attempts+1))
		
		// Generate a new refresh token to avoid unique constraint violation on retry
		if attempts < 2 { // Only regenerate if we're going to retry
			refreshToken, refreshTokenHash, err = s.generateRefreshToken()
			if err != nil {
				logger.Logger.Error("Failed to regenerate refresh token during retry", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
			}
		}
		
		// Short wait before retry to allow potential DB contention to resolve
		time.Sleep(time.Millisecond * 200 * time.Duration(attempts+1))
	}
	
	if saveErr != nil {
		logger.Logger.Error("All attempts to save refresh token failed", zap.Error(saveErr))
		return nil, status.Errorf(codes.Internal, "failed to save refresh token after multiple attempts")
	}

	return &authpb.RegisterResponse{
		UserId:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.jwtExpiration).Unix(),
	}, nil
}

// Login handles user login
func (s *AuthService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	// Validate input
	if req.EmailOrUsername == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email or username and password are required")
	}

	// Try to find user by email or username
	var user *models.User
	var err error

	// Check if it looks like an email
	if isEmail(req.EmailOrUsername) {
		user, err = s.userRepo.GetByEmail(ctx, req.EmailOrUsername)
	} else {
		user, err = s.userRepo.GetByUsername(ctx, req.EmailOrUsername)
	}

	if err != nil {
		logger.Logger.Error("Failed to find user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to find user")
	}

	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Check password
	if !models.CheckPassword(req.Password, user.PasswordHash) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Generate access token
	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		logger.Logger.Error("Failed to generate access token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate access token")
	}

	// Generate refresh token
	refreshToken, refreshTokenHash, err := s.generateRefreshToken()
	if err != nil {
		logger.Logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	// Save refresh token - retry mechanism for potential DB contention
	var saveErr error
	for attempts := 0; attempts < 3; attempts++ {
		saveErr = s.userRepo.SaveRefreshToken(ctx, &models.RefreshToken{
			UserID:    user.ID,
			TokenHash: refreshTokenHash,
			ExpiresAt: time.Now().Add(s.jwtExpiration * 30), // Refresh token lasts 30 times longer than access token
		})
		if saveErr == nil {
			break // Success, exit the retry loop
		}
		
		logger.Logger.Warn("Attempt to save refresh token failed, retrying...", 
			zap.Error(saveErr), 
			zap.Int("attempt", attempts+1))
		
		// Generate a new refresh token to avoid unique constraint violation on retry
		if attempts < 2 { // Only regenerate if we're going to retry
			refreshToken, refreshTokenHash, err = s.generateRefreshToken()
			if err != nil {
				logger.Logger.Error("Failed to regenerate refresh token during retry", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
			}
		}
		
		// Short wait before retry to allow potential DB contention to resolve
		time.Sleep(time.Millisecond * 200 * time.Duration(attempts+1))
	}
	
	if saveErr != nil {
		logger.Logger.Error("All attempts to save refresh token failed", zap.Error(saveErr))
		return nil, status.Errorf(codes.Internal, "failed to save refresh token after multiple attempts")
	}

	return &authpb.LoginResponse{
		UserId:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.jwtExpiration).Unix(),
	}, nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	// Validate input
	if req.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "token is required")
	}

	// Parse token
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	// Check if token is valid
	if !token.Valid {
		return &authpb.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &authpb.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	// Get user ID
	userID, ok := claims["user_id"].(string)
	if !ok {
		return &authpb.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Logger.Error("Failed to get user", zap.Error(err))
		return &authpb.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	if user == nil {
		return &authpb.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}

// GetUserInfo gets user information by ID
func (s *AuthService) GetUserInfo(ctx context.Context, req *authpb.GetUserInfoRequest) (*authpb.UserInfo, error) {
	// Validate input
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, req.UserId)
	if err != nil {
		logger.Logger.Error("Failed to get user", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get user")
	}

	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return &authpb.UserInfo{
		UserId:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	// Validate input
	if req.RefreshToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh_token is required")
	}

	// Hash refresh token
	refreshTokenHash := hashToken(req.RefreshToken)

	// Get refresh token from database
	token, err := s.userRepo.GetRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		logger.Logger.Error("Failed to get refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get refresh token")
	}

	if token == nil {
		return nil, status.Errorf(codes.NotFound, "refresh token not found")
	}

	// Check if refresh token is expired
	if token.ExpiresAt.Before(time.Now()) {
		// Delete expired token
		_ = s.userRepo.DeleteRefreshToken(ctx, refreshTokenHash)
		return nil, status.Errorf(codes.Unauthenticated, "refresh token expired")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(token.UserID)
	if err != nil {
		logger.Logger.Error("Failed to generate access token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate access token")
	}

	// Generate new refresh token
	newRefreshToken, newRefreshTokenHash, err := s.generateRefreshToken()
	if err != nil {
		logger.Logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	// Delete old refresh token
	err = s.userRepo.DeleteRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		logger.Logger.Error("Failed to delete old refresh token", zap.Error(err))
		// Continue even if deletion fails
	}

	// Save new refresh token - retry mechanism for potential DB contention
	var saveErr error
	for attempts := 0; attempts < 3; attempts++ {
		saveErr = s.userRepo.SaveRefreshToken(ctx, &models.RefreshToken{
			UserID:    token.UserID,
			TokenHash: newRefreshTokenHash,
			ExpiresAt: time.Now().Add(s.jwtExpiration * 30), // Refresh token lasts 30 times longer than access token
		})
		if saveErr == nil {
			break // Success, exit the retry loop
		}
		
		logger.Logger.Warn("Attempt to save new refresh token failed, retrying...", 
			zap.Error(saveErr), 
			zap.Int("attempt", attempts+1))
		
		// Generate a new refresh token to avoid unique constraint violation on retry
		if attempts < 2 { // Only regenerate if we're going to retry
			newRefreshToken, newRefreshTokenHash, err = s.generateRefreshToken()
			if err != nil {
				logger.Logger.Error("Failed to regenerate refresh token during retry", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
			}
		}
		
		// Short wait before retry to allow potential DB contention to resolve
		time.Sleep(time.Millisecond * 200 * time.Duration(attempts+1))
	}
	
	if saveErr != nil {
		logger.Logger.Error("All attempts to save refresh token failed", zap.Error(saveErr))
		return nil, status.Errorf(codes.Internal, "failed to save refresh token after multiple attempts")
	}

	return &authpb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(s.jwtExpiration).Unix(),
	}, nil
}

// HealthCheck implements health check
func (s *AuthService) HealthCheck(ctx context.Context, req *authpb.HealthCheckRequest) (*authpb.HealthCheckResponse, error) {
	return &authpb.HealthCheckResponse{
		Status: authpb.HealthCheckResponse_SERVING,
	}, nil
}

// Helper functions

// generateAccessToken generates a JWT access token
func (s *AuthService) generateAccessToken(userID string) (string, error) {
	// Create claims
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(s.jwtExpiration).Unix(),
		"iat":     now.Unix(),
		"iss":     "auth-service",
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// generateRefreshToken generates a refresh token
func (s *AuthService) generateRefreshToken() (string, string, error) {
	// Generate a random token
	token := uuid.New().String()

	// Hash token for storage
	tokenHash := hashToken(token)

	return token, tokenHash, nil
}

// hashToken hashes a token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// isEmail checks if a string looks like an email
func isEmail(s string) bool {
	// Very basic check, you might want to use a more sophisticated validation
	return contains(s, "@")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}