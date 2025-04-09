package middleware

import (
	"context"
	"net/http"
	"strings"

	"instagram-clone/pkg/auth"
	"instagram-clone/pkg/logger"
	"go.uber.org/zap"
)

// AuthMiddleware middleware для аутентификации
type AuthMiddleware struct {
	tokenService auth.TokenService
}

// NewAuthMiddleware создает новый AuthMiddleware
func NewAuthMiddleware(tokenService auth.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// AuthUserKey ключ для хранения ID пользователя в контексте запроса
type AuthUserKey string

// ContextUserKey ключ для доступа к ID пользователя в контексте
const ContextUserKey AuthUserKey = "user_id"

// Authenticate middleware для проверки аутентификации
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Если токен отсутствует, продолжаем выполнение с пустым контекстом пользователя
			next.ServeHTTP(w, r)
			return
		}

		// Обычно токен передается в формате "Bearer <token>"
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) != 2 {
			// Если токен некорректен, продолжаем выполнение с пустым контекстом пользователя
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем токен
		claims, err := m.tokenService.ValidateToken(splitToken[1])
		if err != nil {
			logger.Logger.Warn("Invalid token", zap.Error(err))
			// Если токен недействителен, продолжаем выполнение с пустым контекстом пользователя
			next.ServeHTTP(w, r)
			return
		}

		// Если токен действителен, добавляем ID пользователя в контекст запроса
		ctx := context.WithValue(r.Context(), ContextUserKey, claims.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// RequireAuth middleware для проверки обязательной аутентификации
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Обычно токен передается в формате "Bearer <token>"
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) != 2 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Проверяем токен
		claims, err := m.tokenService.ValidateToken(splitToken[1])
		if err != nil {
			logger.Logger.Warn("Invalid token", zap.Error(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Если токен действителен, добавляем ID пользователя в контекст запроса
		ctx := context.WithValue(r.Context(), ContextUserKey, claims.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromContext получает ID пользователя из контекста запроса
func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(ContextUserKey).(string)
	if !ok {
		return ""
	}
	return userID
}