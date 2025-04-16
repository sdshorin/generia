package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/sdshorin/generia/pkg/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracer initializes an OpenTelemetry tracer
func InitTracer(cfg *config.TelemetryConfig) (*sdktrace.TracerProvider, error) {
	if cfg.DisableTracing {
		return sdktrace.NewTracerProvider(), nil
	}

	// Create OTLP exporter
	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(), // Using insecure for simplicity
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Configure trace provider with exporter and resource
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRatio)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Configure propagator
	var propagator propagation.TextMapPropagator
	switch cfg.PropagatorType {
	case "b3":
		// Use B3 propagator
		propagator = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{})
	case "all":
		// Use composite propagator with all standard propagators
		propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	default:
		// Default to W3C TraceContext (recommended)
		propagator = propagation.TraceContext{}
	}

	otel.SetTextMapPropagator(propagator)

	return tp, nil
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}
