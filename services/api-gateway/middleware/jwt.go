package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"instagram-clone/pkg/logger"
	"go.uber.org/zap"
)

// Key for user ID in context
type contextKey string

const (
	// UserIDKey is the key to store the user ID in the request context
	UserIDKey = contextKey("user_id")
)

// JWTMiddleware handles JWT authentication
type JWTMiddleware struct {
	jwtSecret string
}

// NewJWTMiddleware creates a new JWTMiddleware
func NewJWTMiddleware(jwtSecret string) *JWTMiddleware {
	return &JWTMiddleware{
		jwtSecret: jwtSecret,
	}
}

// RequireAuth enforces authentication
func (m *JWTMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Token format: "Bearer {token}"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := tokenParts[1]
		
		// Validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			logger.Logger.Debug("Invalid token", zap.Error(err))
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userId, ok := claims["user_id"].(string)
			if !ok {
				logger.Logger.Error("User ID not found in token")
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userId)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			logger.Logger.Error("Failed to parse token claims")
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
		}
	})
}

// Optional tries to authenticate but allows unauthenticated requests
func (m *JWTMiddleware) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from header
		authHeader := r.Header.Get("Authorization")
		
		// If no token, continue without authentication
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Try to parse token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, but it's optional, so continue
			next.ServeHTTP(w, r)
			return
		}

		tokenString := tokenParts[1]
		
		// Validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		// If token is valid, add user ID to context
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if userId, ok := claims["user_id"].(string); ok {
					ctx := context.WithValue(r.Context(), UserIDKey, userId)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// If token is invalid, continue without authentication
		next.ServeHTTP(w, r)
	})
}