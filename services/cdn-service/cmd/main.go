package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	// "strconv"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sdshorin/generia/pkg/config"
	"github.com/sdshorin/generia/pkg/discovery"
	"github.com/sdshorin/generia/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	cdnpb "github.com/sdshorin/generia/api/grpc/cdn"
)

// CDNService implements the CDN service
type CDNService struct {
	cdnpb.UnimplementedCDNServiceServer
	logger     *zap.Logger
	domain     string
	defaultTTL int
	signingKey string
}

// GetSignedURL implements the GetSignedURL method
func (s *CDNService) GetSignedURL(ctx context.Context, req *cdnpb.GetSignedURLRequest) (*cdnpb.GetSignedURLResponse, error) {
	// Basic implementation
	ttl := s.defaultTTL
	if req.ExpiresIn > 0 {
		ttl = int(req.ExpiresIn)
	}

	expiry := time.Now().Add(time.Duration(ttl) * time.Second).Unix()
	path := req.Path

	// Generate signature
	signature := s.generateSignature(path, expiry)

	// Build URL
	url := fmt.Sprintf("https://%s/%s?expires=%d&signature=%s", s.domain, path, expiry, signature)

	return &cdnpb.GetSignedURLResponse{
		Url:       url,
		ExpiresAt: expiry,
	}, nil
}

// InvalidateCache implements the InvalidateCache method
func (s *CDNService) InvalidateCache(ctx context.Context, req *cdnpb.InvalidateCacheRequest) (*cdnpb.InvalidateCacheResponse, error) {
	// Placeholder implementation
	s.logger.Info("InvalidateCache called", zap.Strings("paths", req.Paths))
	
	// Generate unique operation ID
	operationID := fmt.Sprintf("op-%d", time.Now().UnixNano())
	
	return &cdnpb.InvalidateCacheResponse{
		Success:     true,
		OperationId: operationID,
	}, nil
}

// GetCDNConfig implements the GetCDNConfig method
func (s *CDNService) GetCDNConfig(ctx context.Context, req *cdnpb.GetCDNConfigRequest) (*cdnpb.GetCDNConfigResponse, error) {
	// Placeholder implementation
	return &cdnpb.GetCDNConfigResponse{
		CdnDomain:          s.domain,
		DefaultTtl:         int32(s.defaultTTL),
		AllowedOrigins:     []string{"*"},
		AllowedHttpMethods: []string{"GET", "HEAD", "OPTIONS"},
	}, nil
}

// HealthCheck implements the HealthCheck method
func (s *CDNService) HealthCheck(ctx context.Context, req *cdnpb.HealthCheckRequest) (*cdnpb.HealthCheckResponse, error) {
	// Basic implementation
	return &cdnpb.HealthCheckResponse{
		Status: cdnpb.HealthCheckResponse_SERVING,
	}, nil
}

// Helper function to generate signature
func (s *CDNService) generateSignature(path string, expiry int64) string {
	// Create HMAC
	h := hmac.New(sha256.New, []byte(s.signingKey))
	
	// Create string to sign
	stringToSign := fmt.Sprintf("%s/%d", path, expiry)
	
	// Write to HMAC
	h.Write([]byte(stringToSign))
	
	// Get signature
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	// Initialize logger
	if err := logger.InitProduction(); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize OpenTelemetry
	tp, err := initTracer(cfg.Jaeger.Host)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Logger.Error("Error shutting down tracer provider", zap.Error(err))
		}
	}()

	// Initialize service discovery client
	discoveryClient, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		logger.Logger.Fatal("Failed to create service discovery client", zap.Error(err))
	}

	// Register service with Consul
	serviceID := fmt.Sprintf("%s-%s", cfg.Service.Name, cfg.Service.Host)
	err = discoveryClient.Register(serviceID, cfg.Service.Name, cfg.Service.Host, cfg.Service.Port, []string{"cdn", "api"})
	if err != nil {
		logger.Logger.Fatal("Failed to register service", zap.Error(err))
	}
	defer discoveryClient.Deregister(serviceID)

	// Initialize CDN service
	cdnService := &CDNService{
		logger:     logger.Logger,
		domain:     cfg.CDN.Domain,
		defaultTTL: cfg.CDN.DefaultTTL,
		signingKey: cfg.CDN.SigningKey,
	}

	// Create gRPC server with middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger.Logger),
			otelgrpc.UnaryServerInterceptor(),
		)),
	)

	// Register services
	cdnpb.RegisterCDNServiceServer(grpcServer, cdnService)
	
	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("cdn.CDNService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Initialize metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := cfg.Service.Port + 10000 // Metrics on port+10000
		logger.Logger.Info("Starting metrics server", zap.Int("port", metricsPort))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			logger.Logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Service.Host, cfg.Service.Port))
	if err != nil {
		logger.Logger.Fatal("Failed to listen", zap.Error(err))
	}

	logger.Logger.Info("Starting CDN service", 
		zap.String("host", cfg.Service.Host), 
		zap.Int("port", cfg.Service.Port))
	
	// Handle graceful shutdown
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("Shutting down CDN service...")
	grpcServer.GracefulStop()
	logger.Logger.Info("CDN service stopped")
}

func initTracer(jaegerHost string) (*tracesdk.TracerProvider, error) {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost(jaegerHost)))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("cdn-service"),
		)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
}