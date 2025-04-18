package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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

		// Capture request parameters
		params := map[string]interface{}{}
		
		// Add URL query parameters
		for key, values := range r.URL.Query() {
			if len(values) == 1 {
				params[key] = values[0]
			} else {
				params[key] = values
			}
		}

		// Add request body for POST, PUT, PATCH methods
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if r.Body != nil && r.Header.Get("Content-Type") == "application/json" {
				// Read the body
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					// Close the original body as it's been read
					r.Body.Close()
					
					// Create a new ReadCloser to replace the original body
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					
					// Try to parse JSON
					var bodyJSON map[string]interface{}
					if err := json.Unmarshal(bodyBytes, &bodyJSON); err == nil {
						for k, v := range bodyJSON {
							// Skip sensitive fields
							if k != "password" && k != "password_hash" && k != "token" && k != "refresh_token" {
								params[k] = v
							}
						}
					}
				}
			}
		}

		// Add path parameters (from gorilla/mux)
		if routeVars := mux.Vars(r); len(routeVars) > 0 {
			for k, v := range routeVars {
				params["path_"+k] = v
			}
		}
		
		// Call the next handler
		next.ServeHTTP(rw, r)

		// Calculate request processing time
		duration := time.Since(start)

		// Log the request details
		logFields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Int("status", rw.statusCode),
			zap.Duration("duration", duration),
			zap.String("user_agent", r.UserAgent()),
		}
		
		// Add parameters if available
		if len(params) > 0 {
			paramsJSON, _ := json.Marshal(params)
			logFields = append(logFields, zap.String("params", string(paramsJSON)))
		}

		logger.Logger.Info("HTTP Request", logFields...)
	})
}