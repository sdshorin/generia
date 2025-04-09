package middleware

import (
	"net/http"
	"time"

	"github.com/sdshorin/generia/pkg/logger"
	"go.uber.org/zap"
)

// responseWriter is a custom response writer that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader overrides the original WriteHeader method to capture the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Calculate request processing time
		duration := time.Since(start)

		// Log the request details
		logger.Logger.Info("HTTP Request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Int("status", rw.statusCode),
			zap.Duration("duration", duration),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}