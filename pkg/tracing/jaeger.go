package tracing

import (
	"context"
	"io"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// NewJaegerTracer initializes a new Jaeger tracer
func NewJaegerTracer(serviceName, jaegerHost string) (tracesdk.TracerProvider, io.Closer, error) {
	// Create Jaeger exporter with explicit port configuration
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(
		jaeger.WithAgentHost(jaegerHost),
		jaeger.WithAgentPort("6831"), // Explicit port specification to avoid UDP connection issues
	))
	if err != nil {
		return nil, nil, err
	}

	tp := tracesdk.NewTracerProvider(
		// Configure batching of spans with reasonable timeout to prevent connection issues
		tracesdk.WithBatcher(exp,
			tracesdk.WithMaxExportBatchSize(10),
			tracesdk.WithBatchTimeout(5*time.Second),
			tracesdk.WithMaxQueueSize(10),
		),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
		// Add sampling configuration to reduce trace volume in production
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(0.5))),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Return provider and a closer function
	closer := &tracerCloser{tp: tp}
	return tp, closer, nil
}

// tracerCloser implements io.Closer for the tracer provider
type tracerCloser struct {
	tp *tracesdk.TracerProvider
}

func (c *tracerCloser) Close() error {
	// Use a timeout when shutting down
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.tp.Shutdown(ctx)
}