package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware adds distributed tracing to HTTP requests
func TracingMiddleware(tracer trace.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start a new span for the request
			ctx, span := tracer.Start(
				r.Context(),
				"HTTP "+r.Method+" "+r.URL.Path,
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.host", r.Host),
					attribute.String("http.user_agent", r.UserAgent()),
				),
			)
			defer span.End()

			// Inject current span into the response headers for frontend tracing
			w.Header().Set("X-Trace-ID", span.SpanContext().TraceID().String())

			// Call the next handler with the context containing the span
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}